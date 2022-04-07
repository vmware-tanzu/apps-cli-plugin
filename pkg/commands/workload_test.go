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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	ggcrregistry "github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

func TestWorkloadCommand(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	table := clitesting.CommandTestSuite{
		{
			Name: "empty",
			Args: []string{},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, "Commands:") {
					t.Errorf("output expected to contain help with nested commands to call")
				}
			},
		},
	}

	table.Run(t, scheme, commands.NewWorkloadCommand)
}

func TestWorkloadOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name: "empty namespace",
			Validatable: &commands.WorkloadOptions{
				Name: "my-resource",
			},
			ExpectFieldErrors: validation.ErrInvalidValue("", flags.NamespaceFlagName),
		},
		{
			Name: "empty name",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
			},
			ExpectFieldErrors: validation.ErrInvalidValue("", cli.NameArgumentName),
		},
		{
			Name: "valid env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Env:       []string{"FOO=bar"},
			},
			ShouldValidate: true,
		},
		{
			Name: "env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Env:       []string{"foo=bar", "bleep=bloop"},
			},
			ShouldValidate: true,
		},
		{
			Name: "remove env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Env:       []string{"FOO-"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Env:       []string{"FOO"},
			},
			ExpectFieldErrors: validation.ErrInvalidArrayValue("FOO", flags.EnvFlagName, 0),
		},
		{
			Name: "valid build env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				BuildEnv:  []string{"FOO=bar"},
			},
			ShouldValidate: true,
		},
		{
			Name: "build env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				BuildEnv:  []string{"foo=bar", "bleep=bloop"},
			},
			ShouldValidate: true,
		},
		{
			Name: "remove build env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				BuildEnv:  []string{"FOO-"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid build env",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				BuildEnv:  []string{"FOO"},
			},
			ExpectFieldErrors: validation.ErrInvalidArrayValue("FOO", flags.BuildEnvFlagName, 0),
		},
		{
			Name: "params",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Params:    []string{"foo=bar", "bleep=bloop"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid params",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Params:    []string{"foo=bar", "bleep"},
			},
			ShouldValidate:    false,
			ExpectFieldErrors: validation.ErrInvalidValue("bleep", flags.ParamFlagName+"[1]"),
		},
		{
			Name: "valid resources limits",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				LimitCPU:    "1",
				LimitMemory: "1Gi",
			},
			ShouldValidate: true,
		},
		{
			Name: "valid resources limits and request",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				LimitCPU:      "2",
				LimitMemory:   "2Gi",
				RequestCPU:    "1",
				RequestMemory: "1Gi",
			},
			ShouldValidate: true,
		},
		{
			Name: "valid equal resources limits and request",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				LimitCPU:      "2",
				LimitMemory:   "2Gi",
				RequestCPU:    "2",
				RequestMemory: "2Gi",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid resource limits",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				LimitCPU:    "nan-cpu",
				LimitMemory: "nan-memory",
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrInvalidValue("nan-cpu", flags.LimitCPUFlagName),
				validation.ErrInvalidValue("nan-memory", flags.LimitMemoryFlagName),
			),
		},
		{
			Name: "valid resources requests",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RequestCPU:    "1",
				RequestMemory: "1Gi",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid resource requests",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RequestCPU:    "nan-cpu",
				RequestMemory: "nan-memory",
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrInvalidValue("nan-cpu", flags.RequestCPUFlagName),
				validation.ErrInvalidValue("nan-memory", flags.RequestMemoryFlagName),
			),
		},
		{
			Name: "invalid CPU requests compared to limit",
			Validatable: &commands.WorkloadOptions{
				Namespace:  "default",
				Name:       "my-resource",
				RequestCPU: "2",
				LimitCPU:   "1",
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrInvalidValue("2", flags.RequestCPUFlagName),
			),
		},
		{
			Name: "invalid memory requests compared to limit",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RequestMemory: "2Gi",
				LimitMemory:   "1Gi",
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrInvalidValue("2Gi", flags.RequestMemoryFlagName),
			),
		},
		{
			Name: "label",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Labels:    []string{"foo=bar", "bleep=bloop"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid label",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Labels:    []string{"foo=bar", "bleep"},
			},
			ShouldValidate:    false,
			ExpectFieldErrors: validation.ErrInvalidValue("bleep", flags.LabelFlagName+"[1]"),
		},
		{
			Name: "remove label",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Labels:    []string{"FOO-"},
			},
			ShouldValidate: true,
		},
		{
			Name: "annotations",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				Annotations: []string{"foo=bar", "bleep=bloop"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid annotations",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				Annotations: []string{"foo=bar", "bleep"},
			},
			ShouldValidate:    false,
			ExpectFieldErrors: validation.ErrInvalidValue("bleep", flags.AnnotationFlagName+"[1]"),
		},
		{
			Name: "remove annotations",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				Annotations: []string{"FOO-"},
			},
			ShouldValidate: true,
		},
		{
			Name: "valid service references",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				ServiceRefs: []string{"database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db"},
			},
			ShouldValidate: true,
		},
		{
			Name: "valid service references with namespace",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				ServiceRefs: []string{"database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-ns:my-prod-db"},
			},
			ShouldValidate: true,
		},
		{
			Name: "valid delete service references",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				ServiceRefs: []string{"database-"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid service references",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				ServiceRefs: []string{"database=PostgreSQL:my-prod-db"},
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrInvalidArrayValue("PostgreSQL:my-prod-db", flags.ServiceRefFlagName, 0),
			),
		},
		{
			Name: "invalid service references with spaces",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				ServiceRefs: []string{"data base=v1:PostgreSQL:my-prod-db"},
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrInvalidArrayValue("data base", flags.ServiceRefFlagName, 0),
			),
		},
		{
			Name: "git source",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				GitRepo:   "https://example.com/repo.git",
				GitBranch: "main",
			},
			ShouldValidate: true,
		},
		{
			Name: "source image",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				SourceImage: "repo.example/image:tag",
			},
			ShouldValidate: true,
		},
		{
			Name: "pre-built image",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Image:     "repo.example/image:tag",
			},
			ShouldValidate: true,
		},
		{
			Name: "all sources",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				GitRepo:     "https://example.com/repo.git",
				GitBranch:   "main",
				SourceImage: "repo.example/image:tag",
				Image:       "repo.example/image:tag",
			},
			ShouldValidate: false,
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMultipleOneOf(flags.GitFlagWildcard, flags.SourceImageFlagName, flags.ImageFlagName),
			),
		},
		{
			Name: "wait",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				Wait:      true,
			},
			ShouldValidate: true,
		},
		{
			Name: "wait timeout",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				WaitTimeout: time.Second,
				Wait:        true,
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run",
			Validatable: &commands.WorkloadOptions{
				Namespace: "default",
				Name:      "my-resource",
				DryRun:    true,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestWorkloadOptionsApplyOptionsToWorkload(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	appName := "my-app"
	typeName := "job"
	gitRepo := "https://example.com/repo.git"
	gitBranch := "main"
	subPath := "./cmd"
	gitTag := "v0.0.1"
	gitCommit := "abcdefg"

	scheme := runtime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)
	c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

	tests := []struct {
		name     string
		args     []string
		input    *cartov1alpha1.Workload
		expected *cartov1alpha1.Workload
	}{
		{
			name: "add/update/remove label",
			args: []string{flags.LabelFlagName, "NEW=value", flags.LabelFlagName, "FOO=bar", flags.LabelFlagName, "BAR-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"FOO": "foo",
						"BAR": "bar",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"NEW": "value",
						"FOO": "bar",
					},
				},
			},
		},
		{
			name: "Support comma separated list of labels",
			args: []string{flags.LabelFlagName, "NEW=value,FOO=bar,BAR-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"FOO": "foo",
						"BAR": "bar",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"NEW": "value",
						"FOO": "bar",
					},
				},
			},
		},
		{
			name: "workload with labels add/delete",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.LabelFlagName, "apps.tanzu.vmware.com/workload-type=web", flags.LabelFlagName, "apps.tanzu.vmware.com/workload-type-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
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
		{
			name: "add/update annotation",
			args: []string{flags.AnnotationFlagName, "NEW=value", flags.AnnotationFlagName, "FOO=bar", flags.AnnotationFlagName, "removeme-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
					Params: []cartov1alpha1.Param{
						{
							Name:  "annotations",
							Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"baz","removeme":"xyz"}`)},
						},
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
					Params: []cartov1alpha1.Param{
						{
							Name:  "annotations",
							Value: apiextensionsv1.JSON{Raw: []byte(`{"FOO":"bar","NEW":"value","foo":"baz"}`)},
						},
					},
				},
			},
		},
		{
			name: "update params",
			args: []string{flags.ParamFlagName, "foo=bar", flags.ParamFlagName, "removeme-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
					Params: []cartov1alpha1.Param{
						{
							Name:  "removeme",
							Value: apiextensionsv1.JSON{Raw: []byte(`"bye"`)},
						},
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
					Params: []cartov1alpha1.Param{
						{
							Name:  "foo",
							Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
						},
					},
				},
			},
		},
		{
			name: "update app",
			args: []string{flags.AppFlagName, appName},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: appName,
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "update type",
			args: []string{flags.TypeFlagName, typeName},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: typeName,
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "workload debug",
			args: []string{flags.DebugFlagName},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{},
					Image:  "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						{
							Name:  "debug",
							Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
						},
					},
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "workload debug disabled",
			args: []string{fmt.Sprintf("%s=%s", flags.DebugFlagName, "false")},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						{
							Name:  "debug",
							Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
						},
					},
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						// the debug param was removed
					},
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "workload with live-update",
			args: []string{flags.LiveUpdateFlagName, flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						{
							Name:  "live-update",
							Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
						},
					},
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
		{
			name: "workload with live-update disabled",
			args: []string{fmt.Sprintf("%s=%s", flags.LiveUpdateFlagName, "false")},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						{
							Name:  "live-update",
							Value: apiextensionsv1.JSON{Raw: []byte(`"true"`)},
						},
					},
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						// the live update param was removed
					},
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "workload git repo",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.SubPathFlagName, subPath},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: gitRepo,
							Ref: cartov1alpha1.GitRef{
								Branch: gitBranch,
							},
						},
						Subpath: "./cmd",
					},
				},
			},
		},
		{
			name: "workload with git tag",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.GitTagFlagName, gitTag},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: gitRepo,
							Ref: cartov1alpha1.GitRef{
								Branch: gitBranch,
								Tag:    gitTag,
							},
						},
					},
				},
			},
		},
		{
			name: "workload with git commit",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.GitCommitFlagName, gitCommit},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: gitRepo,
							Ref: cartov1alpha1.GitRef{
								Branch: gitBranch,
								Commit: gitCommit,
							},
						},
					},
				},
			},
		},
		{
			name: "workload with source image",
			args: []string{flags.SourceImageFlagName, "repo.example/image:tag", flags.SubPathFlagName, "workspace"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image:   "repo.example/image:tag",
						Subpath: "workspace",
					},
				},
			},
		},
		{
			name: "workload with image",
			args: []string{flags.ImageFlagName, "docker.io/library/ubuntu:bionic"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "docker.io/library/ubuntu:bionic",
				},
			},
		},
		{
			name: "add/update/remove env",
			args: []string{flags.EnvFlagName, "NEW=value", flags.EnvFlagName, "FOO=bar", flags.EnvFlagName, "BAR-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Env: []corev1.EnvVar{
						{Name: "FOO", Value: "foo"},
						{Name: "BAR", Value: "bar"},
					},
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Env: []corev1.EnvVar{
						{Name: "FOO", Value: "bar"},
						{Name: "NEW", Value: "value"},
					},
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "add/update/remove build env",
			args: []string{flags.BuildEnvFlagName, "NEW=value", flags.BuildEnvFlagName, "FOO=bar", flags.BuildEnvFlagName, "BAR-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Build: &cartov1alpha1.WorkloadBuild{
						Env: []corev1.EnvVar{
							{Name: "FOO", Value: "foo"},
							{Name: "BAR", Value: "bar"},
						},
					},
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Build: &cartov1alpha1.WorkloadBuild{
						Env: []corev1.EnvVar{
							{Name: "FOO", Value: "bar"},
							{Name: "NEW", Value: "value"},
						},
					},
					Image: "ubuntu:bionic",
				},
			},
		},
		{
			name: "workload with optional flags",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.AppFlagName, appName, flags.TypeFlagName, typeName, flags.ParamFlagName, "foo=bar", flags.ParamFlagName, "bleep=bloop", flags.ParamFlagName, "bleep-", flags.EnvFlagName, "FOO=bar", flags.BuildEnvFlagName, "BAR=baz", flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.LimitCPUFlagName, "500m", flags.LimitMemoryFlagName, "1Gi", flags.LabelFlagName, "build.tanzu.vmware.com/supply-chain=custom", flags.YesFlagName},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName:               appName,
						apis.WorkloadTypeLabelName:            typeName,
						"build.tanzu.vmware.com/supply-chain": "custom",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{{
						Name:  "foo",
						Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
					}},
					Build: &cartov1alpha1.WorkloadBuild{
						Env: []corev1.EnvVar{{
							Name:  "BAR",
							Value: "baz",
						}},
					},
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: gitRepo,
							Ref: cartov1alpha1.GitRef{
								Branch: gitBranch,
							},
						},
					},
					Env: []corev1.EnvVar{{
						Name:  "FOO",
						Value: "bar",
					}},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
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
		{
			name: "workload with CPU and memory flags",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.AppFlagName, appName, flags.TypeFlagName, typeName, flags.LimitCPUFlagName, "500m", flags.LimitMemoryFlagName, "1Gi", flags.RequestCPUFlagName, "250m", flags.RequestMemoryFlagName, "1Gi", flags.LabelFlagName, "build.tanzu.vmware.com/supply-chain=custom", flags.YesFlagName},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName:               appName,
						apis.WorkloadTypeLabelName:            typeName,
						"build.tanzu.vmware.com/supply-chain": "custom",
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
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		{
			name: "add/remove service references",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-db", flags.ServiceRefFlagName, "cache-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
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
					ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
						{
							Name: "cache",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "redis",
								Name:       "my-cache",
							},
						},
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
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
		{
			name: "add/remove service references with namespace",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:PostgreSQL:my-prod-ns:my-prod-db", flags.ServiceRefFlagName, "cache-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"cache":{"namespace":"my-prod-ns"}}}}`,
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
					ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
						{
							Name: "cache",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "redis",
								Name:       "my-cache",
							},
						},
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
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
		{
			name: "update service references",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-db"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
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
					ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
						{
							Name: "database",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "PostgreSQL",
								Name:       "my-prod",
							},
						},
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
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
					ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
						{
							Name: "database",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "MySQL",
								Name:       "my-prod-db",
							},
						},
					},
				},
			},
		},
		{
			name: "update service references with namespaces",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-ns:my-prod-db"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
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
					ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
						{
							Name: "database",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "PostgreSQL",
								Name:       "my-prod",
							},
						},
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName: workloadName,
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
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
					ServiceClaims: []cartov1alpha1.WorkloadServiceClaim{
						{
							Name: "database",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "MySQL",
								Name:       "my-prod-db",
							},
						},
					},
				},
			},
		},
		{
			name: "update resources limits",
			args: []string{flags.LimitCPUFlagName, "500m", flags.LimitMemoryFlagName, "1Gi"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		cmd := &cobra.Command{}
		ctx := cli.WithCommand(context.Background(), cmd)

		opts := &commands.WorkloadOptions{}
		opts.DefineFlags(ctx, c, cmd)
		cmd.ParseFlags(test.args)

		actual := test.input.DeepCopy()
		opts.ApplyOptionsToWorkload(ctx, actual)
		t.Run(test.name, func(t *testing.T) {
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("ApplyOptionsToWorkload() (-want, +got) = %s", diff)
			}
		})
	}
}

func TestWorkloadOptionsPublishLocalSource(t *testing.T) {
	registry, err := ggcrregistry.TLS("localhost")
	utilruntime.Must(err)
	defer registry.Close()
	u, err := url.Parse(registry.URL)
	utilruntime.Must(err)
	registryHost := u.Host

	tests := []struct {
		name           string
		args           []string
		input          string
		expected       string
		shouldError    bool
		expectedOutput string
	}{
		{
			name:     "local source",
			args:     []string{flags.LocalPathFlagName, "testdata/local-source", flags.YesFlagName},
			input:    fmt.Sprintf("%s/hello:source", registryHost),
			expected: fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
			expectedOutput: `
Publishing source in "testdata/local-source" to "` + registryHost + `/hello:source"...
Published source
`,
		},
		{
			name:     "with digest",
			args:     []string{flags.LocalPathFlagName, "testdata/local-source", flags.YesFlagName},
			input:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "0000000000000000000000000000000000000000000000000000000000000000"),
			expected: fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
			expectedOutput: `
Publishing source in "testdata/local-source" to "` + registryHost + `/hello:source"...
Published source
`,
		},
		{
			name:           "no local path",
			args:           []string{},
			input:          fmt.Sprintf("%s/hello:source", registryHost),
			expected:       fmt.Sprintf("%s/hello:source", registryHost),
			expectedOutput: "",
		},
		{
			name:        "publish local source with error",
			args:        []string{flags.LocalPathFlagName, "testdata/local-source", flags.YesFlagName},
			input:       "a",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)
			ctx = source.StashGgcrRemoteOptions(ctx, remote.WithTransport(registry.Client().Transport))

			opts := &commands.WorkloadOptions{}
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)

			workload := &cartov1alpha1.Workload{
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: test.input,
					},
				},
			}

			_, err := opts.PublishLocalSource(ctx, c, workload)
			if err != nil && !test.shouldError {
				t.Errorf("PublishLocalSource() errored %v", err)
			}

			if err == nil && test.shouldError {
				t.Errorf("PublishLocalSource() expected error")
			}

			if test.shouldError {
				return
			}

			if test.expected != workload.Spec.Source.Image {
				t.Errorf("PublishLocalSource() wanted %q, got %q", test.expected, workload.Spec.Source.Image)
			}

			if diff := cmp.Diff(strings.TrimSpace(test.expectedOutput), strings.TrimSpace(output.String())); diff != "" {
				t.Errorf("PublishLocalSource() (-want, +got) = %s", diff)
			}
		})
	}
}

func TestWorkloadOptionsCreate(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		input          *cartov1alpha1.Workload
		shouldError    bool
		expectedOutput string
		withReactors   []clitesting.ReactionFunc
	}{
		{
			name: "Create workload succesfully",
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			shouldError: false,
			expectedOutput: `
Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: my-workload
      6 + |  namespace: default
      7 + |spec:
      8 + |  image: ubuntu:bionic

Created workload "my-workload"`,
		},
		{
			name: "Create workload error",
			withReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("create", "Workload"),
			},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = cartov1alpha1.AddToScheme(scheme)
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			client := clitesting.NewFakeClient(scheme)
			c.Client = clitesting.NewFakeCliClient(client)

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)

			opts := &commands.WorkloadOptions{}
			opts.DefineFlags(ctx, c, cmd)
			opts.Yes = true

			for i := range test.withReactors {
				// in reverse order since we prepend
				reactor := test.withReactors[len(test.withReactors)-1-i]
				client.PrependReactor("*", "*", reactor)
			}

			_, err := opts.Create(ctx, c, test.input)

			if err != nil && !test.shouldError {
				t.Errorf("Create() errored %v", err)
			}
			if err == nil && test.shouldError {
				t.Errorf("Create() expected error")
			}
			if test.shouldError {
				return
			}

			if diff := cmp.Diff(strings.TrimSpace(test.expectedOutput), strings.TrimSpace(output.String())); diff != "" {
				t.Errorf("Create() (-want, +got) = %s", diff)
			}
		})
	}
}

func TestWorkloadOptionsUpdate(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	scheme := runtime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)
	c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

	tests := []struct {
		name           string
		args           []string
		givenWorkload  *cartov1alpha1.Workload
		shouldError    bool
		expectedOutput string
		withReactors   []clitesting.ReactionFunc
	}{
		{
			name: "Update workload successfully",
			args: []string{flags.LabelFlagName, "NEW=value", flags.YesFlagName},
			givenWorkload: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"FOO": "bar",
					},
				},
			},
			shouldError: false,
			expectedOutput: `
Update workload:
...
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6,  6   |    FOO: bar
      7 + |    NEW: value
  7,  8   |  name: my-workload
  8,  9   |  namespace: default
  9, 10   |spec: {}

Updated workload "my-workload"
`,
		},
		{
			name: "Update workload error",
			args: []string{flags.LabelFlagName, "NEW=value", flags.YesFlagName},
			withReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("update", "Workload"),
			},
			givenWorkload: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			shouldError: true,
		},
		{
			name: "Update no change",
			args: []string{},
			givenWorkload: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			shouldError:    false,
			expectedOutput: "Workload is unchanged, skipping update",
		},
		{
			name: "Update conflict",
			args: []string{flags.LabelFlagName, "NEW=value", flags.YesFlagName},
			withReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("update", "Workload", clitesting.InduceFailureOpts{
					Error: apierrs.NewConflict(schema.GroupResource{Group: "carto.run", Resource: "workloads"}, workloadName, fmt.Errorf("induced conflict")),
				}),
			},
			givenWorkload: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = cartov1alpha1.AddToScheme(scheme)
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			fakeClient := clitesting.NewFakeClient(scheme, test.givenWorkload)

			for i := range test.withReactors {
				// in reverse order since we prepend
				reactor := test.withReactors[len(test.withReactors)-1-i]
				fakeClient.PrependReactor("*", "*", reactor)
			}

			c.Client = clitesting.NewFakeCliClient(fakeClient)

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)

			currentWorkload := &cartov1alpha1.Workload{}
			err := c.Get(ctx, client.ObjectKey{Namespace: defaultNamespace, Name: workloadName}, currentWorkload)

			if err != nil {
				t.Errorf("Update() errored %v", err)
			}

			opts := &commands.WorkloadOptions{}
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)

			workload := currentWorkload.DeepCopy()
			opts.ApplyOptionsToWorkload(ctx, workload)
			_, err = opts.Update(ctx, c, currentWorkload, workload)

			if err != nil && !test.shouldError {
				t.Errorf("Update() errored %v", err)
			}
			if err == nil && test.shouldError {
				t.Errorf("Update() expected error")
			}
			if test.shouldError {
				return
			}

			if diff := cmp.Diff(strings.TrimSpace(test.expectedOutput), strings.TrimSpace(output.String())); diff != "" {
				t.Errorf("Update() (-want, +got) = %s", diff)
			}
		})
	}
}

func TestLoadInputWorkload(t *testing.T) {
	scheme := runtime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)

	tests := []struct {
		name        string
		file        string
		shouldError bool
		stdin       io.Reader
	}{
		{
			name:  "loads workload from file",
			file:  "testdata/workload.yaml",
			stdin: c.Stdin,
		},
		{
			name: "loads workload from stdin",
			file: "-",
			stdin: strings.NewReader(`
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
`,
			),
		},
		{
			name:        "error loading non-existent file",
			file:        "testdata/workload1.yaml",
			stdin:       c.Stdin,
			shouldError: true,
		},
		{
			name: "error with workload type",
			file: "-",
			stdin: strings.NewReader(`
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
`,
			),
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := &commands.WorkloadOptions{
				FilePath: test.file,
			}

			err := opts.LoadInputWorkload(test.stdin, &cartov1alpha1.Workload{})

			if (err == nil) == test.shouldError {
				t.Errorf("Load() shouldErr %t, got %v", test.shouldError, err)
			} else if test.shouldError {
				return
			}
		})
	}
}
