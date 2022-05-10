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

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type WorkloadGetOptions struct {
	Namespace string
	Name      string

	Export bool
	Output string
}

var (
	_ validation.Validatable = (*WorkloadGetOptions)(nil)
	_ cli.Executable         = (*WorkloadGetOptions)(nil)
)

func (opts *WorkloadGetOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(validation.ErrMissingField(flags.NamespaceFlagName))
	}

	if opts.Name == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NameArgumentName))
	}

	if opts.Output != "" {
		errs = errs.Also(validation.Enum(opts.Output, flags.OutputFlagName, []string{printer.OutputFormatJson, printer.OutputFormatYaml, printer.OutputFormatYml}))
	}

	return errs
}

func (opts *WorkloadGetOptions) Exec(ctx context.Context, c *cli.Config) error {
	workload := &cartov1alpha1.Workload{}
	err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: opts.Name}, workload)
	if err != nil {
		if apierrs.IsNotFound(err) {
			nsGet := &corev1.Namespace{}
			if getErr := c.Get(ctx, types.NamespacedName{Name: opts.Namespace}, nsGet); getErr != nil && apierrs.IsNotFound(getErr) {
				c.Eprintf("%s %s\n", printer.Serrorf("Error:"), fmt.Sprintf("namespace %q not found, it may not exist or user does not have permissions to read it.", opts.Namespace))
				return cli.SilenceError(getErr)
			}
			c.Errorf("Workload %q not found\n", fmt.Sprintf("%s/%s", opts.Namespace, opts.Name))
			return cli.SilenceError(err)
		}

		return err
	}

	if opts.Export {
		var format printer.OutputFormat
		if opts.Output == "" {
			format = printer.OutputFormat(printer.OutputFormatYaml)
		} else {
			format = printer.OutputFormat(opts.Output)
		}

		export, err := printer.ExportResource(workload, format, c.Scheme)
		if err != nil {
			c.Eprintf("%s %s\n", printer.Serrorf("Failed to export workload:"), err)
			return cli.SilenceError(err)
		}
		c.Printf("%s\n", export)
		return nil
	}

	if opts.Output != "" {
		export, err := printer.OutputResource(workload, printer.OutputFormat(opts.Output), c.Scheme)
		if err != nil {
			c.Eprintf("%s %s\n", printer.Serrorf("Failed to output workload:"), err)
			return cli.SilenceError(err)
		}

		c.Printf("%s\n", export)
		return nil
	}

	c.Printf(printer.ResourceStatus(workload.Name, printer.FindCondition(workload.Status.Conditions, cartov1alpha1.WorkloadConditionReady)))

	if workload.Spec.Image != "" || workload.Spec.Source != nil {
		c.Printf("\n")
		c.Printf("Source\n")

		if workload.Spec.Image != "" {
			if err := printer.WorkloadSourceImagePrinter(c.Stdout, workload); err != nil {
				return err
			}
		}

		if workload.Spec.Source != nil {
			if workload.Spec.Source.Image != "" {
				if err := printer.WorkloadLocalSourceImagePrinter(c.Stdout, workload); err != nil {
					return err
				}
			}
			if workload.Spec.Source.Git != nil {
				if err := printer.WorkloadSourceGitPrinter(c.Stdout, workload); err != nil {
					return err
				}
			}
		}
	}

	c.Printf("\n")
	if len(workload.Status.Resources) > 0 {
		if err := printer.WorkloadResourcesPrinter(c.Stdout, workload); err != nil {
			return err
		}
	} else {
		c.Infof("Supply Chain resources not found\n")
	}

	if len(workload.Spec.ServiceClaims) > 0 {
		c.Printf("\n")
		c.Printf("Services\n")
		if err := cartov1alpha1.WorkloadServiceClaimPrinter(c.Stdout, workload); err != nil {
			return err
		}
	}

	pods := &corev1.PodList{}
	err = c.List(ctx, pods, client.InNamespace(workload.Namespace), client.MatchingLabels{cartov1alpha1.WorkloadLabelName: workload.Name})
	if err != nil {
		c.Eprintf("\n")
		c.Eerrorf("Failed to list pods:\n")
		c.Eprintf("  %s\n", err)
	} else {
		if len(pods.Items) == 0 {
			c.Printf("\n")
			c.Infof("No pods found for workload.\n")
		} else {
			pods = pods.DeepCopy()
			printer.SortByNamespaceAndName(pods.Items)
			c.Printf("\n")
			c.Printf("Pods\n")
			if err := printer.PodTablePrinter(c, pods); err != nil {
				return err
			}
		}
	}

	ksvcs := &knativeservingv1.ServiceList{}
	_ = c.List(ctx, ksvcs, client.InNamespace(workload.Namespace), client.MatchingLabels{cartov1alpha1.WorkloadLabelName: workload.Name})
	if len(ksvcs.Items) > 0 {
		ksvcs = ksvcs.DeepCopy()
		printer.SortByNamespaceAndName(ksvcs.Items)
		c.Printf("\n")
		c.Printf("Knative Services\n")
		if err := printer.KnativeServicePrinter(c, ksvcs); err != nil {
			return err
		}
	}
	return nil
}

func NewWorkloadGetCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadGetOptions{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details from a workload",
		Long:  strings.TrimSpace(`Get details from a workload`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload get my-workload", c.Name),
		}, "\n"),
		PreRunE:           cli.ValidateE(ctx, opts),
		RunE:              cli.ExecE(ctx, c, opts),
		ValidArgsFunction: completion.SuggestWorkloadNames(ctx, c),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(ctx, cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.Export, cli.StripDash(flags.ExportFlagName), false, "export workload in yaml format")
	cmd.Flags().StringVarP(&opts.Output, cli.StripDash(flags.OutputFlagName), "o", "", "output the Workload formatted. Supported formats: \"json\", \"yaml\", \"yml\"")

	return cmd
}
