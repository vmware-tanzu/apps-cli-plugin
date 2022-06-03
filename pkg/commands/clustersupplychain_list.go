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

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

type ClusterSupplyChainListOptions struct {
	// none for now
}

var (
	_ validation.Validatable = (*ClusterSupplyChainListOptions)(nil)
	_ cli.Executable         = (*ClusterSupplyChainListOptions)(nil)
)

func (opts *ClusterSupplyChainListOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	// none for now

	return errs
}

func (opts *ClusterSupplyChainListOptions) Exec(ctx context.Context, c *cli.Config) error {
	supplyChain := &cartov1alpha1.ClusterSupplyChainList{}
	err := c.List(ctx, supplyChain)
	if err != nil {
		return err
	}

	if len(supplyChain.Items) == 0 {
		c.Infof("No cluster supply chains found.\n")
		return nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{
		// none for now
	}).With(func(h table.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	supplyChain = supplyChain.DeepCopy()
	printer.SortByNamespaceAndName(supplyChain.Items)

	if err := tablePrinter.PrintObj(supplyChain, c.Stdout); err != nil {
		return err
	}

	c.Printf("\n")
	c.Infof("View supply chain details by running \"tanzu apps cluster-supply-chain get <name>\"\n")
	c.Printf("\n")

	return nil
}

func NewClusterSupplyChainListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ClusterSupplyChainListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of cluster supply chains",
		Long: strings.TrimSpace(`
List cluster supply chains.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s cluster-supply-chain list", c.Name),
		}, "\n"),
		PreRunE: cli.ValidateE(ctx, opts),
		RunE:    cli.ExecE(ctx, c, opts),
	}

	return cmd
}

func (opts *ClusterSupplyChainListOptions) printList(supplyChains *cartov1alpha1.ClusterSupplyChainList, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(supplyChains.Items))
	for i := range supplyChains.Items {
		r, err := opts.print(&supplyChains.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *ClusterSupplyChainListOptions) print(supplyChain *cartov1alpha1.ClusterSupplyChain, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: supplyChain},
	}
	row.Cells = append(row.Cells,
		supplyChain.Name,
		printer.ConditionStatus(printer.FindCondition(supplyChain.Status.Conditions, "Ready")),
		printer.TimestampSince(supplyChain.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *ClusterSupplyChainListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Ready", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
