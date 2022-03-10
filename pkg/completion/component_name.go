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

package completion

import (
	"context"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func SuggestComponentNames(ctx context.Context, c *cli.Config) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		namespace := cmd.Flag(cli.StripDash(flags.NamespaceFlagName)).Value.String()
		if namespace == "" {
			namespace = c.DefaultNamespace()
		}

		pods := &corev1.PodList{}
		// Filter by component label and namespace
		if err := c.List(ctx, pods, client.InNamespace(namespace), client.HasLabels{apis.ComponentLabelName}); err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}

		suggestionsSet := sets.NewString()
		for _, n := range pods.Items {
			suggestionsSet.Insert(n.Labels[apis.ComponentLabelName])
		}

		return suggestionsSet.List(), cobra.ShellCompDirectiveNoFileComp
	}
}
