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

package commands_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestWorkloadListOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:              "empty",
			Validatable:       &commands.WorkloadListOptions{},
			ExpectFieldErrors: validation.ErrMissingOneOf(flags.NamespaceFlagName, flags.AllNamespacesFlagName),
		},
		{
			Name: "invalid namespace + all",
			Validatable: &commands.WorkloadListOptions{
				Namespace:     "default",
				AllNamespaces: true,
			},
			ExpectFieldErrors: validation.ErrMultipleOneOf(flags.NamespaceFlagName, flags.AllNamespacesFlagName),
		},
		{
			Name: "app",
			Validatable: &commands.WorkloadListOptions{
				Namespace: "default",
				App:       "hello",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid app",
			Validatable: &commands.WorkloadListOptions{
				Namespace: "default",
				App:       "hello-",
			},
			ExpectFieldErrors: validation.ErrInvalidValue("hello-", flags.AppFlagName),
		},
	}

	table.Run(t)
}

func TestWorkloadListCommand(t *testing.T) {
	workloadName := "test-workload"
	workloadOtherName := "test-other-workload"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	table := clitesting.CommandTestSuite{
		{
			Name: "empty",
			Args: []string{},
			ExpectOutput: `
No workloads found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
				}),
			},
			ExpectOutput: `
NAME            APP       READY       AGE
test-workload   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "lists an item, with detail",
			Args: []string{},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							apis.AppPartOfLabelName: "hello",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{},
					Status: cartov1alpha1.WorkloadStatus{
						OwnerStatus: cartov1alpha1.OwnerStatus{
							Conditions: []metav1.Condition{
								{
									Type:   cartov1alpha1.WorkloadConditionReady,
									Status: metav1.ConditionTrue,
								},
							},
						},
					},
				}),
			},
			ExpectOutput: `
NAME            APP     READY   AGE
test-workload   hello   Ready   <unknown>
`,
		},
		{
			Name: "filters by app",
			Args: []string{flags.AppFlagName, "hello"},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadOtherName,
						Namespace: defaultNamespace,
					},
				}),
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							apis.AppPartOfLabelName: "hello",
						},
					},
				}),
			},
			ExpectOutput: `
NAME            READY       AGE
test-workload   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{flags.NamespaceFlagName, otherNamespace},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
				}),
			},
			ExpectOutput: `
No workloads found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{flags.AllNamespacesFlagName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
				}),
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadOtherName,
						Namespace: otherNamespace,
					},
				}),
			},
			ExpectOutput: `
NAMESPACE         NAME                  APP       READY       AGE
default           test-workload         <empty>   <unknown>   <unknown>
other-namespace   test-other-workload   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("list", "WorkloadList"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, scheme, commands.NewWorkloadListCommand)
}
