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

	diemetav1 "dies.dev/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diecartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
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
			GivenObjects: []client.Object{
				diecartov1alpha1.ClusterSupplyChainBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name(supplyChainName)
					}).
					StatusDie(func(d *diecartov1alpha1.SupplyChainStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.ClusterSupplyChainConditionReadyBlank.Status(metav1.ConditionTrue),
						)
					}),
			},
			ExpectOutput: `
NAME                READY   AGE
test-supply-chain   Ready   <unknown>

View supply chain details by running "tanzu apps cluster-supply-chain get <name>"

`,
		},
		{
			Name: "lists an item with empty values",
			Args: []string{},
			GivenObjects: []client.Object{
				diecartov1alpha1.ClusterSupplyChainBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name(supplyChainName)
					}),
			},
			ExpectOutput: `
NAME                READY       AGE
test-supply-chain   <unknown>   <unknown>

View supply chain details by running "tanzu apps cluster-supply-chain get <name>"

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
