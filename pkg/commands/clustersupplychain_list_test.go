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
	"time"

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
	//timezone differences and daylight savings are causing issues
	//and when comparing against time.now, the output is not always 2Y as expected.
	//for example; when creating timestamp in time.Now will be in Central Standard Time,
	// but that same date 2 years in the future is Central Daylight Time. setting a default location(with no day light savings)
	// to mitiage this problem.
	loc, _ := time.LoadLocation("America/Puerto_Rico")
	objTimeStamp := metav1.NewTime(time.Now().In(loc).AddDate(-2, 0, 0))
	parent := diecartov1alpha1.ClusterSupplyChainBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(supplyChainName)
			d.CreationTimestamp(objTimeStamp)
		})
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
				parent.
					StatusDie(func(d *diecartov1alpha1.SupplyChainStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.ClusterSupplyChainConditionReadyBlank.Status(metav1.ConditionTrue),
						)
					}),
			},
			ExpectOutput: `
NAME                READY   AGE
test-supply-chain   Ready   2y

To view details: "tanzu apps cluster-supply-chain get <name>"

`,
		},
		{
			Name: "lists an item with empty values",
			Args: []string{},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
NAME                READY       AGE
test-supply-chain   <unknown>   2y

To view details: "tanzu apps cluster-supply-chain get <name>"

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
