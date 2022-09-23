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
	runtm "runtime"
	"strings"
	"testing"
	"time"

	diemetav1 "dies.dev/apis/meta/v1"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/wait"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diecartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestWorkloadDeleteOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:        "empty",
			Validatable: &commands.WorkloadDeleteOptions{},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(flags.NamespaceFlagName),
				validation.ErrMissingOneOf(flags.AllFlagName, cli.NamesArgumentName, flags.FilePathFlagName),
			),
		},
		{
			Name: "name",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default",
				Names:     []string{"my-workload"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default",
				Names:     []string{"my-"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid namespace",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default-",
				Names:     []string{"my"},
			},
			ShouldValidate: true,
		},
		{
			Name: "all",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default",
				All:       true,
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name + all",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default",
				Names:     []string{"my-workload"},
				All:       true,
			},
			ExpectFieldErrors: validation.ErrMultipleOneOf(flags.AllFlagName, cli.NamesArgumentName),
		},
		{
			Name: "invalid file + all",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default",
				FilePath:  "testdata/workload.yaml",
				All:       true,
			},
			ExpectFieldErrors: validation.ErrMultipleOneOf(flags.AllFlagName, flags.FilePathFlagName),
		},
		{
			Name: "wait",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace: "default",
				Names:     []string{"my-workload"},
				Wait:      true,
			},
			ShouldValidate: true,
		},
		{
			Name: "wait timeout",
			Validatable: &commands.WorkloadDeleteOptions{
				Namespace:   "default",
				Names:       []string{"my-workload"},
				WaitTimeout: time.Second,
				Wait:        true,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestWorkloadDeleteCommand(t *testing.T) {
	workloadName := "test-workload"
	workloadOtherName := "test-other-workload"
	defaultNamespace := "default"

	scheme := runtime.NewScheme()
	cartov1alpha1.AddToScheme(scheme)

	previousBackOffTime := wait.BackOffTime
	defer func() {
		wait.BackOffTime = previousBackOffTime
	}()
	wait.BackOffTime = 10 * time.Millisecond

	failingReactionFunc := func(verb, resource string) clitesting.ReactionFunc {
		apiCount, callCountToFail := 0, 1
		return func(action clitesting.Action) (bool, runtime.Object, error) {
			if verb == action.GetVerb() && resource == action.GetResource().Resource {
				if apiCount == callCountToFail {
					return true, nil, fmt.Errorf("client error")
				}
				apiCount++
			}
			return true, nil, nil
		}
	}

	parent := diecartov1alpha1.WorkloadBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(workloadName)
			d.Namespace(defaultNamespace)
		})

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all workloads",
			Args: []string{flags.AllFlagName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectDeleteCollections: []rtesting.DeleteCollectionRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				// TOOD: (@shashwathi) Remove the following fields and label fields after the following fix is merged
				// https://github.com/vmware-labs/reconciler-runtime/pull/263 in reconciler-runtime.
				Fields: fields.Everything(),
				Labels: labels.NewSelector(),
			}},
			ExpectOutput: `
Deleted workloads in namespace "default"
`,
		},
		{
			// TODO figure out how to send input to the confirmation
			Skip:  true,
			Name:  "delete all workloads, prompt confirmed",
			Args:  []string{flags.AllFlagName},
			Stdin: []byte("yes"),
			GivenObjects: []client.Object{
				parent,
			},
			ExpectDeleteCollections: []rtesting.DeleteCollectionRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
			}},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, `Really delete all workloads in the namespace "default"?`) {
					t.Errorf("expected output to contain delete prompt")
				}
				if !strings.Contains(output, `Deleted workloads in namespace "default"`) {
					t.Errorf("expected output to contain skip confirmation")
				}
			},
		},
		{
			Name:  "delete all workloads, prompt denied",
			Args:  []string{flags.AllFlagName},
			Stdin: []byte("no"),
			GivenObjects: []client.Object{
				parent,
			},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, `Really delete all workloads in the namespace "default"?`) {
					t.Errorf("expected output to contain delete prompt")
				}
				if !strings.Contains(output, `Skipping workloads in namespace "default"`) {
					t.Errorf("expected output to contain skip confirmation")
				}
			},
		},
		{
			Name: "delete all workloads error",
			Args: []string{flags.AllFlagName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("delete-collection", "Workload"),
			},
			ExpectDeleteCollections: []rtesting.DeleteCollectionRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				// TOOD: (@shashwathi) Remove the following fields and label fields after the following fix is merged
				// https://github.com/vmware-labs/reconciler-runtime/pull/263 in reconciler-runtime.
				Fields: fields.Everything(),
				Labels: labels.NewSelector(),
			}},
			ShouldError: true,
		},
		{
			Name: "delete workload",
			Args: []string{workloadName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}},
			ExpectOutput: `
Deleted workload "test-workload"
`,
		},
		{
			// TODO figure out how to send input to the confirmation
			Skip:  true,
			Name:  "delete workload, prompt confirmed",
			Args:  []string{workloadName},
			Stdin: []byte("yes"),
			GivenObjects: []client.Object{
				parent,
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, `Really delete the workload "test-workload"?`) {
					t.Errorf("expected output to contain delete prompt")
				}
				if !strings.Contains(output, `Deleted workload "test-workload"`) {
					t.Errorf("expected output to contain delete confirmation")
				}
			},
		},
		{
			Name:  "delete workload, prompt denied",
			Args:  []string{workloadName},
			Stdin: []byte("no"),
			GivenObjects: []client.Object{
				parent,
			},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, `Really delete the workload "test-workload"?`) {
					t.Errorf("expected output to contain delete prompt")
				}
				if !strings.Contains(output, `Skipping workload "test-workload"`) {
					t.Errorf("expected output to contain skip confirmation")
				}
			},
		},
		{
			Name: "delete workloads",
			Args: []string{workloadName, workloadOtherName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent,
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name(workloadOtherName)
						d.Namespace(defaultNamespace)
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}, {
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadOtherName,
			}},
			ExpectOutput: `
Deleted workload "test-workload"
Deleted workload "test-other-workload"
`,
		},
		{
			Name: "workload does not exist",
			Args: []string{workloadName, flags.YesFlagName},
			ExpectOutput: `
Workload "test-workload" does not exist
`,
		},
		{
			Name: "delete error",
			Args: []string{workloadName, flags.YesFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("delete", "Workload"),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete workload confirmed after wait",
			Args: []string{workloadName, flags.YesFlagName, flags.WaitFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}},
			ExpectOutput: `
Deleted workload "test-workload"
Waiting for workload "test-workload" to be deleted...
Workload "test-workload" was deleted
`,
		},
		{
			Name: "delete workload failed with wait",
			Args: []string{workloadName, flags.YesFlagName, flags.WaitFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			WithReactors: []clitesting.ReactionFunc{
				// 1st get call needs to be successul
				// and 3rd call to Get should return error for UntilDelete func to error
				failingReactionFunc("get", "Workload"),
			},
			ShouldError: true,
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}},
		}, {
			Name: "delete workload failed with wait timeout error",
			Skip: runtm.GOOS == "windows",
			Args: []string{workloadName, flags.YesFlagName, flags.WaitFlagName},
			GivenObjects: []client.Object{
				parent,
			},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				return ctx, nil
			},
			ShouldError: true,
			ExpectOutput: `
Deleted workload "test-workload"
Waiting for workload "test-workload" to be deleted...
Error: timeout after 1m0s waiting for "test-workload" to be deleted
To view status run: tanzu apps workload get test-workload --namespace default
`,
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      workloadName,
			}},
		},
		{
			Name: "accept yaml file through stdin",
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
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      "spring-petclinic",
			}},
			ExpectOutput: `
Deleted workload "spring-petclinic"
`,
		},
		{
			Name: "fail to accept yaml file through stdin due to wrong api version",
			Args: []string{flags.FilePathFlagName, "-", flags.YesFlagName},
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
			Name: "delete workload from file",
			Args: []string{flags.FilePathFlagName, "testdata/workload.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.Namespace(defaultNamespace)
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: defaultNamespace,
				Name:      "spring-petclinic",
			}},
			ExpectOutput: `
Deleted workload "spring-petclinic"
`,
		},
		{
			Name:        "file - missing",
			Args:        []string{flags.FilePathFlagName, "testdata/missing.yaml", flags.YesFlagName},
			ShouldError: true,
		},
		{
			Name: "delete workload from file with custom namespace",
			Args: []string{flags.NamespaceFlagName, "test-namespace", flags.FilePathFlagName, "testdata/workload.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.Namespace("test-namespace")
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: "test-namespace",
				Name:      "spring-petclinic",
			}},
			ExpectOutput: `
Deleted workload "spring-petclinic"
`,
		},
		{
			Name: "delete workload from file with custom namespace in file",
			Args: []string{flags.FilePathFlagName, "testdata/workload-custom-namespace.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.Namespace("test-namespace")
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: "test-namespace",
				Name:      "spring-petclinic",
			}},
			ExpectOutput: `
Deleted workload "spring-petclinic"
`,
		},
		{
			Name: "delete workload from file with namespace from cli args",
			Args: []string{workloadName, flags.NamespaceFlagName, "test", flags.FilePathFlagName, "testdata/workload-custom-namespace.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.Namespace("test")
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{{
				Group:     "carto.run",
				Kind:      "Workload",
				Namespace: "test",
				Name:      "spring-petclinic",
			}},
			ExpectOutput: `
Workload "test-workload" does not exist
Deleted workload "spring-petclinic"
`,
		},
		{
			Name: "delete workload with file and a name from cli args",
			Args: []string{workloadName, flags.FilePathFlagName, "testdata/workload.yaml", flags.YesFlagName},
			GivenObjects: []client.Object{
				parent,

				diecartov1alpha1.WorkloadBlank.
					MetadataDie(func(d *diemetav1.ObjectMetaDie) {
						d.Name("spring-petclinic")
						d.Namespace(defaultNamespace)
					}),
			},
			ExpectDeletes: []rtesting.DeleteRef{
				{
					Group:     "carto.run",
					Kind:      "Workload",
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				{
					Group:     "carto.run",
					Kind:      "Workload",
					Namespace: defaultNamespace,
					Name:      "spring-petclinic",
				},
			},
			ExpectOutput: `
Deleted workload "test-workload"
Deleted workload "spring-petclinic"
`,
		},
	}

	table.Run(t, scheme, commands.NewWorkloadDeleteCommand)
}
