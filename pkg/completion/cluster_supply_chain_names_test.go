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

package completion_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
)

func TestSuggestClusterSupplyChainNames(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name               string
		scheme             *runtime.Scheme
		given              []client.Object
		reactor            clitesting.ReactionFunc
		sugestions         []string
		shellCompDirective cobra.ShellCompDirective
	}{{
		name:               "no supply chains",
		scheme:             scheme,
		given:              []client.Object{},
		reactor:            nil,
		sugestions:         []string{},
		shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
	}, {
		name:   "supply chains",
		scheme: scheme,
		given: []client.Object{
			&cartov1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: "default",
				},
			},
			&cartov1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "barfoo",
					Namespace: "default",
				},
			},
		},
		reactor: nil,
		sugestions: []string{
			"barfoo",
			"foobar",
		},
		shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
	}, {
		name:   "list error",
		scheme: scheme,
		given: []client.Object{
			&cartov1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foobar",
				},
			},
			&cartov1alpha1.ClusterSupplyChain{
				ObjectMeta: metav1.ObjectMeta{
					Name: "barfoo",
				},
			},
		},
		reactor:            clitesting.InduceFailure("list", "ClusterSupplyChainList"),
		sugestions:         []string{},
		shellCompDirective: cobra.ShellCompDirectiveError,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()

			c := cli.NewDefaultConfig("test", scheme)
			client := clitesting.NewFakeClient(scheme, test.given...)
			if test.reactor != nil {
				client.AddReactor("*", "*", test.reactor)
			}
			c.Client = clitesting.NewFakeCliClient(client)
			cmd := &cobra.Command{}
			//cmd.Flags().String("namespace", test.namespace, "")

			suggestions, directive := completion.SuggestClusterSupplyChainNames(ctx, c)(cmd, []string{}, "")
			if diff := cmp.Diff(suggestions, test.sugestions); diff != "" {
				t.Errorf("SuggestClusterSupplyChainNames() sugestions (-want, +got) = %v", diff)

			}
			if want, got := test.shellCompDirective, directive; want != got {
				t.Errorf("SuggestClusterSupplyChainNames() ShellCompDirective: want %d, got %d", want, got)
			}
		})
	}
}
