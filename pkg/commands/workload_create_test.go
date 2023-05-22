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
	"os"
	"path/filepath"
	runtm "runtime"
	"strings"
	"testing"
	"time"

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	"github.com/Netflix/go-expect"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
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
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestWorkloadCreateOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name: "valid options",
			Validatable: &commands.WorkloadCreateOptions{
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
			Validatable: &commands.WorkloadCreateOptions{
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
			Validatable: &commands.WorkloadCreateOptions{
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

func TestWorkloadCreateCommand(t *testing.T) {
	localSource := filepath.Join("testdata", "local-source")
	defaultNamespace := "default"
	workloadName := "my-workload"
	gitRepo := "https://example.com/repo.git"
	gitBranch := "main"
	serviceAccountName := "my-service-account"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	var cmd *cobra.Command

	givenNamespaceDefault := []client.Object{
		diecorev1.NamespaceBlank.
			MetadataDie(func(d *diemetav1.ObjectMetaDie) {
				d.Name(defaultNamespace)
			}),
	}

	myWorkloadHeader := http.Header{
		"Content-Type":          []string{"text/html", "application/json", "application/octet-stream"},
		"Docker-Content-Digest": []string{"sha256:111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"},
	}
	respCreator := func(status int, body string, headers map[string][]string) *http.Response {
		return &http.Response{
			Status:     http.StatusText(status),
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     headers,
		}
	}

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:         "no source",
			Args:         []string{workloadName},
			GivenObjects: givenNamespaceDefault,
		},
		{
			Name:         "dry run",
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.DryRunFlagName, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectOutput: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: my-workload
  namespace: default
spec:
  source:
    git:
      ref:
        branch: main
      url: https://example.com/repo.git
status:
  supplyChainRef: {}
`,
		},
		{
			Name: "create - output yaml with wait",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch,
				flags.OutputFlagName, printer.OutputFormatYaml, flags.WaitFlagName, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
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
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: my-workload
  namespace: default
  resourceVersion: "1"
spec:
  source:
    git:
      ref:
        branch: main
      url: https://example.com/repo.git
status:
  supplyChainRef: {}
`,
		},
		{
			Name: "create - output json",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch,
				flags.OutputFlagName, printer.OutputFormatJson, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"creationTimestamp": null,
		"labels": {
			"apps.tanzu.vmware.com/workload-type": "web"
		},
		"name": "my-workload",
		"namespace": "default",
		"resourceVersion": "1"
	},
	"spec": {
		"source": {
			"git": {
				"ref": {
					"branch": "main"
				},
				"url": "https://example.com/repo.git"
			}
		}
	},
	"status": {
		"supplyChainRef": {}
	}
}
`,
		},
		{
			Name: "wait error for false condition",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.WaitFlagName},
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
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ShouldError: true,
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error waiting for ready condition: Failed to become ready: a hopefully informative message about what went wrong
`,
		},
		{
			Name: "wait with timeout error",
			Skip: runtm.GOOS == "windows",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.WaitFlagName, flags.WaitTimeoutFlagName, "1ns"},
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
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ShouldError: true,
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error waiting for ready condition: timeout after 1ns waiting for "my-workload" to become ready
`,
		},
		{
			Name: "successful wait for ready cond",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.WaitFlagName},
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
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Workload "my-workload" is ready

`,
		},
		{
			Name: "tail while waiting for ready cond",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.TailFlagName},
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
			GivenObjects: givenNamespaceDefault,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready

`,
		},
		{
			Name: "tail with timestamp while waiting for ready cond",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.TailTimestampFlagName},
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
			GivenObjects: givenNamespaceDefault,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready

`,
		},
		{
			Name: "error existing workload",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			GivenObjects: []client.Object{
				diecartov1alpha1.WorkloadBlank.MetadataDie(func(d *diemetav1.ObjectMetaDie) {
					d.Namespace(defaultNamespace)
					d.Name(workloadName)
				}),
			},
			ExpectOutput: `
Error: workload "default/my-workload" already exists
`,
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("create", "Workload"),
			},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "watcher error",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.WaitFlagName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				fakewatch := watchfakes.NewFakeWithWatch(true, config.Client, []watch.Event{})
				ctx = watchhelper.WithWatcher(ctx, fakewatch)
				return ctx, nil
			},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name:         "accept yaml file through stdin - using --yes flag",
			Args:         []string{flags.FilePathFlagName, "-", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
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
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
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
						Env: []corev1.EnvVar{
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
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
      8 + |  name: spring-petclinic
      9 + |  namespace: default
     10 + |spec:
     11 + |  env:
     12 + |  - name: SPRING_PROFILES_ACTIVE
     13 + |    value: mysql
     14 + |  resources:
     15 + |    limits:
     16 + |      cpu: 500m
     17 + |      memory: 1Gi
     18 + |    requests:
     19 + |      cpu: 100m
     20 + |      memory: 1Gi
     21 + |  source:
     22 + |    git:
     23 + |      ref:
     24 + |        branch: main
     25 + |      url: https://github.com/spring-projects/spring-petclinic.git
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name:         "accept yaml file through stdin - using --dry-run flag",
			Args:         []string{flags.FilePathFlagName, "-", flags.DryRunFlagName},
			GivenObjects: givenNamespaceDefault,
			Stdin: []byte(`
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  name: my-workload
  namespace: default
spec:
  source:
    git:
      ref:
        branch: main
      url: https://example.com/repo.git
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
			Name:        "filepath - missing",
			Args:        []string{workloadName, flags.FilePathFlagName, "testdata/missing.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name:        "filepath invalid name",
			Args:        []string{flags.FilePathFlagName, "testdata/workload-invalid-name.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name:         "add annotation",
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.AnnotationFlagName, "NEW=value", flags.AnnotationFlagName, "FOO=bar", flags.AnnotationFlagName, "removeme-"},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "my-workload",
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Params: []cartov1alpha1.Param{
							{
								Name:  "annotations",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"FOO":"bar","NEW":"value"}`)},
							},
						},
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://example.com/repo.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:         "create with serviceAccountName specifying other flags from cli",
			Args:         []string{flags.FilePathFlagName, "testdata/service-account-name.yaml", flags.GitTagFlagName, "tap-1.2", flags.TypeFlagName, "whatever", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							apis.AppPartOfLabelName:               "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "whatever",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.2",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: whatever
      8 + |  name: spring-petclinic
      9 + |  namespace: default
     10 + |spec:
     11 + |  serviceAccountName: my-service-account
     12 + |  source:
     13 + |    git:
     14 + |      ref:
     15 + |        tag: tap-1.2
     16 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name:         "create from maven artifact using paramyaml",
			Args:         []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples"}`, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  params:
     11 + |  - name: maven
     12 + |    value:
     13 + |      artifactId: spring-petclinic
     14 + |      groupId: org.springframework.samples
     15 + |      version: 2.6.0
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "create from maven artifact using flags",
			Args:         []string{workloadName, flags.MavenArtifactFlagName, "spring-petclinic", flags.MavenVersionFlagName, "2.6.0", flags.MavenGroupFlagName, "org.springframework.samples", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  params:
     11 + |  - name: maven
     12 + |    value:
     13 + |      artifactId: spring-petclinic
     14 + |      groupId: org.springframework.samples
     15 + |      version: 2.6.0
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "create from maven artifact using paramyaml with type",
			Args:         []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples", "type": "jar"}`, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
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
								Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","type":"jar","version":"2.6.0"}`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  params:
     11 + |  - name: maven
     12 + |    value:
     13 + |      artifactId: spring-petclinic
     14 + |      groupId: org.springframework.samples
     15 + |      type: jar
     16 + |      version: 2.6.0
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:        "error missing maven flags",
			Args:        []string{workloadName, flags.MavenArtifactFlagName, "spring-petclinic", flags.MavenVersionFlagName, "1.2.3", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name: "create with multiple param-yaml using valid json and yaml",
			Args: []string{flags.FilePathFlagName, "testdata/param-yaml.yaml",
				flags.ParamYamlFlagName, `ports_json={"name": "smtp", "port": 1026}`,
				flags.ParamYamlFlagName, "ports_nesting_yaml=- deployment:\n    name: smtp\n    port: 1026",
				flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							apis.AppPartOfLabelName:               "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Image: "my-reponame/my-image:my-tag",
						Params: []cartov1alpha1.Param{
							{
								Name:  "ports_json",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"name":"smtp","port":1026}`)},
							}, {
								Name:  "ports_nesting_yaml",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"deployment":{"name":"smtp","port":1026}}]`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
      8 + |  name: spring-petclinic
      9 + |  namespace: default
     10 + |spec:
     11 + |  image: my-reponame/my-image:my-tag
     12 + |  params:
     13 + |  - name: ports_json
     14 + |    value:
     15 + |      name: smtp
     16 + |      port: 1026
     17 + |  - name: ports_nesting_yaml
     18 + |    value:
     19 + |    - deployment:
     20 + |        name: smtp
     21 + |        port: 1026
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
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
			Name: "create with multiple param-yaml using valid json and yaml from file",
			Args: []string{flags.FilePathFlagName, "testdata/workload-param-yaml.yaml",
				flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
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
								Name:  "services",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"},{"image":"postgres:9.6","name":"postgres"}]`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
      8 + |  name: spring-petclinic
      9 + |  namespace: default
     10 + |spec:
     11 + |  params:
     12 + |  - name: ports
     13 + |    value:
     14 + |      ports:
     15 + |      - name: http
     16 + |        port: 8080
     17 + |        protocol: TCP
     18 + |        targetPort: 8080
     19 + |      - name: https
     20 + |        port: 8443
     21 + |        protocol: TCP
     22 + |        targetPort: 8443
     23 + |  - name: services
     24 + |    value:
     25 + |    - image: mysql:5.7
     26 + |      name: mysql
     27 + |    - image: postgres:9.6
     28 + |      name: postgres
     29 + |  source:
     30 + |    git:
     31 + |      ref:
     32 + |        branch: main
     33 + |      url: https://github.com/spring-projects/spring-petclinic.git
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		}, {
			Name: "git source with non-allowed env var",
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				os.Setenv("TANZU_APPS_LABEL", "foo=var")
				return ctx, nil
			},
			GivenObjects: givenNamespaceDefault,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_LABEL")
				return nil
			},
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			ExpectCreates: []client.Object{
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
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		}, {
			Name: "git source with allowed and non-allowed env var",
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				os.Setenv("TANZU_APPS_LABEL", "foo=var")
				os.Setenv("TANZU_APPS_TYPE", "web")
				return ctx, nil
			},
			GivenObjects: givenNamespaceDefault,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_LABEL")
				os.Unsetenv("TANZU_APPS_TYPE")
				return nil
			},
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		}, {
			Name: "git source with allowed env var overwritten",
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				os.Setenv("TANZU_APPS_TYPE", "jar")
				return ctx, nil
			},
			GivenObjects: givenNamespaceDefault,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_TYPE")
				return nil
			},
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.TypeFlagName, "web", flags.YesFlagName},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		}, {
			Name:         "git source with terminal interaction",
			GivenObjects: givenNamespaceDefault,
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.TypeFlagName, "web"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`👍 Created workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
			},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

👍 Created workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: y"), workloadName),
		}, {
			Name:         "git source with terminal interaction reject",
			GivenObjects: givenNamespaceDefault,
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.TypeFlagName, "web"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("n"))
				c.ExpectString(clitesting.ToInteractOutput("Skipping workload %q", workloadName))
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

Skipping workload %q`, clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: n"), workloadName),
		}, {
			Name:         "git source with wrong answer terminal interaction reject",
			GivenObjects: givenNamespaceDefault,
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.TypeFlagName, "web"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("m"))
				c.ExpectString(clitesting.ToInteractTerminal("invalid input (not y, n, yes, or no)"))
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("n"))
				c.ExpectString(clitesting.ToInteractOutput("Skipping workload %q", workloadName))
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

invalid input (not y, n, yes, or no)
%s

Skipping workload %q`,
				clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: m"),
				clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: n"), workloadName),
		},

		{
			Name:         "output workload after create in yaml format",
			GivenObjects: givenNamespaceDefault,
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch,
				flags.TypeFlagName, "web", flags.OutputFlagName, printer.OutputFormatYaml},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`👍 Created workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
			},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

👍 Created workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: null
  labels:
    apps.tanzu.vmware.com/workload-type: web
  name: my-workload
  namespace: default
  resourceVersion: "1"
spec:
  source:
    git:
      ref:
        branch: main
      url: https://example.com/repo.git
status:
  supplyChainRef: {}
`, clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: y"), workloadName),
		},
		{
			Name:         "output workload after create in json format",
			GivenObjects: givenNamespaceDefault,
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch,
				flags.TypeFlagName, "web", flags.OutputFlagName, printer.OutputFormatJson},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`👍 Created workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
			},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

👍 Created workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"creationTimestamp": null,
		"labels": {
			"apps.tanzu.vmware.com/workload-type": "web"
		},
		"name": "my-workload",
		"namespace": "default",
		"resourceVersion": "1"
	},
	"spec": {
		"source": {
			"git": {
				"ref": {
					"branch": "main"
				},
				"url": "https://example.com/repo.git"
			}
		}
	},
	"status": {
		"supplyChainRef": {}
	}
}
`, clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: y"), workloadName),
		},
		{
			Name:         "output workload after create in json format with tail error",
			GivenObjects: givenNamespaceDefault,
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch,
				flags.TypeFlagName, "web", flags.OutputFlagName, printer.OutputFormatJson, flags.TailFlagName},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`👍 Created workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
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
				fakeWatcher := watchfakes.NewFakeWithWatch(true, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)

				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Minute, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)

				return ctx, nil
			},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

👍 Created workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Error waiting for ready condition: failed to create watcher
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"creationTimestamp": null,
		"labels": {
			"apps.tanzu.vmware.com/workload-type": "web"
		},
		"name": "my-workload",
		"namespace": "default",
		"resourceVersion": "1"
	},
	"spec": {
		"source": {
			"git": {
				"ref": {
					"branch": "main"
				},
				"url": "https://example.com/repo.git"
			}
		}
	},
	"status": {
		"supplyChainRef": {}
	}
}
`, clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: y"), workloadName),
		},
		{
			Name:         "output workload after create in json format with wait error",
			GivenObjects: givenNamespaceDefault,
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch,
				flags.TypeFlagName, "web", flags.OutputFlagName, printer.OutputFormatJson, flags.WaitFlagName},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`👍 Created workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
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
				fakeWatcher := watchfakes.NewFakeWithWatch(true, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)

				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Minute, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)

				return ctx, nil
			},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: gitRepo,
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  source:
     11 + |    git:
     12 + |      ref:
     13 + |        branch: main
     14 + |      url: https://example.com/repo.git
%s

👍 Created workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error waiting for ready condition: failed to create watcher
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"creationTimestamp": null,
		"labels": {
			"apps.tanzu.vmware.com/workload-type": "web"
		},
		"name": "my-workload",
		"namespace": "default",
		"resourceVersion": "1"
	},
	"spec": {
		"source": {
			"git": {
				"ref": {
					"branch": "main"
				},
				"url": "https://example.com/repo.git"
			}
		}
	},
	"status": {
		"supplyChainRef": {}
	}
}
`, clitesting.ToInteractTerminal("❓ Do you want to create this workload? [yN]: y"), workloadName),
		},
		{
			Name:                "create from local source using lsp",
			Skip:                runtm.GOOS == "windows",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`, myWorkloadHeader)),
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
						Annotations: map[string]string{
							"local-source-proxy.apps.tanzu.vmware.com": ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Image: ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
Publishing source in "%s" to "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload"...
📥 Published source

🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  annotations:
      6 + |    local-source-proxy.apps.tanzu.vmware.com: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
      7 + |  labels:
      8 + |    apps.tanzu.vmware.com/workload-type: web
      9 + |  name: my-workload
     10 + |  namespace: default
     11 + |spec:
     12 + |  source:
     13 + |    image: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, localSource),
		},

		{
			Name:                "create using lsp with subpath",
			Skip:                runtm.GOOS == "windows",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.SubPathFlagName, subpath, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`, myWorkloadHeader)),
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
						Annotations: map[string]string{
							"local-source-proxy.apps.tanzu.vmware.com": ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Image:   ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
							Subpath: "testdata/local-source/subpath",
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
Publishing source in "%s" to "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload"...
📥 Published source

🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  annotations:
      6 + |    local-source-proxy.apps.tanzu.vmware.com: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
      7 + |  labels:
      8 + |    apps.tanzu.vmware.com/workload-type: web
      9 + |  name: my-workload
     10 + |  namespace: default
     11 + |spec:
     12 + |  source:
     13 + |    image: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
     14 + |    subPath: testdata/local-source/subpath
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, localSource),
		},
		{
			Name:                "create using lsp and taking fields from file",
			Skip:                runtm.GOOS == "windows",
			GivenObjects:        givenNamespaceDefault,
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.FilePathFlagName, "testdata/workload.yaml", flags.YesFlagName},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`, myWorkloadHeader)),
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "my-workload",
						Labels: map[string]string{
							apis.AppPartOfLabelName:               "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
						Annotations: map[string]string{
							"local-source-proxy.apps.tanzu.vmware.com": ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Image: ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
						Env: []corev1.EnvVar{
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
			ExpectOutput: fmt.Sprintf(`
Publishing source in "%s" to "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload"...
📥 Published source

🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  annotations:
      6 + |    local-source-proxy.apps.tanzu.vmware.com: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
      7 + |  labels:
      8 + |    app.kubernetes.io/part-of: spring-petclinic
      9 + |    apps.tanzu.vmware.com/workload-type: web
     10 + |  name: my-workload
     11 + |  namespace: default
     12 + |spec:
     13 + |  env:
     14 + |  - name: SPRING_PROFILES_ACTIVE
     15 + |    value: mysql
     16 + |  resources:
     17 + |    limits:
     18 + |      cpu: 500m
     19 + |      memory: 1Gi
     20 + |    requests:
     21 + |      cpu: 100m
     22 + |      memory: 1Gi
     23 + |  source:
     24 + |    image: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, localSource),
		},
		{
			Name:                "create from local source using lsp with 204 response",
			Skip:                runtm.GOOS == "windows",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "204"}`, myWorkloadHeader)),
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							apis.WorkloadTypeLabelName: "web",
						},
						Annotations: map[string]string{
							"local-source-proxy.apps.tanzu.vmware.com": ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Image: ":default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69",
						},
					},
				},
			},
			ExpectOutput: fmt.Sprintf(`
Publishing source in "%s" to "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload"...
📥 Published source

🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  annotations:
      6 + |    local-source-proxy.apps.tanzu.vmware.com: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
      7 + |  labels:
      8 + |    apps.tanzu.vmware.com/workload-type: web
      9 + |  name: my-workload
     10 + |  namespace: default
     11 + |spec:
     12 + |  source:
     13 + |    image: :default-my-workload@sha256:978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, localSource),
		},
		{
			Name:                "create from local source using lsp with redirect registry error",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "302", "message": "Registry moved found"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nReason: Local source proxy was unable to authenticate against the target registry.\nMessages:\n- Registry moved found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:                "create from local source using lsp with no upstream auth",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "401", "message": "401 Status user UNAUTHORIZED for registry"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nReason: Local source proxy was unable to authenticate against the target registry.\nMessages:\n- 401 Status user UNAUTHORIZED for registry"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:                "create from local source using lsp with not found registry error",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "404", "message": "Registry not found"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nReason: Local source proxy was unable to authenticate against the target registry.\nMessages:\n- Registry not found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:                "create from local source using lsp with internal error server registry error",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "500", "message": "Registry internal error"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy failed to upload source to the repository\nReason: Local source proxy was unable to authenticate against the target registry.\nMessages:\n- Registry internal error"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:                "create from local source using lsp with redirect response",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusFound, `{"message": "302 Status Found"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Either Local Source Proxy is not installed on the Cluster or you don't have permissions to access it\nReason: Local source proxy was moved and is not reachable in the defined url.\nMessages:\n- 302 Status Found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:                "create from local source using lsp with no user permission",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusUnauthorized, `{"message": "401 Status Unauthorized"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Either Local Source Proxy is not installed on the Cluster or you don't have permissions to access it\nReason: The current user does not have permission to access the local source proxy.\nMessages:\n- 401 Status Unauthorized"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:                "create from local source using lsp with no found error",
			Args:                []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects:        givenNamespaceDefault,
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusNotFound, `{"message": "404 Status Not Found"}`, myWorkloadHeader)),
			ShouldError:         true,
			Verify: func(t *testing.T, output string, err error) {
				msg := "Local source proxy is not installed or the deployment is not healthy. Either install it or use --source-image flag\nReason: Local source proxy is not installed on the cluster.\nMessages:\n- 404 Status Not Found"
				if err.Error() != msg {
					t.Errorf("Expected error to be %q but got %q", msg, err.Error())
				}
			},
		},
		{
			Name:         "create from local source using lsp with transport error",
			Args:         []string{workloadName, flags.LocalPathFlagName, localSource, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ShouldError:  true,
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
		cmd = commands.NewWorkloadCreateCommand(ctx, c)
		return cmd
	})
}
