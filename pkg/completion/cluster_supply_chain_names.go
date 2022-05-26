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

package completion

import (
	"context"

	"github.com/spf13/cobra"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func SuggestClusterSupplyChainNames(ctx context.Context, c *cli.Config) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		suggestions := []string{}
		clustersupplychains := &cartov1alpha1.ClusterSupplyChainList{}

		err := c.List(ctx, clustersupplychains)
		if err != nil {
			return suggestions, cobra.ShellCompDirectiveError
		}
		for _, w := range clustersupplychains.Items {
			suggestions = append(suggestions, w.Name)
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}
}
