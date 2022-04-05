/*
Copyright 2021 VMware, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

type WorkloadListOptions struct {
	Namespace     string
	AllNamespaces bool
	App           string
	Output        string
}

var (
	_ validation.Validatable = (*WorkloadListOptions)(nil)
	_ cli.Executable         = (*WorkloadListOptions)(nil)
)

func (opts *WorkloadListOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Namespace == "" && !opts.AllNamespaces {
		errs = errs.Also(validation.ErrMissingOneOf(flags.NamespaceFlagName, flags.AllNamespacesFlagName))
	}
	if opts.Namespace != "" && opts.AllNamespaces {
		errs = errs.Also(validation.ErrMultipleOneOf(flags.NamespaceFlagName, flags.AllNamespacesFlagName))
	}

	if opts.App != "" {
		errs = errs.Also(validation.K8sName(opts.App, flags.AppFlagName))
	}

	if opts.Output != "" {
		errs = errs.Also(validation.Enum(opts.Output, flags.OutputFlagName, []string{printer.OutputFormatJson, printer.OutputFormatYaml}))
	}

	return errs
}

func (opts *WorkloadListOptions) Exec(ctx context.Context, c *cli.Config) error {
	workloads := &cartov1alpha1.WorkloadList{}
	labels := map[string]string{}
	if opts.App != "" {
		labels[apis.AppPartOfLabelName] = opts.App
	}
	err := c.List(ctx, workloads, client.InNamespace(opts.Namespace), client.MatchingLabels(labels))
	if err != nil {
		return err
	}

	if opts.Output != "" {
		var list []printer.Object
		for _, w := range workloads.Items {
			list = append(list, &w)
		}
		export, err := printer.OutputResources(list, printer.OutputFormat(opts.Output), c.Scheme)
		if err != nil {
			c.Eprintf("%s %s\n", printer.Serrorf("Failed to output workload:"), err)
			return cli.SilenceError(err)
		}

		c.Printf("%s\n", export)
		return nil
	}

	if len(workloads.Items) == 0 {
		c.Infof("No workloads found.\n")
		return nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h table.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	workloads = workloads.DeepCopy()
	printer.SortByNamespaceAndName(workloads.Items)

	return tablePrinter.PrintObj(workloads, c.Stdout)
}

func NewWorkloadListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Table listing of workloads",
		Long: strings.TrimSpace(`
List workloads in a namespace or across all namespaces.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload list", c.Name),
			fmt.Sprintf("%s workload list %s", c.Name, flags.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateE(ctx, opts),
		RunE:    cli.ExecE(ctx, c, opts),
	}

	cli.AllNamespacesFlag(ctx, cmd, c, &opts.Namespace, &opts.AllNamespaces)
	cmd.Flags().StringVar(&opts.App, cli.StripDash(flags.AppFlagName), "", "application `name` the workload is a part of")
	cmd.Flags().StringVarP(&opts.Output, cli.StripDash(flags.OutputFlagName), "o", "", "output the Workloads formatted. Supported formats: \"json\", \"yaml\"")

	return cmd
}

func (opts *WorkloadListOptions) printList(workloads *cartov1alpha1.WorkloadList, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(workloads.Items))
	for i := range workloads.Items {
		r, err := opts.print(&workloads.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *WorkloadListOptions) print(workload *cartov1alpha1.Workload, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: workload},
	}

	labels := workload.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	row.Cells = append(row.Cells, workload.Name)
	if opts.App == "" {
		row.Cells = append(row.Cells, printer.EmptyString(labels[apis.AppPartOfLabelName]))
	}
	row.Cells = append(row.Cells,
		printer.ConditionStatus(printer.FindCondition(workload.Status.Conditions, cartov1alpha1.WorkloadConditionReady)),
		printer.TimestampSince(workload.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *WorkloadListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	cols := []metav1beta1.TableColumnDefinition{}

	cols = append(cols, metav1beta1.TableColumnDefinition{Name: "Name", Type: "string"})
	if opts.App == "" {
		cols = append(cols, metav1beta1.TableColumnDefinition{Name: "App", Type: "string"})
	}
	cols = append(cols,
		metav1beta1.TableColumnDefinition{Name: "Ready", Type: "string"},
		metav1beta1.TableColumnDefinition{Name: "Age", Type: "string"},
	)

	return cols
}
