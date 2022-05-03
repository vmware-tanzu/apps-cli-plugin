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

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diev1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
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
		{
			Name: "valid output format",
			Validatable: &commands.WorkloadListOptions{
				Namespace: "default",
				Output:    "json",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid output format",
			Validatable: &commands.WorkloadListOptions{
				Namespace: "default",
				Output:    "myFormat",
			},
			ExpectFieldErrors: validation.EnumInvalidValue("myFormat", flags.OutputFlagName, []string{"json", "yaml", "yml"}),
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
	_ = corev1.AddToScheme(scheme)

	parent := diev1alpha1.WorkloadBlank.
		MetadataDie(
			func(d *diemetav1.ObjectMetaDie) {
				d.Name(workloadName)
				d.Namespace(defaultNamespace)
			})

	table := clitesting.CommandTestSuite{
		{
			Name: "empty",
			Args: []string{},
			GivenObjects: []client.Object{
				diecorev1.NamespaceBlank.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.Name(defaultNamespace)
					},
				),
			},
			ExpectOutput: `
No workloads found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
NAME            APP       READY       AGE
test-workload   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "lists all items in json format",
			Args: []string{flags.OutputFlagName, "json"},
			GivenObjects: []client.Object{
				parent.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.CreationTimestamp(metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC))
					},
				),
			},
			ExpectOutput: `
[
	{
		"kind": "Workload",
		"apiVersion": "carto.run/v1alpha1",
		"metadata": {
			"name": "test-workload",
			"namespace": "default",
			"resourceVersion": "999",
			"creationTimestamp": "2021-09-10T15:00:00Z"
		},
		"spec": {},
		"status": {
			"supplyChainRef": {}
		}
	}
]
`,
		},
		{
			Name: "lists all items in yml format",
			Args: []string{flags.OutputFlagName, "yml"},
			GivenObjects: []client.Object{
				parent.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.CreationTimestamp(metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC))
					},
				),
			},
			ExpectOutput: `
---
- apiVersion: carto.run/v1alpha1
  kind: Workload
  metadata:
    creationTimestamp: "2021-09-10T15:00:00Z"
    name: test-workload
    namespace: default
    resourceVersion: "999"
  spec: {}
  status:
    supplyChainRef: {}
`,
		},
		{
			Name: "lists an item, with detail",
			Args: []string{},
			GivenObjects: []client.Object{
				parent.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.Labels(
							map[string]string{
								apis.AppPartOfLabelName: "hello",
							},
						)
					},
				).StatusDie(
					func(d *diev1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diev1alpha1.WorkloadConditionReadyBlank.Status(metav1.ConditionTrue),
						)
					},
				),
			},
			ExpectOutput: `
NAME            APP     READY   AGE
test-workload   hello   Ready   <unknown>
`,
		},
		{
			Name: "filters by app",
			Args: []string{flags.AppFlagName, "hello"},
			GivenObjects: []client.Object{
				parent.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.Labels(
							map[string]string{
								apis.AppPartOfLabelName: "hello",
							},
						)
					}),
				diev1alpha1.WorkloadBlank.
					MetadataDie(
						func(d *diemetav1.ObjectMetaDie) {
							d.Name(workloadOtherName)
							d.Namespace(defaultNamespace)
						}),
			},
			ExpectOutput: `
NAME            READY       AGE
test-workload   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace that exists",
			Args: []string{flags.NamespaceFlagName, otherNamespace},
			GivenObjects: []client.Object{
				diecorev1.NamespaceBlank.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.Name(otherNamespace)
					},
				),
				parent,
				diev1alpha1.WorkloadBlank.
					MetadataDie(
						func(d *diemetav1.ObjectMetaDie) {
							d.Name("something-else")
						}),
			},
			ExpectOutput: `
No workloads found.
`,
		},
		{
			Name: "namespace not found",
			Args: []string{flags.NamespaceFlagName, "foo"},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Namespace", clitesting.InduceFailureOpts{
					Error: apierrors.NewNotFound(corev1.Resource("Namespace"), "foo"),
				}),
			},
			ShouldError: true,
			ExpectOutput: `
Error: namespace "foo" not found, it may not exist or user does not have permissions to read it.
`,
		},
		{
			Name: "all namespace",
			Args: []string{flags.AllNamespacesFlagName},
			GivenObjects: []client.Object{
				diecorev1.NamespaceBlank.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.Name(otherNamespace)
					},
				),
				parent,
				diev1alpha1.WorkloadBlank.
					MetadataDie(
						func(d *diemetav1.ObjectMetaDie) {
							d.Name("test-other-workload")
							d.Namespace(otherNamespace)
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
