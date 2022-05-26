/*
Copyright 2022 VMware, Inc.

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
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type ClusterSupplyChainGetOptions struct {
	Name string
}

var (
	_ validation.Validatable = (*ClusterSupplyChainGetOptions)(nil)
	_ cli.Executable         = (*ClusterSupplyChainGetOptions)(nil)
)

func (opts *ClusterSupplyChainGetOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}
	if opts.Name == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NameArgumentName))
	}
	return errs
}

func (opts *ClusterSupplyChainGetOptions) Exec(ctx context.Context, c *cli.Config) error {
	supplyChain := &cartov1alpha1.ClusterSupplyChain{}
	err := c.Get(ctx, client.ObjectKey{Name: opts.Name}, supplyChain)
	if err != nil {
		if apierrs.IsNotFound(err) {
			c.Errorf("Cluster Supply chain %q not found\n", opts.Name)
			return cli.SilenceError(err)
		}
		return err
	}
	c.Printf(printer.ResourceStatus(supplyChain.Name, printer.FindCondition(supplyChain.Status.Conditions, cartov1alpha1.SupplyChainReady)))

	c.Printf("Supply Chain Selectors\n")
	if len(supplyChain.Spec.LegacySelector.Selector) == 0 && len(supplyChain.Spec.LegacySelector.SelectorMatchExpressions) == 0 && len(supplyChain.Spec.LegacySelector.SelectorMatchFields) == 0 {
		c.Infof("No supply chain selectors found\n")
	} else {
		if err := printer.ClusterSupplyChainPrinter(c.Stdout, supplyChain); err != nil {
			return err
		}
	}
	return nil
}
func NewClusterSupplyChainGetCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ClusterSupplyChainGetOptions{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details from a cluster supply chain",
		Long:  strings.TrimSpace(`Get details from a cluster supply chain`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s cluster-supply-chain get", c.Name),
		}, "\n"),
		PreRunE:           cli.ValidateE(ctx, opts),
		RunE:              cli.ExecE(ctx, c, opts),
		ValidArgsFunction: completion.SuggestClusterSupplyChainNames(ctx, c),
	}
	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	return cmd
}
