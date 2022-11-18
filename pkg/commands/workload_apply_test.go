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
	"os"
	runtm "runtime"
	"testing"
	"time"

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	"github.com/Netflix/go-expect"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
)

func TestWorkloadApplyOptionsValidate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	table := clitesting.ValidatableTestSuite{
		{
			Name: "valid options",
			Validatable: &commands.WorkloadApplyOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					Env:       []string{"FOO=bar"},
				},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid options",
			Validatable: &commands.WorkloadApplyOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					Env:       []string{"FOO"},
				},
			},
			ExpectFieldErrors: validation.ErrInvalidArrayValue("FOO", flags.EnvFlagName, 0),
		},
		{
			Name: "update strategy without filepath",
			Validatable: &commands.WorkloadApplyOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
				},
				UpdateStrategy: "replace",
			},
			Prepare: func(t *testing.T, ctx context.Context) (context.Context, error) {
				cmd := commands.NewWorkloadApplyCommand(ctx, cli.NewDefaultConfig("test", scheme))
				if err := cmd.Flags().Set(cli.StripDash(flags.UpdateStrategyFlagName), "replace"); err != nil {
					return ctx, err
				}
				ctx = cli.WithCommand(ctx, cmd)
				return ctx, nil
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.FilePathFlagName),
		},
		{
			Name: "filepath with update strategy",
			Validatable: &commands.WorkloadApplyOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					FilePath:  "my-folder/my-filepath.yaml",
				},
				UpdateStrategy: "merge",
			},
			Prepare: func(t *testing.T, ctx context.Context) (context.Context, error) {
				cmd := commands.NewWorkloadApplyCommand(ctx, cli.NewDefaultConfig("test", scheme))
				if err := cmd.Flags().Set(cli.StripDash(flags.UpdateStrategyFlagName), "merge"); err != nil {
					return ctx, err
				}
				ctx = cli.WithCommand(ctx, cmd)
				return ctx, nil
			},
			ShouldValidate: true,
		},
		{
			Name: "filepath with invalid value for update strategy",
			Validatable: &commands.WorkloadApplyOptions{
				WorkloadOptions: commands.WorkloadOptions{
					Namespace: "default",
					Name:      "my-resource",
					FilePath:  "my-folder/my-filepath.yaml",
				},
				UpdateStrategy: "invalid",
			},
			Prepare: func(t *testing.T, ctx context.Context) (context.Context, error) {
				cmd := commands.NewWorkloadApplyCommand(ctx, cli.NewDefaultConfig("test", scheme))
				if err := cmd.Flags().Set(cli.StripDash(flags.UpdateStrategyFlagName), "invalid"); err != nil {
					return ctx, err
				}
				ctx = cli.WithCommand(ctx, cmd)
				return ctx, nil
			},
			ExpectFieldErrors: validation.EnumInvalidValue("invalid", flags.UpdateStrategyFlagName, []string{"merge", "replace"}),
		},
	}

	table.Run(t)
}

func TestWorkloadApplyCommand(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	file := "testdata/workload.yaml"
	gitRepo := "https://example.com/repo.git"
	gitBranch := "main"
	serviceAccountName := "my-service-account"
	serviceAccountNameUpdated := "my-service-account-updated"
	fileFromUrl := "https://raw.githubusercontent.com/vmware-tanzu/apps-cli-plugin/main/pkg/commands/testdata/workload.yaml"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	var cmd *cobra.Command

	parent := diecartov1alpha1.WorkloadBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(workloadName)
			d.Namespace(defaultNamespace)
		})

	givenNamespaceDefault := []client.Object{
		diecorev1.NamespaceBlank.
			MetadataDie(func(d *diemetav1.ObjectMetaDie) {
				d.Name(defaultNamespace)
			}),
	}

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "get failed",
			Args: []string{flags.FilePathFlagName, file},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Workload"),
			},
			ShouldError: true,
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
			Name:         "git source with subPath",
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.SubPathFlagName, "./app", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
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
							Subpath: "./app",
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
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  source:
      9 + |    git:
     10 + |      ref:
     11 + |        branch: main
     12 + |      url: https://example.com/repo.git
     13 + |    subPath: ./app
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "create git source with invalid namespace",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.NamespaceFlagName, "foo", flags.YesFlagName},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Namespace", clitesting.InduceFailureOpts{
					Error: apierrs.NewNotFound(corev1.Resource("Namespace"), "foo"),
				}),
			},
			ShouldError: true,
			ExpectOutput: `
Error: namespace "foo" not found, it may not exist or user does not have permissions to read it.
`,
		},
		{
			Name: "Update git source with subPath from file",
			Args: []string{workloadName, flags.FilePathFlagName, "./testdata/workload-subPath.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
							Subpath: "./app",
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  9,  9   |    git:
 10, 10   |      ref:
 11, 11   |        branch: main
 12, 12   |      url: https://github.com/spring-projects/spring-petclinic.git
     13 + |    subPath: ./app
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "Create git source with subPath from file",
			Args:         []string{workloadName, flags.FilePathFlagName, "./testdata/workload-subPath.yaml", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: gitBranch,
								},
							},
							Subpath: "./app",
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Create workload:
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
     12 + |      url: https://github.com/spring-projects/spring-petclinic.git
     13 + |    subPath: ./app
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:        "subPath with no source",
			Args:        []string{workloadName, flags.SubPathFlagName, "./app", flags.YesFlagName},
			ShouldError: true,
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
🔎 Create workload:
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
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error: timeout after 1ns waiting for "my-workload" to become ready
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
🔎 Create workload:
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
🔎 Create workload:
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
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready
`,
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
			GivenObjects: givenNamespaceDefault,
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
			Name: "create - wait error for false condition",
			Args: []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.LabelFlagName, "apps.tanzu.vmware.com/workload-type=web", flags.LabelFlagName, "apps.tanzu.vmware.com/workload-type-", flags.YesFlagName, flags.WaitFlagName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
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
🔎 Create workload:
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
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error: Failed to become ready: a hopefully informative message about what went wrong
`,
		},
		{
			Name:         "filepath from url",
			Args:         []string{flags.FilePathFlagName, fileFromUrl, flags.YesFlagName},
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			Name:        "invalid url",
			Args:        []string{flags.FilePathFlagName, "https://github.com/vmware-tanzu/apps-cli-plugin/blob/main/pkg/commands/testdata/workload.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name:        "not http/https url",
			Args:        []string{flags.FilePathFlagName, "ftp://raw.githubusercontent.com/vmware-tanzu/apps-cli-plugin/main/pkg/commands/testdata/workload.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name:         "filepath",
			Args:         []string{flags.FilePathFlagName, file, flags.YesFlagName},
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			Name: "create - accept yaml file through stdin - using --yes flag",
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			Name: "update - accept yaml file through stdin - using --yes flag",
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
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.Namespace(defaultNamespace)
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(
							corev1.EnvVar{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
						)
					}),
				givenNamespaceDefault[0],
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			Name: "update - accept yaml file through stdin - using --dry-run flag",
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
			Name:         "create - accept yaml file through stdin - using --dry-run flag",
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
			Name:         "filepath - service account build-env",
			Args:         []string{flags.FilePathFlagName, "testdata/workload-build-env.yaml", flags.YesFlagName},
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
						Build: &cartov1alpha1.WorkloadBuild{Env: []corev1.EnvVar{
							{Name: "BP_MAVEN_POM_FILE", Value: "skip-pom.xml"},
						}},
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
     11 + |  build:
     12 + |    env:
     13 + |    - name: BP_MAVEN_POM_FILE
     14 + |      value: skip-pom.xml
     15 + |  env:
     16 + |  - name: SPRING_PROFILES_ACTIVE
     17 + |    value: mysql
     18 + |  resources:
     19 + |    limits:
     20 + |      cpu: 500m
     21 + |      memory: 1Gi
     22 + |    requests:
     23 + |      cpu: 100m
     24 + |      memory: 1Gi
     25 + |  source:
     26 + |    git:
     27 + |      ref:
     28 + |        branch: main
     29 + |      url: https://github.com/spring-projects/spring-petclinic.git
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name:        "fail to accept yaml file through stdin missing --yes flag",
			Args:        []string{flags.FilePathFlagName, "-"},
			ShouldError: true,
		},
		{
			Name:        "filepath - missing",
			Args:        []string{workloadName, flags.FilePathFlagName, "testdata/missing.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name: "noop",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectOutput: `
Workload is unchanged, skipping update
`,
		},
		{
			Name: "no source resource",
			Args: []string{workloadName},
			GivenObjects: []client.Object{
				parent,
			},
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
			Name: "update - dry run",
			Args: []string{workloadName, flags.DebugFlagName, flags.DryRunFlagName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
					}),
			},
			ExpectOutput: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  creationTimestamp: "1970-01-01T00:00:01Z"
  name: my-workload
  namespace: default
  resourceVersion: "999"
spec:
  image: ubuntu:bionic
  params:
  - name: debug
    value: "true"
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
						Labels:    map[string]string{},
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
			Name: "conflict during update",
			Args: []string{workloadName, flags.DebugFlagName, flags.YesFlagName},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("update", "Workload", clitesting.InduceFailureOpts{
					Error: apierrs.NewConflict(schema.GroupResource{Group: "carto.run", Resource: "workloads"}, workloadName, fmt.Errorf("induced conflict")),
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
						Labels:    map[string]string{},
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
🔎 Update workload:
...
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8,  8   |  image: ubuntu:bionic
      9 + |  params:
     10 + |  - name: debug
     11 + |    value: "true"
Error: conflict updating workload, the object was modified by another user; please run the update command again
`,
		},
		{
			Name: "update - wait error with timeout",
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
🔎 Update workload:
...
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8,  8   |  image: ubuntu:bionic
      9 + |  serviceClaims:
     10 + |  - name: database
     11 + |    ref:
     12 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     13 + |      kind: PostgreSQL
     14 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error: timeout after 1ns waiting for "my-workload" to become ready
`,
		},
		{
			Name: "update - wait error for false condition",
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
🔎 Update workload:
...
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8,  8   |  image: ubuntu:bionic
      9 + |  serviceClaims:
     10 + |  - name: database
     11 + |    ref:
     12 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     13 + |      kind: PostgreSQL
     14 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Error: Failed to become ready: a hopefully informative message about what went wrong
`,
		},
		{
			Name: "update - successful wait for ready condition",
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
🔎 Update workload:
...
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8,  8   |  image: ubuntu:bionic
      9 + |  serviceClaims:
     10 + |  - name: database
     11 + |    ref:
     12 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     13 + |      kind: PostgreSQL
     14 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
Workload "my-workload" is ready
`,
		},
		{
			Name: "update - tail with timestamp while waiting for ready condition",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.YesFlagName, flags.TailTimestampFlagName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				workload := &cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
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
🔎 Update workload:
...
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8,  8   |  image: ubuntu:bionic
      9 + |  serviceClaims:
     10 + |  - name: database
     11 + |    ref:
     12 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     13 + |      kind: PostgreSQL
     14 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

Waiting for workload "my-workload" to become ready...
...tail output...
Workload "my-workload" is ready
`,
		},
		{
			Name: "update - filepath",
			Args: []string{flags.FilePathFlagName, file, flags.SubPathFlagName, "./cmd", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(
							corev1.EnvVar{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
						)
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
							Subpath: "./cmd",
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
     29 + |    subPath: ./cmd
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - filepath from url",
			Args: []string{flags.FilePathFlagName, fileFromUrl, flags.SubPathFlagName, "./cmd", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(
							corev1.EnvVar{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
						)
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
							Subpath: "./cmd",
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
     29 + |    subPath: ./cmd
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - filepath - custom namespace and name",
			Args: []string{workloadName, flags.NamespaceFlagName, "test-namespace", flags.FilePathFlagName, file, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Namespace("test-namespace")
						d.AddLabel("preserve-me", "should-exist")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(
							corev1.EnvVar{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
						)
					}),
			},
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			Name:         "local path - missing fields",
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.LocalPathFlagName, "testdata/local-source", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ShouldError:  true,
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				if expected, actual := false, cmd.SilenceUsage; expected != actual {
					t.Errorf("expected cmd.SilenceUsage to be %t, actually %t", expected, actual)
				}

				return nil
			},
		},
		{
			Name:        "filepath invalid name",
			Args:        []string{flags.FilePathFlagName, "testdata/workload-invalid-name.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name: "update - serviceclaim with deprecation warning",
			Args: []string{workloadName, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-ns:my-prod-db", flags.YesFlagName},
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
						Annotations: map[string]string{
							apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
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
❗ WARNING: Cross namespace service claims are deprecated. Please use ` + "`tanzu service claim create`" + ` instead.
🔎 Update workload:
  1,  1   |---
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
      5 + |  annotations:
      6 + |    serviceclaims.supplychain.apps.x-tanzu.vmware.com/extensions: '{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}'
  5,  7   |  name: my-workload
  6,  8   |  namespace: default
  7,  9   |spec:
  8, 10   |  image: ubuntu:bionic
     11 + |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: PostgreSQL
     16 + |      name: my-prod-db
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "create - serviceclaim with deprecation warning",
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-ns:my-prod-db", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Annotations: map[string]string{
							apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{URL: "https://example.com/repo.git", Ref: cartov1alpha1.GitRef{Branch: "main"}},
						},
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
❗ WARNING: Cross namespace service claims are deprecated. Please use ` + "`tanzu service claim create`" + ` instead.
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  annotations:
      6 + |    serviceclaims.supplychain.apps.x-tanzu.vmware.com/extensions: '{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}'
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  serviceClaims:
     11 + |  - name: database
     12 + |    ref:
     13 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     14 + |      kind: PostgreSQL
     15 + |      name: my-prod-db
     16 + |  source:
     17 + |    git:
     18 + |      ref:
     19 + |        branch: main
     20 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update serviceAccountName via file",
			Args: []string{flags.FilePathFlagName, "testdata/service-account-name.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
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
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
  9     - |spec: {}
     11 + |spec:
     12 + |  serviceAccountName: my-service-account
     13 + |  source:
     14 + |    git:
     15 + |      ref:
     16 + |        tag: tap-1.1
     17 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "delete serviceAccountName by setting to empty in file",
			Args: []string{flags.FilePathFlagName, "testdata/no-service-account-name.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
					}).SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
					d.ServiceAccountName(&serviceAccountName)
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
						ServiceAccountName: nil,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
 10     - |  serviceAccountName: my-service-account
     12 + |  source:
     13 + |    git:
     14 + |      ref:
     15 + |        tag: tap-1.1
     16 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "updated serviceAccountName taking priority from flag",
			Args: []string{flags.FilePathFlagName, "testdata/no-service-account-name.yaml", flags.ServiceAccountFlagName, serviceAccountNameUpdated, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
					}).SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
					d.ServiceAccountName(&serviceAccountName)
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
						ServiceAccountName: &serviceAccountNameUpdated,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
 10     - |  serviceAccountName: my-service-account
     12 + |  serviceAccountName: my-service-account-updated
     13 + |  source:
     14 + |    git:
     15 + |      ref:
     16 + |        tag: tap-1.1
     17 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "delete serviceAccountName field",
			Args: []string{flags.FilePathFlagName, "testdata/no-service-account-name.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
					}).SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
					d.ServiceAccountName(&serviceAccountName)
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
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
 10     - |  serviceAccountName: my-service-account
     12 + |  source:
     13 + |    git:
     14 + |      ref:
     15 + |        tag: tap-1.1
     16 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "do not delete serviceAccountName when updating another field",
			Args: []string{workloadName, flags.GitTagFlagName, "tap-1.1", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(
						func(d *diecartov1alpha1.WorkloadSpecDie) {
							d.ServiceAccountName(&serviceAccountName)
							d.Source(&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/sample-accelerators/spring-petclinic",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							})
						}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
									Tag:    "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  9,  9   |  source:
 10, 10   |    git:
 11, 11   |      ref:
 12, 12   |        branch: main
     13 + |        tag: tap-1.1
 13, 14   |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update serviceAccountName field via flag",
			Args: []string{workloadName, flags.ServiceAccountFlagName, serviceAccountNameUpdated, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(
						func(d *diecartov1alpha1.WorkloadSpecDie) {
							d.ServiceAccountName(&serviceAccountName)
							d.Source(&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/sample-accelerators/spring-petclinic",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							})
						}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountNameUpdated,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  4,  4   |metadata:
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8     - |  serviceAccountName: my-service-account
      8 + |  serviceAccountName: my-service-account-updated
  9,  9   |  source:
 10, 10   |    git:
 11, 11   |      ref:
 12, 12   |        branch: main
...
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "delete serviceAccountName via flag",
			Args: []string{workloadName, flags.ServiceAccountFlagName, "", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(
						func(d *diecartov1alpha1.WorkloadSpecDie) {
							d.ServiceAccountName(&serviceAccountName)
							d.Source(&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/sample-accelerators/spring-petclinic",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							})
						}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: nil,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  4,  4   |metadata:
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7,  7   |spec:
  8     - |  serviceAccountName: my-service-account
  9,  8   |  source:
 10,  9   |    git:
 11, 10   |      ref:
 12, 11   |        branch: main
...
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "create with serviceAccountName",
			Args:         []string{flags.FilePathFlagName, "testdata/service-account-name.yaml", flags.YesFlagName},
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
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
     11 + |  serviceAccountName: my-service-account
     12 + |  source:
     13 + |    git:
     14 + |      ref:
     15 + |        tag: tap-1.1
     16 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name:         "create with serviceAccountName via flag",
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.ServiceAccountFlagName, serviceAccountName, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountName,
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
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  serviceAccountName: my-service-account
      9 + |  source:
     10 + |    git:
     11 + |      ref:
     12 + |        branch: main
     13 + |      url: https://example.com/repo.git
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "create with serviceAccountName from file and flag",
			Args:         []string{flags.FilePathFlagName, "testdata/service-account-name.yaml", flags.ServiceAccountFlagName, serviceAccountNameUpdated, flags.YesFlagName},
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
						ServiceAccountName: &serviceAccountNameUpdated,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
     11 + |  serviceAccountName: my-service-account-updated
     12 + |  source:
     13 + |    git:
     14 + |      ref:
     15 + |        tag: tap-1.1
     16 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Created workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			Name:         "create from maven artifact using paramyaml",
			Args:         []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples"}`, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
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
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  params:
      9 + |  - name: maven
     10 + |    value:
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      version: 2.6.0
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
						Labels:    map[string]string{},
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
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  params:
      9 + |  - name: maven
     10 + |    value:
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      version: 2.6.0
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "create from maven artifact taking priority from flags",
			Args: []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic-test", "version": "2.6.1", "groupId": "org.springframework.samples.test"}`,
				flags.MavenArtifactFlagName, "spring-petclinic", flags.MavenVersionFlagName, "2.6.0", flags.MavenGroupFlagName, "org.springframework.samples", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
			ExpectCreates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
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
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  params:
      9 + |  - name: maven
     10 + |    value:
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      version: 2.6.0
❗ NOTICE: Maven configuration flags have overwritten values provided by "--params-yaml".
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update from maven artifact taking priority from flags",
			Args: []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "foo", "version": "1.0.0", "groupId": "bar"}`,
				flags.MavenArtifactFlagName, "spring-petclinic-test", flags.MavenVersionFlagName, "2.6.1", flags.MavenGroupFlagName, "org.springframework.samples.test", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Params(cartov1alpha1.Param{
							Name:  "maven",
							Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","version":"2.6.0"}`)},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Params: []cartov1alpha1.Param{
							{
								Name:  "maven",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic-test","groupId":"org.springframework.samples.test","version":"2.6.1"}`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  7,  7   |spec:
  8,  8   |  params:
  9,  9   |  - name: maven
 10, 10   |    value:
 11     - |      artifactId: spring-petclinic
 12     - |      groupId: org.springframework.samples
 13     - |      version: 2.6.0
     11 + |      artifactId: spring-petclinic-test
     12 + |      groupId: org.springframework.samples.test
     13 + |      version: 2.6.1
❗ NOTICE: Maven configuration flags have overwritten values provided by "--params-yaml".
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
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
						Labels:    map[string]string{},
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
🔎 Update workload:
...
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7     - |spec: {}
      7 + |spec:
      8 + |  params:
      9 + |  - name: maven
     10 + |    value:
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      version: 2.6.0
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update workload to add maven through flags",
			Args: []string{workloadName, flags.MavenArtifactFlagName, "spring-petclinic", flags.MavenVersionFlagName, "2.6.0", flags.MavenGroupFlagName, "org.springframework.samples", flags.MavenTypeFlagName, "jar", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Params: []cartov1alpha1.Param{
							{
								Name:  "maven",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","version":"2.6.0","type":"jar"}`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  name: my-workload
  6,  6   |  namespace: default
  7     - |spec: {}
      7 + |spec:
      8 + |  params:
      9 + |  - name: maven
     10 + |    value:
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      type: jar
     14 + |      version: 2.6.0
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update workload to change maven info through flags",
			Args: []string{workloadName, flags.MavenVersionFlagName, "2.6.1", flags.MavenTypeFlagName, "jar", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Params(cartov1alpha1.Param{
							Name:  "maven",
							Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","version":"2.6.0"}`)},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Params: []cartov1alpha1.Param{
							{
								Name:  "maven",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","version":"2.6.1","type":"jar"}`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  9,  9   |  - name: maven
 10, 10   |    value:
 11, 11   |      artifactId: spring-petclinic
 12, 12   |      groupId: org.springframework.samples
 13     - |      version: 2.6.0
     13 + |      type: jar
     14 + |      version: 2.6.1
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name: "update workload to override maven info through params yaml",
			Args: []string{workloadName, flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples"}`, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Params(cartov1alpha1.Param{
							Name:  "maven",
							Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"1.0.0","type":"baz"}`)},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels:    map[string]string{},
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
🔎 Update workload:
...
  7,  7   |spec:
  8,  8   |  params:
  9,  9   |  - name: maven
 10, 10   |    value:
 11     - |      artifactId: foo
 12     - |      groupId: bar
 13     - |      type: baz
 14     - |      version: 1.0.0
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      version: 2.6.0
👍 Updated workload "my-workload"

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
						Labels:    map[string]string{},
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
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  params:
      9 + |  - name: maven
     10 + |    value:
     11 + |      artifactId: spring-petclinic
     12 + |      groupId: org.springframework.samples
     13 + |      type: jar
     14 + |      version: 2.6.0
👍 Created workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

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
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

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
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_LABEL")
				return nil
			},
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
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
🔎 Create workload:
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
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_LABEL")
				os.Unsetenv("TANZU_APPS_TYPE")
				return nil
			},
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
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
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_TYPE")
				return nil
			},
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.TypeFlagName, "web", flags.YesFlagName},
			GivenObjects: givenNamespaceDefault,
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
			Name: "update type via allowed env var",
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				os.Setenv("TANZU_APPS_TYPE", "web")
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				os.Unsetenv("TANZU_APPS_TYPE")
				return nil
			},
			Args: []string{workloadName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "jar")
					}).
					SpecDie(
						func(d *diecartov1alpha1.WorkloadSpecDie) {
							d.Source(&cartov1alpha1.Source{
								Git: &cartov1alpha1.GitSource{
									URL: "https://github.com/sample-accelerators/spring-petclinic",
									Ref: cartov1alpha1.GitRef{
										Branch: "main",
									},
								},
							})
						}),
			},
			ExpectUpdates: []client.Object{
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
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: jar
      6 + |    apps.tanzu.vmware.com/workload-type: web
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  source:
...
👍 Updated workload "my-workload"

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`,
		},
		{
			Name:         "create - git source with terminal interaction",
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
		},
		{
			Name:         "create git source with terminal interaction no color",
			Config:       &cli.Config{NoColor: true, Scheme: scheme},
			GivenObjects: givenNamespaceDefault,
			Args:         []string{workloadName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.TypeFlagName, "web"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("Do you want to create this workload? [yN]: "))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`Created workload %q

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
Create workload:
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

Created workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, clitesting.ToInteractTerminal("? Do you want to create this workload? [yN]: y"), workloadName),
		},
		{
			Name:         "create - git source with terminal interaction reject",
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
		},
		{
			Name: "update - git source with terminal interaction",
			Args: []string{workloadName, flags.TypeFlagName, "api"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("? Really update the workload %q? [yN]: ", workloadName))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`👍 Updated workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
			},
			GivenObjects: func() []client.Object {
				w := &cartov1alpha1.Workload{
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
				}
				return append(givenNamespaceDefault, w)
			}(),
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "api",
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
🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: web
      6 + |    apps.tanzu.vmware.com/workload-type: api
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  source:
...
%s

👍 Updated workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, clitesting.ToInteractTerminal("❓ Really update the workload %q? [yN]: y", workloadName), workloadName),
		},
		{
			Name:   "update git source with terminal interaction no color",
			Config: &cli.Config{NoColor: true, Scheme: scheme},
			Args:   []string{workloadName, flags.TypeFlagName, "api"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("? Really update the workload %q? [yN]: ", workloadName))
				c.Send(clitesting.InteractInputLine("y"))
				c.ExpectString(clitesting.ToInteractOutput(`Updated workload %q

To see logs:   "tanzu apps workload tail %s --timestamp --since 1h"
To get status: "tanzu apps workload get %s"`, workloadName, workloadName, workloadName))
			},
			GivenObjects: func() []client.Object {
				w := &cartov1alpha1.Workload{
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
				}
				return append(givenNamespaceDefault, w)
			}(),
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      workloadName,
						Labels: map[string]string{
							"apps.tanzu.vmware.com/workload-type": "api",
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
Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: web
      6 + |    apps.tanzu.vmware.com/workload-type: api
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  source:
...
%s

Updated workload %q

To see logs:   "tanzu apps workload tail my-workload --timestamp --since 1h"
To get status: "tanzu apps workload get my-workload"

`, clitesting.ToInteractTerminal("? Really update the workload %q? [yN]: y", workloadName), workloadName),
		},
		{
			Name: "update - git source with terminal interaction rejected",
			Args: []string{workloadName, flags.TypeFlagName, "api"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("? Really update the workload %q? [yN]: ", workloadName))
				c.Send(clitesting.InteractInputLine("n"))
				c.ExpectString(clitesting.ToInteractOutput("Skipping workload %q", workloadName))
			},
			GivenObjects: func() []client.Object {
				w := &cartov1alpha1.Workload{
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
				}
				return append(givenNamespaceDefault, w)
			}(),
			ExpectOutput: fmt.Sprintf(`
🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: web
      6 + |    apps.tanzu.vmware.com/workload-type: api
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  source:
...
%s

Skipping workload %q`, clitesting.ToInteractTerminal("❓ Really update the workload %q? [yN]: n", workloadName), workloadName),
		},
		{
			Name: "update - git source with wrong answer terminal interaction",
			Args: []string{workloadName, flags.TypeFlagName, "api"},
			WithConsoleInteractions: func(t *testing.T, c *expect.Console) {
				c.ExpectString(clitesting.ToInteractTerminal("? Really update the workload %q? [yN]: ", workloadName))
				c.Send(clitesting.InteractInputLine("m"))
				c.ExpectString(clitesting.ToInteractTerminal("invalid input (not y, n, yes, or no)"))
				c.ExpectString(clitesting.ToInteractTerminal("? Really update the workload %q? [yN]: ", workloadName))
				c.Send(clitesting.InteractInputLine("n"))
				c.ExpectString(clitesting.ToInteractOutput("Skipping workload %q", workloadName))
			},
			GivenObjects: func() []client.Object {
				w := &cartov1alpha1.Workload{
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
				}
				return append(givenNamespaceDefault, w)
			}(),
			ExpectOutput: fmt.Sprintf(`
🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: web
      6 + |    apps.tanzu.vmware.com/workload-type: api
  7,  7   |  name: my-workload
  8,  8   |  namespace: default
  9,  9   |spec:
 10, 10   |  source:
...
%s

invalid input (not y, n, yes, or no)
%s

Skipping workload %q`,
				clitesting.ToInteractTerminal("❓ Really update the workload %q? [yN]: m", workloadName),
				clitesting.ToInteractTerminal("❓ Really update the workload %q? [yN]: n", workloadName), workloadName),
		},
		{
			Name: "update - replace update strategy",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/all-fields-workload.yaml",
				flags.TypeFlagName, "my-type",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Image("ubuntu:bionic")
						d.Env(
							corev1.EnvVar{
								Name:  "OVERRIDE_VAR",
								Value: "doesnt matter",
							},
						)
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "my-type",
						},
						Annotations: map[string]string{
							"controller-gen.kubebuilder.io/version": "v0.7.0",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
							Subpath: "./app",
						},
						Build: &cartov1alpha1.WorkloadBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "BP_MAVEN_POM_FILE",
									Value: "skip-pom.xml",
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
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "Secret",
									Name:       "stub-db",
								},
							},
						},
						Params: []cartov1alpha1.Param{
							{
								Name:  "services",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"},{"image":"postgres:9.6","name":"postgres"}]`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
  1,  1   |---
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
      5 + |  annotations:
      6 + |    controller-gen.kubebuilder.io/version: v0.7.0
      7 + |  labels:
      8 + |    app.kubernetes.io/part-of: spring-petclinic
      9 + |    apps.tanzu.vmware.com/workload-type: my-type
  5, 10   |  name: spring-petclinic
  6, 11   |  namespace: default
  7, 12   |spec:
     13 + |  build:
     14 + |    env:
     15 + |    - name: BP_MAVEN_POM_FILE
     16 + |      value: skip-pom.xml
  8, 17   |  env:
  9     - |  - name: OVERRIDE_VAR
 10     - |    value: doesnt matter
 11     - |  image: ubuntu:bionic
     18 + |  - name: SPRING_PROFILES_ACTIVE
     19 + |    value: mysql
     20 + |  params:
     21 + |  - name: services
     22 + |    value:
     23 + |    - image: mysql:5.7
     24 + |      name: mysql
     25 + |    - image: postgres:9.6
     26 + |      name: postgres
     27 + |  resources:
     28 + |    limits:
     29 + |      cpu: 500m
     30 + |      memory: 1Gi
     31 + |    requests:
     32 + |      cpu: 100m
     33 + |      memory: 1Gi
     34 + |  serviceAccountName: my-service-account
     35 + |  serviceClaims:
     36 + |  - name: database
     37 + |    ref:
     38 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     39 + |      kind: Secret
     40 + |      name: stub-db
     41 + |  source:
     42 + |    git:
     43 + |      ref:
     44 + |        branch: main
     45 + |      url: https://github.com/spring-projects/spring-petclinic.git
     46 + |    subPath: ./app
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace annotations",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-annotations.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddAnnotation("preserve-me", "should-exist")
						d.AddAnnotation("dont-preserve-me", "should-not-exist")
						d.AddAnnotation("my-annotation", "my-value")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Annotations: map[string]string{
							"controller-gen.kubebuilder.io/version": "v0.7.0",
							"preserve-me":                           "should-exist",
							"my-annotation":                         "my-updated-value",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  annotations:
  6     - |    dont-preserve-me: should-not-exist
  7     - |    my-annotation: my-value
      6 + |    controller-gen.kubebuilder.io/version: v0.7.0
      7 + |    my-annotation: my-updated-value
  8,  8   |    preserve-me: should-exist
  9,  9   |  name: spring-petclinic
 10, 10   |  namespace: default
 11, 11   |spec:
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace labels",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-labels.yaml",
				flags.LabelFlagName, "my-label=my-updated-value",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
						d.AddLabel("dont-preserve-me", "should-not-exist")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "should-overwrite")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"preserve-me":                         "should-exist",
							"my-label":                            "my-updated-value",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: should-overwrite
  7     - |    dont-preserve-me: should-not-exist
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
      8 + |    my-label: my-updated-value
  8,  9   |    preserve-me: should-exist
  9, 10   |  name: spring-petclinic
 10, 11   |  namespace: default
 11, 12   |spec:
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace labels with no color",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-labels.yaml",
				flags.LabelFlagName, "my-label=my-updated-value",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			Config: &cli.Config{NoColor: true, Scheme: scheme},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("preserve-me", "should-exist")
						d.AddLabel("dont-preserve-me", "should-not-exist")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "should-overwrite")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"preserve-me":                         "should-exist",
							"my-label":                            "my-updated-value",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

Update workload:
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    apps.tanzu.vmware.com/workload-type: should-overwrite
  7     - |    dont-preserve-me: should-not-exist
      6 + |    app.kubernetes.io/part-of: spring-petclinic
      7 + |    apps.tanzu.vmware.com/workload-type: web
      8 + |    my-label: my-updated-value
  8,  9   |    preserve-me: should-exist
  9, 10   |  name: spring-petclinic
 10, 11   |  namespace: default
 11, 12   |spec:
...
Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - add serviceAccountName",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-service-account-name.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  7,  7   |    apps.tanzu.vmware.com/workload-type: web
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
     11 + |  serviceAccountName: my-service-account
 11, 12   |  source:
 12, 13   |    git:
 13, 14   |      ref:
 14, 15   |        tag: tap-1.1
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - delete serviceAccountName",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-no-service-account-name.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.ServiceAccountName(&serviceAccountName)
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: nil,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  7,  7   |    apps.tanzu.vmware.com/workload-type: web
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
 11     - |  serviceAccountName: my-service-account
 12, 11   |  source:
 13, 12   |    git:
 14, 13   |      ref:
 15, 14   |        tag: tap-1.1
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - change serviceAccountName",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-service-account-name.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.ServiceAccountName(&serviceAccountNameUpdated)
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceAccountName: &serviceAccountName,
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  7,  7   |    apps.tanzu.vmware.com/workload-type: web
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
 11     - |  serviceAccountName: my-service-account-updated
     11 + |  serviceAccountName: my-service-account
 12, 12   |  source:
 13, 13   |    git:
 14, 14   |      ref:
 15, 15   |        tag: tap-1.1
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace buildenv",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-build-env.yaml",
				flags.BuildEnvFlagName, "my-new-build-env=my-new-value",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Build(&cartov1alpha1.WorkloadBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "my-build-env",
									Value: "my-build-env-value",
								},
								{
									Name:  "preserve-me",
									Value: "should-not-exist",
								},
								{
									Name:  "do-preserve-me",
									Value: "should-exist",
								},
							},
						})
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Build: &cartov1alpha1.WorkloadBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "my-new-build-env",
									Value: "my-new-value",
								},
								{
									Name:  "BP_MAVEN_POM_FILE",
									Value: "skip-pom.xml",
								},
								{
									Name:  "my-build-env",
									Value: "my-build-env-updated-value",
								},
								{
									Name:  "do-preserve-me",
									Value: "should-exist",
								},
							},
						},
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  9,  9   |  namespace: default
 10, 10   |spec:
 11, 11   |  build:
 12, 12   |    env:
     13 + |    - name: my-new-build-env
     14 + |      value: my-new-value
     15 + |    - name: BP_MAVEN_POM_FILE
     16 + |      value: skip-pom.xml
 13, 17   |    - name: my-build-env
 14     - |      value: my-build-env-value
 15     - |    - name: preserve-me
 16     - |      value: should-not-exist
     18 + |      value: my-build-env-updated-value
 17, 19   |    - name: do-preserve-me
 18, 20   |      value: should-exist
 19, 21   |  source:
 20, 22   |    git:
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace env",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-env.yaml",
				flags.EnvFlagName, "my-new-envvar=my-new-value",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Env(
							corev1.EnvVar{
								Name:  "my-envvar",
								Value: "my-envvar-value",
							},
							corev1.EnvVar{
								Name:  "dont-preserve-me",
								Value: "should-not-exist",
							},
							corev1.EnvVar{
								Name:  "preserve-me",
								Value: "should-exist",
							},
						)
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Env: []corev1.EnvVar{
							{
								Name:  "SPRING_PROFILES_ACTIVE",
								Value: "mysql",
							},
							{
								Name:  "my-envvar",
								Value: "my-envvar-updated-value",
							},
							{
								Name:  "my-new-envvar",
								Value: "my-new-value",
							},
							{
								Name:  "preserve-me",
								Value: "should-exist",
							},
						},
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
 11, 11   |  env:
     12 + |  - name: SPRING_PROFILES_ACTIVE
     13 + |    value: mysql
 12, 14   |  - name: my-envvar
 13     - |    value: my-envvar-value
 14     - |  - name: dont-preserve-me
 15     - |    value: should-not-exist
     15 + |    value: my-envvar-updated-value
     16 + |  - name: my-new-envvar
     17 + |    value: my-new-value
 16, 18   |  - name: preserve-me
 17, 19   |    value: should-exist
 18, 20   |  source:
 19, 21   |    git:
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace resources",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-resources.yaml",
				flags.LimitCPUFlagName, "400m",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Resources(&corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("500m"),
							},
						})
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Resources: &corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("400m"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  9,  9   |  namespace: default
 10, 10   |spec:
 11, 11   |  resources:
 12, 12   |    limits:
 13     - |      cpu: 500m
 14     - |      memory: 2Gi
     13 + |      cpu: 400m
     14 + |      memory: 1Gi
 15, 15   |    requests:
 16     - |      cpu: 500m
     16 + |      memory: 1Gi
 17, 17   |  source:
 18, 18   |    git:
 19, 19   |      ref:
 20, 20   |        tag: tap-1.1
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace service claims",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-service-claims.yaml",
				flags.ServiceRefFlagName, "my-new-svc-claim=my.api/v1:my-new-db-manager:my-new-db",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.ServiceClaims(cartov1alpha1.WorkloadServiceClaim{
							Name: "my-service-claim",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "PostgreSQL",
								Name:       "my-prod-db",
							},
						}, cartov1alpha1.WorkloadServiceClaim{
							Name: "my-second-service-claim",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "mysql",
								Name:       "my-sql-db",
							},
						}, cartov1alpha1.WorkloadServiceClaim{
							Name: "should-delete",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1",
								Kind:       "my-kind",
								Name:       "my-db",
							},
						})
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
							{
								Name: "database",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "Secret",
									Name:       "stub-db",
								},
							}, {
								Name: "my-service-claim",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "PostgreSQL",
									Name:       "my-prod-db-updated",
								},
							}, {
								Name: "my-second-service-claim",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "services.tanzu.vmware.com/v1alpha1",
									Kind:       "mysql",
									Name:       "my-sql-db",
								},
							}, {
								Name: "my-new-svc-claim",
								Ref: &cartov1alpha1.WorkloadServiceClaimReference{
									APIVersion: "my.api/v1",
									Kind:       "my-new-db-manager",
									Name:       "my-new-db",
								},
							},
						},
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
 11, 11   |  serviceClaims:
     12 + |  - name: database
     13 + |    ref:
     14 + |      apiVersion: services.tanzu.vmware.com/v1alpha1
     15 + |      kind: Secret
     16 + |      name: stub-db
 12, 17   |  - name: my-service-claim
 13, 18   |    ref:
 14, 19   |      apiVersion: services.tanzu.vmware.com/v1alpha1
 15, 20   |      kind: PostgreSQL
 16     - |      name: my-prod-db
     21 + |      name: my-prod-db-updated
 17, 22   |  - name: my-second-service-claim
 18, 23   |    ref:
 19, 24   |      apiVersion: services.tanzu.vmware.com/v1alpha1
 20, 25   |      kind: mysql
 21, 26   |      name: my-sql-db
 22     - |  - name: should-delete
     27 + |  - name: my-new-svc-claim
 23, 28   |    ref:
 24     - |      apiVersion: services.tanzu.vmware.com/v1
 25     - |      kind: my-kind
 26     - |      name: my-db
     29 + |      apiVersion: my.api/v1
     30 + |      kind: my-new-db-manager
     31 + |      name: my-new-db
 27, 32   |  source:
 28, 33   |    git:
 29, 34   |      ref:
 30, 35   |        tag: tap-1.1
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - add subpath",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-subpath.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Subpath: "./app",
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
 12, 12   |    git:
 13, 13   |      ref:
 14, 14   |        tag: tap-1.1
 15, 15   |      url: https://github.com/sample-accelerators/spring-petclinic
     16 + |    subPath: ./app
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - delete subpath",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-no-subpath.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Subpath: "./app",
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
 12, 12   |    git:
 13, 13   |      ref:
 14, 14   |        tag: tap-1.1
 15, 15   |      url: https://github.com/sample-accelerators/spring-petclinic
 16     - |    subPath: ./app
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - change subpath",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-subpath.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Subpath: "./cmd",
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Subpath: "./app",
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
 12, 12   |    git:
 13, 13   |      ref:
 14, 14   |        tag: tap-1.1
 15, 15   |      url: https://github.com/sample-accelerators/spring-petclinic
 16     - |    subPath: ./cmd
     16 + |    subPath: ./app
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace source",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-source.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/spring-projects/spring-petclinic.git",
								Ref: cartov1alpha1.GitRef{
									Branch: "main",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag:    "tap-1.1",
									Commit: "abcd123",
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
 10, 10   |spec:
 11, 11   |  source:
 12, 12   |    git:
 13, 13   |      ref:
 14     - |        branch: main
 15     - |      url: https://github.com/spring-projects/spring-petclinic.git
     14 + |        commit: abcd123
     15 + |        tag: tap-1.1
     16 + |      url: https://github.com/sample-accelerators/spring-petclinic
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update - replace params",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/replace-params.yaml",
				flags.AnnotationFlagName, "autoscaling.knative.dev/max-scale=4",
				flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Params(cartov1alpha1.Param{
							Name:  "https-ports",
							Value: apiextensionsv1.JSON{Raw: []byte(`{"ports":[{"name":"https","port":8443,"protocol":"TCP","targetPort":8443}]}`)},
						}, cartov1alpha1.Param{
							Name:  "services",
							Value: apiextensionsv1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"},{"image":"postgres:9.6","name":"postgres"}]`)},
						}, cartov1alpha1.Param{
							Name:  "should-delete",
							Value: apiextensionsv1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"}]`)},
						})
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ExpectUpdates: []client.Object{
				&cartov1alpha1.Workload{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "spring-petclinic",
						Labels: map[string]string{
							"app.kubernetes.io/part-of":           "spring-petclinic",
							"apps.tanzu.vmware.com/workload-type": "web",
						},
					},
					Spec: cartov1alpha1.WorkloadSpec{
						Source: &cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						},
						Params: []cartov1alpha1.Param{
							{
								Name:  "ports",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"ports":[{"name":"http","port":8080,"protocol":"TCP","targetPort":8080}]}`)},
							}, {
								Name:  "https-ports",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"ports":[{"name":"https","port":8553,"protocol":"TCP","targetPort":8553}]}`)},
							}, {
								Name:  "services",
								Value: apiextensionsv1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"},{"image":"postgres:9.6","name":"postgres"}]`)},
							}, {
								Name:  "annotations",
								Value: apiextensionsv1.JSON{Raw: []byte(`{"autoscaling.knative.dev/max-scale":"4","autoscaling.knative.dev/min-scale":"2"}`)},
							},
						},
					},
				},
			},
			ExpectOutput: `
❗ WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use "--update-strategy" to control strategy explicitly).

🔎 Update workload:
...
  8,  8   |  name: spring-petclinic
  9,  9   |  namespace: default
 10, 10   |spec:
 11, 11   |  params:
     12 + |  - name: ports
     13 + |    value:
     14 + |      ports:
     15 + |      - name: http
     16 + |        port: 8080
     17 + |        protocol: TCP
     18 + |        targetPort: 8080
 12, 19   |  - name: https-ports
 13, 20   |    value:
 14, 21   |      ports:
 15, 22   |      - name: https
 16     - |        port: 8443
     23 + |        port: 8553
 17, 24   |        protocol: TCP
 18     - |        targetPort: 8443
     25 + |        targetPort: 8553
 19, 26   |  - name: services
 20, 27   |    value:
 21, 28   |    - image: mysql:5.7
 22, 29   |      name: mysql
 23, 30   |    - image: postgres:9.6
 24, 31   |      name: postgres
 25     - |  - name: should-delete
     32 + |  - name: annotations
 26, 33   |    value:
 27     - |    - image: mysql:5.7
 28     - |      name: mysql
     34 + |      autoscaling.knative.dev/max-scale: "4"
     35 + |      autoscaling.knative.dev/min-scale: "2"
 29, 36   |  source:
 30, 37   |    git:
 31, 38   |      ref:
 32, 39   |        tag: tap-1.1
...
👍 Updated workload "spring-petclinic"

To see logs:   "tanzu apps workload tail spring-petclinic --timestamp --since 1h"
To get status: "tanzu apps workload get spring-petclinic"

`,
		},
		{
			Name: "update/replace - missing fields",
			Args: []string{flags.FilePathFlagName, "testdata/replace-update-strategy/invalid.yaml", flags.UpdateStrategyFlagName, "replace", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.AddLabel("apps.tanzu.vmware.com/workload-type", "web")
						d.AddLabel("app.kubernetes.io/part-of", "spring-petclinic")
					}).
					SpecDie(func(d *diecartov1alpha1.WorkloadSpecDie) {
						d.Source(&cartov1alpha1.Source{
							Git: &cartov1alpha1.GitSource{
								URL: "https://github.com/sample-accelerators/spring-petclinic",
								Ref: cartov1alpha1.GitRef{
									Tag: "tap-1.1",
								},
							},
						})
					}),
			},
			ShouldError: true,
		},
	}

	table.Run(t, scheme, func(ctx context.Context, c *cli.Config) *cobra.Command {
		// capture the cobra command so we can make assertions on cleanup, this will fail if tests are run parallel.
		cmd = commands.NewWorkloadApplyCommand(ctx, c)
		return cmd
	})
}
