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
	"context"
	"fmt"
	"io"
	"net/http"
	runtm "runtime"
	"strings"
	"testing"
	"time"

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/logs"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	watchhelper "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/watch"
	watchfakes "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/watch/fake"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diecartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestWorkloadUpdateOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name: "valid options",
			Validatable: &commands.WorkloadUpdateOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					Env:       []string{"FOO=bar"},
					BuildEnv:  []string{"BAR=baz"},
				},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid options",
			Validatable: &commands.WorkloadUpdateOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					Env:       []string{"FOO"},
				},
			},
			ExpectFieldErrors: validation.ErrInvalidArrayValue("FOO", flags.EnvFlagName, 0),
		},
		{
			Name: "invalid build env options",
			Validatable: &commands.WorkloadUpdateOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					BuildEnv:  []string{"FOO"},
				},
			},
			ExpectFieldErrors: validation.ErrInvalidArrayValue("FOO", flags.BuildEnvFlagName, 0),
		},
	}

	table.Run(t)
}

func TestWorkloadUpdateCommand(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	var cmd *cobra.Command

	parent := diecartov1alpha1.WorkloadBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(workloadName)
			d.Namespace(defaultNamespace)
			d.Labels(map[string]string{
				apis.WorkloadTypeLabelName: "web",
			})
		})
	sprintPetclinicWorkload := diecartov1alpha1.WorkloadBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("spring-petclinic")
			d.Namespace(defaultNamespace)
		})
	respCreator := func(status int, body string) *http.Response {
		return &http.Response{
			Status:     http.StatusText(status),
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
		}
	}
	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "noop",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
					d.Image("ubuntu:bionic")
				},
				),
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

Workload is unchanged, skipping update
`,
		},
		{
			Name: "no source",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent,
			},
		},
		{
			Name: "not found",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				diecorev1.NamespaceBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name(defaultNamespace)
					}),
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("foo")
					}),
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Workload", clitesting.InduceFailureOpts{
					Error: apierrors.NewNotFound(cartov1alpha1.Resource("Workload"), workloadName),
				}),
			},
			ShouldError: true,
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

Workload "default/my-workload" not found
`,
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
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

Error: namespace "foo" not found, it may not exist or user does not have permissions to read it.
`,
		},
		{
			Name: "get failed",
			Args: []string{workloadName},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Workload"),
			},
			ShouldError: true,
		},
		{
			Name: "dry run",
			Args: []string{workloadName, flags.DryRunFlagName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "1970-01-01T00:00:01Z"
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: my-workload
  namespace: default
  resourceVersion: "999"
spec:
  image: ubuntu:bionic
status:
  supplyChainRef: {}
`,
		},
		{
			Name: "error during update",
			Args: []string{workloadName, flags.DebugFlagName, flags.YesFlagName},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("update", "Workload"),
			},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						Params: []cartov1alpha1.Param{
							{
								Name:  "debug",
								Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "update subPath for git source",
			Args: []string{workloadName, flags.SubPathFlagName, "./app", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(
							&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/spring-projects/spring-petclinic.git",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							},
						)
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
							Subpath: "./app",
						},
					},
				},
			},
		},
		{
			Name: "override subPath for source image source",
			Args: []string{workloadName, flags.SubPathFlagName, "./app", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(
							&cartov1alpha1.Source{
								Image:   "ubuntu:source",
								Subpath: "./cmd",
							},
						)
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Image:   "ubuntu:source",
							Subpath: "./app",
						},
					},
				},
			},
		},
		{
			Name: "unset subPath for git source",
			Args: []string{workloadName, flags.SubPathFlagName, "./app", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(
							&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/spring-projects/spring-petclinic.git",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							},
						)
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
							Subpath: "./app",
						},
					},
				},
			},
		},
		{
			Name: "conflict during update",
			Args: []string{workloadName, flags.DebugFlagName, flags.YesFlagName},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("update", "Workload", clitesting.InduceFailureOpts{
					Error: apierrors.NewConflict(schema.GroupResource{Group: "carto.run", Resource: "workloads"}, workloadName, fmt.Errorf("induced conflict")),
				}),
			},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						Params: []cartov1alpha1.Param{
							{
								Name:  "debug",
								Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
							},
						},
					},
				},
			},
			ShouldError: true,
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  params:
     12 + |  - name: debug
     13 + |    value: "true"
Error: conflict updating workload, the object was modified by another user; please run the update command again
`,
		},
		{
			Name: "wait error with timeout",
			Skip: runtm.GOOS == "windows",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.WaitFlagName, flags.YesFlagName, flags.WaitTimeoutFlagName, "1ns"},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:   cartov1alpha1.WorkloadConditionReady,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)
				return ctx, nil
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "PostgreSQL",
									Name:       "my-prod-db",
								},
							},
						},
					},
				},
			},
			ShouldError: true,
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: PostgreSQL
     16 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error: timeout after 1ns waiting for "my-workload" to become ready
`,
		},
		{
			Name: "wait error for false condition",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.WaitFlagName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:    cartov1alpha1.WorkloadConditionReady,
								Status:  metav1.ConditionFalse,
								Reason:  "OopsieDoodle",
								Message: "a hopefully informative message about what went wrong",
							},
						},
					},
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)
				return ctx, nil
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "PostgreSQL",
									Name:       "my-prod-db",
								},
							},
						},
					},
				},
			},
			ShouldError: true,
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: PostgreSQL
     16 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error: Failed to become ready: a hopefully informative message about what went wrong
`,
		},
		{
			Name: "successful wait for ready condition",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.WaitFlagName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:   cartov1alpha1.WorkloadConditionReady,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)
				return ctx, nil
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "PostgreSQL",
									Name:       "my-prod-db",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: PostgreSQL
     16 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Workload "my-workload" is ready
`,
		},
		{
			Name: "tail while waiting for ready condition",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.YesFlagName, flags.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:   cartov1alpha1.WorkloadConditionReady,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)

				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Minute, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)

				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "PostgreSQL",
									Name:       "my-prod-db",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: PostgreSQL
     16 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready
`,
		},
		{
			Name: "tail with timestamp while waiting for ready condition",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.YesFlagName, flags.TailFlagName, flags.TailTimestampFlagName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Status: cartov1alpha1.WorkloadStatus{
						Conditions: []metav1.Condition{
							{
								Type:   cartov1alpha1.WorkloadConditionReady,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)

				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Minute, true).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)

				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "PostgreSQL",
									Name:       "my-prod-db",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: PostgreSQL
     16 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready
`,
		},
		{
			Name: "filepath",
			Args: []string{flags.FilePathFlagName, "testdata/workload.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				sprintPetclinicWorkload.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(corev1.EnvVar{
							Name:  "OVERRIDE_VAR",
							Value: "doesnt matter",
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"preserve-me":                         "should-exist",
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
							{
								Name:  "SPRING_PROFILES_ACTIVE",
								Value: "mysql",
							},
						},
						Resources: &corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
  6,  8   |    preserve-me: should-exist
  7,  9   |  name: spring-petclinic
  8, 10   |  namespace: default
  9, 11   |spec:
 10, 12   |  env:
 11, 13   |  - name: OVERRIDE_VAR
 12, 14   |    value: doesnt matter
 13     - |  image: ubuntu:bionic
     15 + |  - name: SPRING_PROFILES_ACTIVE
     16 + |    value: mysql
     17 + |  resources:
     18 + |    limits:
     19 + |      cpu: 500m
     20 + |      memory: 1Gi
     21 + |    requests:
     22 + |      cpu: 100m
     23 + |      memory: 1Gi
     24 + |  source:
     25 + |    git:
     26 + |      ref:
     27 + |        branch: main
     28 + |      url: https://github.com/spring-projects/spring-petclinic.git
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "accept yaml file through stdin - using --yes flag",
			Args: []string{flags.FilePathFlagName, "-", flags.YesFlagName},
			Stdin: []byte(`
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: spring-petclinic
  labels:
    app.kubernetes.io/part-of: spring-petclinic
    apps.tanzu.vmware.com/workload-type: web
spec:
  env:
  - name: SPRING_PROFILES_ACTIVE
    value: mysql
  resources:
    requests:
      memory: 1Gi
      cpu: 100m
    limits:
      memory: 1Gi
      cpu: 500m
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main
`),
			GivenObjects: []client.Object{
				sprintPetclinicWorkload.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(corev1.EnvVar{
							Name:  "OVERRIDE_VAR",
							Value: "doesnt matter",
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"preserve-me":                         "should-exist",
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
							{
								Name:  "SPRING_PROFILES_ACTIVE",
								Value: "mysql",
							},
						},
						Resources: &corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
  6,  8   |    preserve-me: should-exist
  7,  9   |  name: spring-petclinic
  8, 10   |  namespace: default
  9, 11   |spec:
 10, 12   |  env:
 11, 13   |  - name: OVERRIDE_VAR
 12, 14   |    value: doesnt matter
 13     - |  image: ubuntu:bionic
     15 + |  - name: SPRING_PROFILES_ACTIVE
     16 + |    value: mysql
     17 + |  resources:
     18 + |    limits:
     19 + |      cpu: 500m
     20 + |      memory: 1Gi
     21 + |    requests:
     22 + |      cpu: 100m
     23 + |      memory: 1Gi
     24 + |  source:
     25 + |    git:
     26 + |      ref:
     27 + |        branch: main
     28 + |      url: https://github.com/spring-projects/spring-petclinic.git
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "accept yaml file through stdin - using --dry-run flag",
			Args: []string{flags.FilePathFlagName, "-", flags.DryRunFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			Stdin: []byte(`
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  name: my-workload
  namespace: default
  resourceVersion: "999"
spec:
  image: ubuntu:bionic
status:
  supplyChainRef: {}
`),
		},
		{
			Name: "fail to accept yaml file - missing --yes flag",
			Args: []string{flags.FilePathFlagName, "-"},
			Stdin: []byte(`
apiVersion: carto.run/v1alpha2
kind: Workload
metadata:
  name: spring-petclinic
  labels:
    app.kubernetes.io/part-of: spring-petclinic
    apps.tanzu.vmware.com/workload-type: web
spec:
  env:
  - name: SPRING_PROFILES_ACTIVE
    value: mysql
  resources:
    requests:
      memory: 1Gi
      cpu: 100m
    limits:
      memory: 1Gi
      cpu: 500m
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main

`),
			ShouldError: true,
		},
		{
			Name: "filepath - custom namespace and name",
			Args: []string{workloadName, flags.NamespaceFlagName, "test-namespace", flags.FilePathFlagName, "testdata/workload.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Namespace("test-namespace")
						d.Name(workloadName)
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(corev1.EnvVar{
							Name:  "OVERRIDE_VAR",
							Value: "doesnt matter",
						})
					},
					)},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      workloadName,
						Labels: map[string]string{
							"preserve-me":                         "should-exist",
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
							{
								Name:  "SPRING_PROFILES_ACTIVE",
								Value: "mysql",
							},
						},
						Resources: &corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
  6,  8   |    preserve-me: should-exist
  7,  9   |  name: my-workload
  8, 10   |  namespace: test-namespace
  9, 11   |spec:
 10, 12   |  env:
 11, 13   |  - name: OVERRIDE_VAR
 12, 14   |    value: doesnt matter
 13     - |  image: ubuntu:bionic
     15 + |  - name: SPRING_PROFILES_ACTIVE
     16 + |    value: mysql
     17 + |  resources:
     18 + |    limits:
     19 + |      cpu: 500m
     20 + |      memory: 1Gi
     21 + |    requests:
     22 + |      cpu: 100m
     23 + |      memory: 1Gi
     24 + |  source:
     25 + |    git:
     26 + |      ref:
     27 + |        branch: main
     28 + |      url: https://github.com/spring-projects/spring-petclinic.git
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --namespace test-namespace --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload --namespace test-namespace"

`,
		},
		{
			Name:        "filepath - missing",
			Args:        []string{workloadName, flags.FilePathFlagName, "testdata/missing.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name: "update existing param-yaml",
			Args: []string{workloadName, flags.ParamYamlFlagName, `ports_json={"name": "smtp", "port": 2026}`, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(
							&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/spring-projects/spring-petclinic.git",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							},
						).Params(
							cartov1alpha1.Param{
								Name:  "ports_json",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"name":"smtp","port":1026}`)},
							}, cartov1alpha1.Param{
								Name:  "ports_nesting_yaml",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"deployment":{"name":"smtp","port":1026}}]`)},
							},
						)
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
						Params: []cartov1alpha1.Param{
							{
								Name:  "ports_json",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"name":"smtp","port":2026}`)},
							}, {
								Name:  "ports_nesting_yaml",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"deployment":{"name":"smtp","port":1026}}]`)},
							},
						},
					},
				},
			},
		},
		{
			ShouldError: true,
			Name:        "fails create with multiple param-yaml using invalid json",
			Args: []string{flags.FilePathFlagName, "testdata/param-yaml.yaml",
				flags.ParamYamlFlagName, `ports_json={"name": "smtp", "port": 1026`,
				flags.ParamYamlFlagName, `ports_nesting_yaml=- deployment:\n    name: smtp\n    port: 1026`,
				flags.YesFlagName},
			ExpectOutput: "",
		},
		{
			Name: "update existing param-yaml from file",
			Args: []string{workloadName, flags.FilePathFlagName, "testdata/workload-param-yaml.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(
							&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/spring-projects/spring-petclinic.git",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							},
						).Params(
							cartov1alpha1.Param{
								Name:  "ports",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"name":"smtp","port":1026}`)},
							}, cartov1alpha1.Param{
								Name:  "ports_nesting_yaml",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"deployment":{"name":"smtp","port":1026}}]`)},
							},
						)
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.AppPartOfLabelName:               "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
						Params: []cartov1alpha1.Param{
							{
								Name:  "ports",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"ports":[{"name":"http","port":8080,"protocol":"TCP","targetPort":8080},{"name":"https","port":8443,"protocol":"TCP","targetPort":8443}]}`)},
							}, {
								Name:  "ports_nesting_yaml",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"deployment":{"name":"smtp","port":1026}}]`)},
							}, {
								Name:  "services",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"},{"image":"postgres:9.6","name":"postgres"}]`)},
							},
						},
					},
				},
			},
		},
		{
			Name: "update workload to add maven param",
			Args: []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples"}`, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Params: []cartov1alpha1.Param{
							{
								Name:  "maven",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","version":"2.6.0"}`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

🔎 Update workload:
...
  5,  5   |  labels:
  6,  6   |    apps.tanzu.vmware.com/workload-type: web
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9     - |spec: {}
      9 + |spec:
     10 + |  params:
     11 + |  - name: maven
     12 + |    value:
     13 + |      artifactId: spring-petclinic
     14 + |      groupId: org.springframework.samples
     15 + |      version: 2.6.0
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:   "update workload with no color",
			Args:   []string{workloadName, flags.DebugFlagName, flags.YesFlagName},
			Config: &cli.Config{NoColor: true, Scheme: scheme},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "ubuntu:bionic",
						Params: []cartov1alpha1.Param{
							{
								Name:  "debug",
								Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

Update workload:
...
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  image: ubuntu:bionic
     11 + |  params:
     12 + |  - name: debug
     13 + |    value: "true"
Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update from local source using lsp",
			Skip: runtm.GOOS == "windows",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`)),
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
						Annotations: map[string]string{
							"local-source-proxy.apps.tanzu.vmware.com": ":default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &v1alpha1.Source{
							Image: ":default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652",
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

Publishing source in "%s" to "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload"...
📥 Published source

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  annotations:
  6     - |    local-source-proxy.apps.tanzu.vmware.com: my-old-image
      6 + |    local-source-proxy.apps.tanzu.vmware.com: :default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652
  7,  7   |  labels:
  8,  8   |    apps.tanzu.vmware.com/workload-type: web
  9,  9   |  name: my-workload
 10, 10   |  namespace: default
 11     - |spec: {}
     11 + |spec:
     12 + |  source:
     13 + |    image: :default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, localSource),
		},
		{
			Name: "update from local source using lsp with status 204",
			Skip: runtm.GOOS == "windows",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "204"}`)),
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
						Annotations: map[string]string{
							"local-source-proxy.apps.tanzu.vmware.com": ":default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &v1alpha1.Source{
							Image: ":default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652",
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
❗ WARNING: the update command has been deprecated and will be removed in a future update. Please use "tanzu apps workload apply" instead.

Publishing source in "%s" to "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload"...
📥 Published source

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  annotations:
  6     - |    local-source-proxy.apps.tanzu.vmware.com: my-old-image
      6 + |    local-source-proxy.apps.tanzu.vmware.com: :default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652
  7,  7   |  labels:
  8,  8   |    apps.tanzu.vmware.com/workload-type: web
  9,  9   |  name: my-workload
 10, 10   |  namespace: default
 11     - |spec: {}
     11 + |spec:
     12 + |  source:
     13 + |    image: :default-my-workload@sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, localSource),
		},
		{
			Name: "update from local source using lsp with redirect registry error",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "302", "message": "Registry moved found"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nError: Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Registry moved found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with no upstream auth",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "401", "message": "401 Status user UNAUTHORIZED for registry"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nError: Local source proxy was unable to authenticate against the target registry.\nErrors:\n- 401 Status user UNAUTHORIZED for registry"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with not found registry error",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "404", "message": "Registry not found"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nError: Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Registry not found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with internal error server registry error",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "500", "message": "Registry internal error"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nError: Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Registry internal error"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with redirect response",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusFound, `{"message": "302 Status Found"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Either Local Source Proxy is not installed on the Cluster or you don't have permissions to access it\nError: Local source proxy was moved and is not reachable in the defined url.\nErrors:\n- 302 Status Found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with no user permission",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusUnauthorized, `{"message": "401 Status Unauthorized"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Either Local Source Proxy is not installed on the Cluster or you don't have permissions to access it\nError: The current user does not have permission to access the local source proxy.\nErrors:\n- 401 Status Unauthorized"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with no found error",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusNotFound, `{"message": "404 Status Not Found"}`)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy is not installed or the deployment is not healthy. Either install it or use --source-image flag\nError: Local source proxy is not installed on the cluster.\nErrors:\n- 404 Status Not Found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name: "update from local source using lsp with transport error",
			Args: []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Annotations(map[string]string{apis.LocalSourceProxyAnnotationName: "my-old-image"})
					}),
			},
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "client transport not provided"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
	}

	table.Run(t, scheme, func(ctx context.Context, c *cli.Config) *cobra.Command {
		// capture the cobra command so we can make assertions on cleanup, this will fail if tests are run parallel.
		cmd = commands.NewWorkloadUpdateCommand(ctx, c)
		return cmd
	})
}
