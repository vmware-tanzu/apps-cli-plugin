/*
Copyright 2019 VMware, Inc.

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

package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

const (
	AllNamespacesFlagName = "--all-namespaces"
	ContextFlagName       = "--context"
	KubeConfigFlagName    = "--kubeconfig"
	NamespaceFlagName     = "--namespace"
	NoColorFlagName       = "--no-color"
)

func AllNamespacesFlag(ctx context.Context, cmd *cobra.Command, c *Config, namespace *string, allNamespaces *bool) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *allNamespaces {
			if cmd.Flag(StripDash(NamespaceFlagName)).Changed {
				// forbid --namespace alongside --all-namespaces
				// Check here since we need the Flag to know if the namespace came from a flag
				return validation.ErrMultipleOneOf(NamespaceFlagName, AllNamespacesFlagName).ToAggregate()
			}
			*namespace = ""
		}
		if prior != nil {
			if err := prior(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}

	NamespaceFlag(ctx, cmd, c, namespace)
	cmd.Flags().BoolVarP(allNamespaces, StripDash(AllNamespacesFlagName), "A", false, "use all kubernetes namespaces")
}

func NamespaceFlag(ctx context.Context, cmd *cobra.Command, c *Config, namespace *string) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *namespace == "" {
			*namespace = c.DefaultNamespace()
		}
		if prior != nil {
			if err := prior(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}

	cmd.Flags().StringVarP(namespace, StripDash(NamespaceFlagName), "n", "", "kubernetes `name`space (defaulted from kube config)")
	cmd.RegisterFlagCompletionFunc(StripDash(NamespaceFlagName), func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		suggestions := []string{}
		namespaces := &corev1.NamespaceList{}
		err := c.List(ctx, namespaces)
		if err != nil {
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}
		for _, n := range namespaces.Items {
			suggestions = append(suggestions, n.Name)
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	})
}

func StripDash(flagName string) string {
	return strings.Replace(flagName, "--", "", 1)
}
