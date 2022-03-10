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

	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
)

func TestClusterSupplyChainListOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:           "empty",
			Validatable:    &commands.ClusterSupplyChainListOptions{},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestClusterSupplyChainListCommand(t *testing.T) {
	supplyChainName := "test-supply-chain"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	table := clitesting.CommandTestSuite{
		{
			Name: "empty",
			Args: []string{},
			ExpectOutput: `
No cluster supply chains found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.ClusterSupplyChain{
					ObjectMeta: metav1.ObjectMeta{
						Name: supplyChainName,
					},
					Spec: cartov1alpha1.SupplyChainSpec{
						Selector: map[string]string{
							"app.tanzu.vmware.com/workload-type": "web",
							"app.kubernetes.io/component":        "runtime",
						},
					},
					Status: cartov1alpha1.SupplyChainStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				}),
			},
			ExpectOutput: `
NAME                READY   AGE         LABEL SELECTOR
test-supply-chain   Ready   <unknown>   app.kubernetes.io/component=runtime,app.tanzu.vmware.com/workload-type=web
`,
		},
		{
			Name: "lists an item with empty values",
			Args: []string{},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.ClusterSupplyChain{
					ObjectMeta: metav1.ObjectMeta{
						Name: supplyChainName,
					},
				}),
			},
			ExpectOutput: `
NAME                READY       AGE         LABEL SELECTOR
test-supply-chain   <unknown>   <unknown>   <empty>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("list", "ClusterSupplyChainList"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, scheme, commands.NewClusterSupplyChainListCommand)
}
