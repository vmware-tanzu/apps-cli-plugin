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
	"testing"
	"time"

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
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
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
	defaultNamespace := "default"
	workloadName := "my-workload"
	gitRepo := "https://example.com/repo.git"
	gitBranch := "main"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	var cmd *cobra.Command

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:        "missing source",
			Args:        []string{workloadName},
			ShouldError: true,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				if expected, actual := false, cmd.SilenceUsage; expected != actual {
					t.Errorf("expected cmd.SilenceUsage to be %t, actually %t", expected, actual)
				}
				return nil
			},
		},
		{
			Name: "dry run",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.DryRunFlagName, flags.YesFlagName},
			ExpectOutput: `
---
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
					},
					Status: cartov1alpha1.WorkloadStatus{
						OwnerStatus: cartov1alpha1.OwnerStatus{
							Conditions: []metav1.Condition{
								{
									Type:    cartov1alpha1.WorkloadConditionReady,
									Status:  metav1.ConditionFalse,
									Reason:  "OopsieDoodle",
									Message: "a hopefully informative message about what went wrong",
								},
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
						Labels:    map[string]string{},
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
Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  source:
      9 + |    git:
     10 + |      ref:
     11 + |        branch: main
     12 + |      url: https://example.com/repo.git

Created workload "my-workload"
Waiting for workload "my-workload" to become ready...
Error: Failed to become ready: a hopefully informative message about what went wrong
`,
		},
		{
			Name: "wait with timeout error",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.WaitFlagName, flags.WaitTimeoutFlagName, "1ns"},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
					},
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
						Labels:    map[string]string{},
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
Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  source:
      9 + |    git:
     10 + |      ref:
     11 + |        branch: main
     12 + |      url: https://example.com/repo.git

Created workload "my-workload"
Waiting for workload "my-workload" to become ready...
Error: timeout after 1ns waiting for "my-workload" to become ready
To view status run: tanzu apps workload get my-workload --namespace default
`,
		},
		{
			Name: "successfull wait for ready cond",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.WaitFlagName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
					},
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
						Labels:    map[string]string{},
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
Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  source:
      9 + |    git:
     10 + |      ref:
     11 + |        branch: main
     12 + |      url: https://example.com/repo.git

Created workload "my-workload"
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
					},
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
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)

				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Second, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)

				return ctx, nil
			},
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
						Labels:    map[string]string{},
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
Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  source:
      9 + |    git:
     10 + |      ref:
     11 + |        branch: main
     12 + |      url: https://example.com/repo.git

Created workload "my-workload"
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
					},
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
				}
				fakeWatcher := watchfakes.NewFakeWithWatch(false, config.Client, []watch.Event{
					{Type: watch.Modified, Object: workload},
				})
				ctx = watchhelper.WithWatcher(ctx, fakeWatcher)

				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Second, true).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)

				return ctx, nil
			},
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
						Labels:    map[string]string{},
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
Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  source:
      9 + |    git:
     10 + |      ref:
     11 + |        branch: main
     12 + |      url: https://example.com/repo.git

Created workload "my-workload"
Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready
`,
		},
		{
			Name: "error existing workload",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			GivenObjects: []clitesting.Factory{
				clitesting.Wrapper(&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
					},
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
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
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
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
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
Create workload:
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

Created workload "spring-petclinic"
`,
		},
		{
			Name: "accept yaml file through stdin - using --dry-run flag",
			Args: []string{flags.FilePathFlagName, "-", flags.DryRunFlagName},
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
			Name:        "local path - missing fields",
			Args:        []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.LocalPathFlagName, "testdata/local-source", flags.YesFlagName},
			ShouldError: true,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				if expected, actual := false, cmd.SilenceUsage; expected != actual {
					t.Errorf("expected cmd.SilenceUsage to be %t, actually %t", expected, actual)
				}

				return nil
			},
		},
		{
			Name: "add annotation",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName, flags.AnnotationFlagName, "NEW=value", flags.AnnotationFlagName, "FOO=bar", flags.AnnotationFlagName, "removeme-"},
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "my-workload",
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
			}},
	}

	table.Run(t, scheme, func(ctx context.Context, c *cli.Config) *cobra.Command {
		// capture the cobra command so we can make assertions on cleanup, this will fail if tests are run parallel.
		cmd = commands.NewWorkloadCreateCommand(ctx, c)
		return cmd
	})
}
