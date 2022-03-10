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

package completion_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
)

func TestSuggestComponentNames(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	namespace := "default"

	tests := []struct {
		name               string
		given              []client.Object
		reactor            clitesting.ReactionFunc
		sugestions         []string
		shellCompDirective cobra.ShellCompDirective
	}{
		{
			name:               "no components",
			given:              []client.Object{},
			sugestions:         []string{},
			shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name: "1 component in default namespace",
			given: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foobar",
						Namespace: namespace,
						Labels:    map[string]string{apis.ComponentLabelName: "runtime"},
					},
				},
			},
			sugestions:         []string{"runtime"},
			shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name: "2 pods with the same component in default namespace",
			given: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foobar",
						Namespace: namespace,
						Labels:    map[string]string{apis.ComponentLabelName: "runtime"},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "barfoo",
						Namespace: namespace,
						Labels:    map[string]string{apis.ComponentLabelName: "runtime"},
					},
				},
			},
			sugestions:         []string{"runtime"},
			shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name: "2 pods with different components in default namespace",
			given: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foobar",
						Namespace: namespace,
						Labels:    map[string]string{apis.ComponentLabelName: "runtime"},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "barfoo",
						Namespace: namespace,
						Labels:    map[string]string{apis.ComponentLabelName: "build"},
					},
				},
			},
			sugestions:         []string{"build", "runtime"},
			shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name: "2 pods with different components in different namespaces",
			given: []client.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foobar",
						Namespace: "runtime",
						Labels:    map[string]string{apis.ComponentLabelName: "runtime"},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "barfoo",
						Namespace: namespace,
						Labels:    map[string]string{apis.ComponentLabelName: "build"},
					},
				},
			},
			sugestions:         []string{"build"},
			shellCompDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:               "list error",
			given:              []client.Object{},
			reactor:            clitesting.InduceFailure("list", "PodList"),
			sugestions:         []string{},
			shellCompDirective: cobra.ShellCompDirectiveError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			c := cli.NewDefaultConfig("test", scheme)
			client := clitesting.NewFakeClient(scheme, test.given...)
			if test.reactor != nil {
				client.AddReactor("*", "*", test.reactor)
			}

			c.Client = clitesting.NewFakeCliClient(client)
			cmd := &cobra.Command{}
			cmd.Flags().String("namespace", namespace, "")

			suggestions, directive := completion.SuggestComponentNames(ctx, c)(cmd, []string{}, "")
			if diff := cmp.Diff(suggestions, test.sugestions); diff != "" {
				t.Errorf("SuggestComponentNames() sugestions (-want, +got) = %v", diff)
			}
			if want, got := test.shellCompDirective, directive; want != got {
				t.Errorf("SuggestComponentNames() ShellCompDirective: want %d, got %d", want, got)
			}
		})
	}
}
