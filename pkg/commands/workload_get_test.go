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
	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diecartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
	diev1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/knative/serving/v1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestWorkloadGetOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:        "invalid empty",
			Validatable: &commands.WorkloadGetOptions{},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(flags.NamespaceFlagName),
				validation.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "valid",
			Validatable: &commands.WorkloadGetOptions{
				Namespace: "default",
				Name:      "my-workload",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Validatable: &commands.WorkloadGetOptions{
				Namespace: "default",
				Name:      "my-",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid namespace",
			Validatable: &commands.WorkloadGetOptions{
				Namespace: "default-",
				Name:      "my",
			},
			ShouldValidate: true,
		},
		{
			Name: "export",
			Validatable: &commands.WorkloadGetOptions{
				Namespace: "default",
				Name:      "my-workload",
				Export:    true,
			},
			ShouldValidate: true,
		},
		{
			Name: "valid output format",
			Validatable: &commands.WorkloadGetOptions{
				Namespace: "default",
				Name:      "my-workload",
				Output:    "json",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid output format",
			Validatable: &commands.WorkloadGetOptions{
				Namespace: "default",
				Name:      "my-workload",
				Output:    "myFormat",
			},
			ExpectFieldErrors: validation.EnumInvalidValue("myFormat", flags.OutputFlagName, []string{"json", "yaml", "yml"}),
		},
	}

	table.Run(t)
}

func TestWorkloadGetCommand(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	url := "https://example.com"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = knativeservingv1.AddToScheme(scheme)

	parent := diecartov1alpha1.WorkloadBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(workloadName)
			d.Namespace(defaultNamespace)
		})

	pod1Die := diecorev1.PodBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("pod1")
			d.Namespace(defaultNamespace)
			d.AddLabel(cartov1alpha1.WorkloadLabelName, workloadName)
		})

	pod2Die := diecorev1.PodBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("pod2")
			d.Namespace(defaultNamespace)
			d.AddLabel(cartov1alpha1.WorkloadLabelName, workloadName)
		})
	ksvcDieWithURL := diev1.ServiceBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("ksvc1")
			d.Namespace(defaultNamespace)
			d.AddLabel(cartov1alpha1.WorkloadLabelName, workloadName)
		}).
		StatusDie(func(d *diev1.ServiceStatusDie) {
			d.Conditions(
				metav1.Condition{
					Status: metav1.ConditionTrue,
					Type:   knativeservingv1.ServiceConditionReady,
				},
			)
			d.URL(url)
		})
	ksvcDieWithNoURL := diev1.ServiceBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("ksvc2")
			d.Namespace(defaultNamespace)
			d.AddLabel(cartov1alpha1.WorkloadLabelName, workloadName)
		}).
		StatusDie(func(d *diev1.ServiceStatusDie) {
			d.Conditions(
				metav1.Condition{
					Status: metav1.ConditionFalse,
					Type:   knativeservingv1.ServiceConditionReady,
				},
			)
		})

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		}, {
			Name:         "no supply chain info",
			Args:         []string{workloadName},
			GivenObjects: []client.Object{parent},
			ExpectOutput: `
---
# my-workload: <unknown>
---
Supply Chain reference not found.

Supply Chain resources not found.

No pods found for workload.
`,
		}, {
			Name: "no supply chain ref but conditions in status",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						)
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          <none>
last update:   <unknown>
ready:         False

Supply Chain resources not found.

No pods found for workload.
`,
		}, {
			Name: "supply chain ref but no condition in status",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
			},
			ExpectOutput: `
---
# my-workload: <unknown>
---
Supply Chain
name:          my-supply-chain
last update:   
ready:         

Supply Chain resources not found.

No pods found for workload.
`,
		}, {
			Name: "show status and service ref",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.ServiceClaims(cartov1alpha1.WorkloadServiceClaim{
							Name: "database",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "PostgreSQL",
								Name:       "my-prod-db",
							},
						})
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         False

Supply Chain resources not found.

Services
CLAIM      NAME         KIND         API VERSION
database   my-prod-db   PostgreSQL   services.tanzu.vmware.com/v1alpha1

No pods found for workload.
`,
		},
		{
			Name: "show status",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         False

Supply Chain resources not found.

No pods found for workload.
`,
		},
		{
			Name: "show source info - git",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: url,
								Ref: cartov1alpha1.GitRef{
									Branch: "master",
									Tag:    "v1.0.0",
									Commit: "abcdef",
								},
							},
						})
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         False

Source
type:     git
url:      https://example.com
branch:   master
tag:      v1.0.0
commit:   abcdef

Supply Chain resources not found.

No pods found for workload.
`,
		},
		{
			Name: "show source info - local path",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(
							&cartov1alpha1.Source{
								Image: "my-registry/my-image:v1.0.0",
							},
						)
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         False

Source
type:    source image
image:   my-registry/my-image:v1.0.0

Supply Chain resources not found.

No pods found for workload.
`,
		},
		{
			Name: "show source info - image",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("docker.io/library/nginx:latest")
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         False

Source
type:    image
image:   docker.io/library/nginx:latest

Supply Chain resources not found.

No pods found for workload.
`,
		},
		{
			Name: "show resources",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: url,
								Ref: cartov1alpha1.GitRef{
									Branch: "master",
									Tag:    "v1.0.0",
									Commit: "abcdef",
								},
							},
						})
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionFalse).Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
						d.Resources(
							diecartov1alpha1.RealizedResourceBlank.
								Name("source-provider").
								ConditionsDie(
									diecartov1alpha1.WorkloadConditionResourceReadyBlank.
										Status(metav1.ConditionTrue),
									diecartov1alpha1.WorkloadConditionResourceSubmittedBlank.
										Status(metav1.ConditionTrue),
								).DieRelease(),
							diecartov1alpha1.RealizedResourceBlank.
								Name("deliverable").
								ConditionsDie(
									diecartov1alpha1.WorkloadConditionResourceReadyBlank.
										Status(metav1.ConditionUnknown),
									diecartov1alpha1.WorkloadConditionResourceSubmittedBlank.
										Status(metav1.ConditionUnknown),
								).DieRelease(),
							diecartov1alpha1.RealizedResourceBlank.
								Name("image-builder").
								ConditionsDie(
									diecartov1alpha1.WorkloadConditionResourceReadyBlank.
										Status(metav1.ConditionFalse),
									diecartov1alpha1.WorkloadConditionResourceSubmittedBlank.
										Status(metav1.ConditionFalse),
								).DieRelease(),
						)
					}),
			},
			ExpectOutput: `
---
# my-workload: OopsieDoodle
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         False

Source
type:     git
url:      https://example.com
branch:   master
tag:      v1.0.0
commit:   abcdef

RESOURCE          READY     TIME
source-provider   True      <unknown>
deliverable       Unknown   <unknown>
image-builder     False     <unknown>

No pods found for workload.
`,
		},
		{
			Name: "show pods",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionUnknown).
								Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
				pod1Die.
					StatusDie(func(d *diecorev1.PodStatusDie) {
						d.Phase(corev1.PodRunning)
					}),
				pod2Die.
					StatusDie(func(d *diecorev1.PodStatusDie) {
						d.Phase(corev1.PodFailed)
					}),
				diecorev1.PodBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("pod-something-else")
						d.Namespace(defaultNamespace)
						d.AddLabel(cartov1alpha1.WorkloadLabelName, "diff-workload")
					}).
					StatusDie(func(d *diecorev1.PodStatusDie) {
						d.Phase(corev1.PodFailed)
					}),
			},
			ExpectOutput: `
---
# my-workload: Unknown
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         Unknown

Supply Chain resources not found.

Pods
NAME   STATUS    RESTARTS   AGE
pod1   Running   0          <unknown>
pod2   Failed    0          <unknown>
`,
		},
		{
			Name: "show knative services",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionUnknown).
								Reason("OopsieDoodle").
								Message("a hopefully informative message about what went wrong"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
				ksvcDieWithURL,
				ksvcDieWithNoURL,
				diev1.ServiceBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("ksvc3")
						d.Namespace(defaultNamespace)
						d.AddLabel(cartov1alpha1.WorkloadLabelName, "diff-workload")
					}).
					StatusDie(func(d *diev1.ServiceStatusDie) {
						d.Conditions(
							metav1.Condition{
								Status: metav1.ConditionTrue,
								Type:   knativeservingv1.ServiceConditionReady,
							},
						)
					}),
			},
			ExpectOutput: `
---
# my-workload: Unknown
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         Unknown

Supply Chain resources not found.

No pods found for workload.

Knative Services
NAME    READY       URL
ksvc1   Ready       https://example.com
ksvc2   not-Ready   <empty>
`,
		},
		{
			Name: "show pods and knative services",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionTrue).
								Reason("Worked").
								Message("Ready"),
						).SupplyChainRef(cartov1alpha1.ObjectReference{
							APIVersion: "supplychains.tanzu.vmware.com/v1alpha1",
							Kind:       "SupplyChain",
							Name:       "my-supply-chain",
							Namespace:  defaultNamespace,
						})
					}),
				ksvcDieWithURL,
				ksvcDieWithNoURL,
				pod1Die.
					StatusDie(func(d *diecorev1.PodStatusDie) {
						d.Phase(corev1.PodRunning)
					}),
				pod2Die.
					StatusDie(func(d *diecorev1.PodStatusDie) {
						d.Phase(corev1.PodFailed)
					}),
			},
			ExpectOutput: `
---
# my-workload: Ready
---
Supply Chain
name:          my-supply-chain
last update:   <unknown>
ready:         True

Supply Chain resources not found.

Pods
NAME   STATUS    RESTARTS   AGE
pod1   Running   0          <unknown>
pod2   Failed    0          <unknown>

Knative Services
NAME    READY       URL
ksvc1   Ready       https://example.com
ksvc2   not-Ready   <empty>
`,
		},
		{
			Name: "not found",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				diecorev1.NamespaceBlank.MetadataDie(
					func(d *diemetav1.ObjectMetaDie) {
						d.Name("default")
					},
				),
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("notfound")
					}),
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Workload", clitesting.InduceFailureOpts{
					Error: apierrors.NewNotFound(cartov1alpha1.Resource("Workload"), workloadName),
				}),
			},
			ExpectOutput: `
Workload "default/my-workload" not found
`,
			ShouldError: true,
		},
		{
			Name: "namespace not found",
			Args: []string{workloadName, flags.NamespaceFlagName, "foo"},
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
			Name: "get error",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionTrue).Reason("OopsieDoodle").
								Message("a hopefully informative message").
								LastTransitionTime(metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								}),
						)
					}),
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Workload"),
			},
			ShouldError: true,
		},
		{
			Name: "get error for listing pods",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("list", "PodList"),
			},
			ExpectOutput: `
---
# my-workload: <unknown>
---
Supply Chain reference not found.

Supply Chain resources not found.

Failed to list pods:
  inducing failure for list PodList
`,
		},
		{
			Name: "get error for listing knative services",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("list", "KnativeServiceList"),
			},
			ExpectOutput: `
---
# my-workload: <unknown>
---
Supply Chain reference not found.

Supply Chain resources not found.

No pods found for workload.
`,
		},
		{
			Name: "get workload exported data",
			Args: []string{workloadName, flags.ExportFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel(apis.AppPartOfLabelName, workloadName)
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionUnknown).Reason("Workload Reason").
								Message("a hopefully informative message about what went wrong"),
						)
					}),
			},
			ExpectOutput: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  labels:
    app.kubernetes.io/part-of: my-workload
  name: my-workload
  namespace: default
spec: {}
`,
		},
		{
			Name: "get workload exported data in json format",
			Args: []string{workloadName, flags.ExportFlagName, flags.OutputFlagName, printer.OutputFormatJson},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel(apis.AppPartOfLabelName, workloadName)
						d.CreationTimestamp(metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC))
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionUnknown).
								Reason("Workload Reason").
								Message("a hopefully informative message about what went wrong"),
						)
					}),
			},
			ExpectOutput: `
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"labels": {
			"app.kubernetes.io/part-of": "my-workload"
		},
		"name": "my-workload",
		"namespace": "default"
	},
	"spec": {}
}
`,
		},
		{
			Name: "get workload outputted data in yaml format",
			Args: []string{workloadName, flags.OutputFlagName, "yaml"},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel(apis.AppPartOfLabelName, workloadName)
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionUnknown).
								Reason("Workload Reason").
								Message("a hopefully informative message about what went wrong"),
						)
					}),
			},
			ExpectOutput: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  labels:
    app.kubernetes.io/part-of: my-workload
  name: my-workload
  namespace: default
  resourceVersion: "999"
spec: {}
status:
  conditions:
  - lastTransitionTime: null
    message: a hopefully informative message about what went wrong
    reason: Workload Reason
    status: Unknown
    type: Ready
  supplyChainRef: {}
`,
		},
		{
			Name: "get workload outputted data in json format",
			Args: []string{workloadName, flags.OutputFlagName, "json"},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel(apis.AppPartOfLabelName, workloadName)
					}).
					StatusDie(func(d *diecartov1alpha1.WorkloadStatusDie) {
						d.ConditionsDie(
							diecartov1alpha1.WorkloadConditionReadyBlank.
								Status(metav1.ConditionUnknown).
								Reason("Workload Reason").
								Message("a hopefully informative message about what went wrong"),
						)
					}),
			},
			ExpectOutput: `
{
	"kind": "Workload",
	"apiVersion": "carto.run/v1alpha1",
	"metadata": {
		"name": "my-workload",
		"namespace": "default",
		"resourceVersion": "999",
		"creationTimestamp": null,
		"labels": {
			"app.kubernetes.io/part-of": "my-workload"
		}
	},
	"spec": {},
	"status": {
		"conditions": [
			{
				"type": "Ready",
				"status": "Unknown",
				"lastTransitionTime": null,
				"reason": "Workload Reason",
				"message": "a hopefully informative message about what went wrong"
			}
		],
		"supplyChainRef": {}
	}
}
`,
		},
	}

	table.Run(t, scheme, commands.NewWorkloadGetCommand)
}
