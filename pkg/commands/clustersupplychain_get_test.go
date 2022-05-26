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

package commands_test

import (
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	diemetav1 "dies.dev/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diecartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
)

func TestSupplyChainGetOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:        "invalid empty",
			Validatable: &commands.ClusterSupplyChainGetOptions{},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "valid",
			Validatable: &commands.ClusterSupplyChainGetOptions{
				Name: "my-csc",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Validatable: &commands.ClusterSupplyChainGetOptions{
				Name: "my-",
			},
			ShouldValidate: true,
		},
	}
	table.Run(t)
}

func TestClusterSupplyChainGetCommand(t *testing.T) {
	supplyChainName := "test-supply-chain"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	parent := diecartov1alpha1.ClusterSupplyChainBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(supplyChainName)
		})

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:         "valid",
			Args:         []string{supplyChainName},
			GivenObjects: []client.Object{parent},
			ExpectOutput: `
---
# test-supply-chain: <unknown>
---
Supply Chain Selectors
No supply chain selectors found
`,
		}, {
			Name: "label selectors",
			Args: []string{supplyChainName},
			GivenObjects: []client.Object{parent.
				SpecDie(func(d *diecartov1alpha1.SupplyChainSpecDie) {
					d.Selector(map[string]string{
						"apps.tanzu.vmware.com/workload-type":               "web",
						"apps.tanzu.vmware.com/workload-deployment-cluster": "test",
					})
				},
				)},
			ExpectOutput: `
---
# test-supply-chain: <unknown>
---
Supply Chain Selectors
TYPE     KEY                                                 OPERATOR   VALUE
labels   apps.tanzu.vmware.com/workload-deployment-cluster              test
labels   apps.tanzu.vmware.com/workload-type                            web
`,
		}, {
			Name: "all selectors",
			Args: []string{supplyChainName},
			GivenObjects: []client.Object{parent.
				SpecDie(func(d *diecartov1alpha1.SupplyChainSpecDie) {
					d.Selector(map[string]string{
						"apps.tanzu.vmware.com/workload-type": "web",
					})
					d.SelectorMatchFields(
						cartov1alpha1.FieldSelectorRequirement{
							Key:      "spec.image",
							Operator: cartov1alpha1.FieldSelectorOperator("Exists"),
						})
					d.SelectorMatchExpressions(
						metav1.LabelSelectorRequirement{
							Key:      "foo",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"bar"},
						})
				},
				)},

			ExpectOutput: `
---
# test-supply-chain: <unknown>
---
Supply Chain Selectors
TYPE          KEY                                   OPERATOR   VALUE
labels        apps.tanzu.vmware.com/workload-type              web
fields        spec.image                            Exists
expressions   foo                                   In         bar
`,
		}, {
			Name: "not found",
			Args: []string{supplyChainName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "ClusterSupplyChain", clitesting.InduceFailureOpts{
					Error: apierrors.NewNotFound(cartov1alpha1.Resource("ClusterSupplyChain"), supplyChainName),
				}),
			},
			ExpectOutput: `
Cluster Supply chain "test-supply-chain" not found
`,
			ShouldError: true,
		}, {
			Name: "get error",
			Args: []string{supplyChainName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "ClusterSupplyChain"),
			},
			ShouldError: true,
		},
	}
	table.Run(t, scheme, commands.NewClusterSupplyChainGetCommand)
}
