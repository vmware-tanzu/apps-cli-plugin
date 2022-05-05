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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
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

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		}, {
			Name: "show status and service ref",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{{
							Name: "database",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "PostgreSQL",
								Name:       "my-prod-db",
							},
						}},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
			},
			ExpectOutput: `
# my-workload: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready

Supply Chain resources not found

Services
CLAIM      NAME         KIND         API VERSION
database   my-prod-db   PostgreSQL   services.tanzu.vmware.com/v1alpha1

No pods found for workload.
`,
		},
		{
			Name: "show status",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
			},
			ExpectOutput: `
# my-workload: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready

Supply Chain resources not found

No pods found for workload.
`,
		},
		{
			Name: "show source info - git",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: url,
								Ref: cartov1alpha1.GitRef{
									Branch: "master",
									Tag:    "v1.0.0",
									Commit: "abcdef",
								},
							},
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
			},
			ExpectOutput: `
# my-workload: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready

Source
type:     git
url:      https://example.com
branch:   master
tag:      v1.0.0
commit:   abcdef

Supply Chain resources not found

No pods found for workload.
`,
		},
		{
			Name: "show source info - local path",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Image: "my-registry/my-image:v1.0.0",
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
			},
			ExpectOutput: `
# my-workload: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready

Source
type:    source image
image:   my-registry/my-image:v1.0.0

Supply Chain resources not found

No pods found for workload.
`,
		},
		{
			Name: "show source info - image",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "docker.io/library/nginx:latest",
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
			},
			ExpectOutput: `
# my-workload: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready

Source
type:    image
image:   docker.io/library/nginx:latest

Supply Chain resources not found

No pods found for workload.
`,
		},
		{
			Name: "show resources",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: url,
								Ref: cartov1alpha1.GitRef{
									Branch: "master",
									Tag:    "v1.0.0",
									Commit: "abcdef",
								},
							},
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
						Resources: []cartov1alpha1.RealizedResource{{
							Name: "source-provider",
							Conditions: []metav1.Condition{
								{
									Type:   "Ready",
									Status: "True",
								},
								{
									Type:   "ResourceSubmitted",
									Status: "True",
								},
							},
						}, {
							Name: "deliverable",
							Conditions: []metav1.Condition{
								{
									Type:   "Ready",
									Status: "Unknown",
								},
								{
									Type:   "ResourceSubmitted",
									Status: "Unknown",
								},
							},
						}, {
							Name: "image-builder",
							Conditions: []metav1.Condition{
								{
									Type:   "Ready",
									Status: "False",
								},
								{
									Type:   "ResourceSubmitted",
									Status: "False",
								},
							},
						}},
					},
				}),
			},
			ExpectOutput: `
# my-workload: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready

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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionUnknown,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
				clitesting.Wrapper(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}),
				clitesting.Wrapper(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				}),
				clitesting.Wrapper(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "diff-workload-pod",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: "diff-workload",
						},
					},
				}),
			},
			ExpectOutput: `
# my-workload: Unknown
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: Unknown
type: Ready

Supply Chain resources not found

Pods
NAME   STATUS    RESTARTS   AGE
pod1   Running   0          <unknown>
pod2   Failed    0          <unknown>
`,
		},
		{
			Name: "show knative services",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionUnknown,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
				clitesting.Wrapper(&knativeservingv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ksvc1",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: knativeservingv1.ServiceStatus{
						Conditions: []metav1.Condition{{
							Status: metav1.ConditionTrue,
							Type:   knativeservingv1.ServiceConditionReady,
						}},
						URL: url,
					},
				}),
				clitesting.Wrapper(&knativeservingv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ksvc2",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: knativeservingv1.ServiceStatus{
						Conditions: []metav1.Condition{{
							Status: metav1.ConditionFalse,
							Type:   knativeservingv1.ServiceConditionReady,
						}},
					},
				}),
				clitesting.Wrapper(&knativeservingv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ksvc3",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: "diff-workload",
						},
					},
					Status: knativeservingv1.ServiceStatus{
						Conditions: []metav1.Condition{{
							Status: metav1.ConditionTrue,
							Type:   knativeservingv1.ServiceConditionReady,
						}},
					},
				}),
			},
			ExpectOutput: `
# my-workload: Unknown
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: Unknown
type: Ready

Supply Chain resources not found

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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionTrue,
								Reason:  "Worked",
								Message: "Ready",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
				}),
				clitesting.Wrapper(&knativeservingv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ksvc1",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: knativeservingv1.ServiceStatus{
						Conditions: []metav1.Condition{{
							Status: metav1.ConditionTrue,
							Type:   knativeservingv1.ServiceConditionReady,
						}},
						URL: url,
					},
				}),
				clitesting.Wrapper(&knativeservingv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ksvc2",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: knativeservingv1.ServiceStatus{
						Conditions: []metav1.Condition{{
							Status: metav1.ConditionFalse,
							Type:   knativeservingv1.ServiceConditionReady,
						}},
					},
				}),
				clitesting.Wrapper(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}),
				clitesting.Wrapper(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							cartov1alpha1.WorkloadLabelName: workloadName,
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				}),
			},
			ExpectOutput: `
# my-workload: Ready
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: Ready
reason: Worked
status: "True"
type: Ready

Supply Chain resources not found

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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultNamespace,
					},
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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionTrue,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
								LastTransitionTime: metav1.Time{
									Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
								},
							},
						},
					},
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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
				}),
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("list", "PodList"),
			},
			ExpectOutput: `
# my-workload: <unknown>

Supply Chain resources not found

Failed to list pods:
  inducing failure for list PodList
`,
		},
		{
			Name: "get error for listing knative services",
			Args: []string{workloadName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
					},
				}),
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("list", "KnativeServiceList"),
			},
			ExpectOutput: `
# my-workload: <unknown>

Supply Chain resources not found

No pods found for workload.
`,
		},
		{
			Name: "get workload exported data",
			Args: []string{workloadName, flags.ExportFlagName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							apis.AppPartOfLabelName: workloadName,
						},
						CreationTimestamp: metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionUnknown,
								Reason:  "Workload Reason",
								Message: "a hopefully informative message about what went wrong",
							},
						},
					},
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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							apis.AppPartOfLabelName: workloadName,
						},
						CreationTimestamp: metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionUnknown,
								Reason:  "Workload Reason",
								Message: "a hopefully informative message about what went wrong",
							},
						},
					},
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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							apis.AppPartOfLabelName: workloadName,
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionUnknown,
								Reason:  "Workload Reason",
								Message: "a hopefully informative message about what went wrong",
							},
						},
					},
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
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workloadName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							apis.AppPartOfLabelName: workloadName,
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionUnknown,
								Reason:  "Workload Reason",
								Message: "a hopefully informative message about what went wrong",
							},
						},
					},
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
