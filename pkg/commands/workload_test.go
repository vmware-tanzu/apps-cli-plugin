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
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	ggcrregistry "github.com/google/go-containerregistry/pkg/registry"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/logger"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/logger/fake"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
	fake_source "github.com/vmware-tanzu/apps-cli-plugin/pkg/source/fake"
)

var localSource = filepath.Join("testdata", "local-source")

func TestWorkloadCommand(t *testing.T) {
	scheme := k8sruntime.NewScheme()
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
	filePathSeparator := string(filepath.Separator)
	localRepo := filepath.Join(filePathSeparator, "path", "to", "local", "repo")
	caCertPath := filepath.Join(filePathSeparator, "path", "to", "ca.crt")
	caCertPath2 := filepath.Join(filePathSeparator, "path", "2", "to", "ca.crt")
	caCertPath3 := filepath.Join(filePathSeparator, "path", "3", "to", "ca.crt")
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
			ShouldValidate:    false,
			ExpectFieldErrors: validation.ErrMultipleSources(commands.LocalPathAndSource, flags.ImageFlagName, flags.GitFlagWildcard),
		},
		{
			Name: "all sources including maven",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				GitRepo:       "https://example.com/repo.git",
				GitBranch:     "main",
				SourceImage:   "repo.example/image:tag",
				Image:         "repo.example/image:tag",
				MavenArtifact: "hello-world",
				MavenType:     "jar",
				MavenVersion:  "0.0.1",
			},
			ShouldValidate:    false,
			ExpectFieldErrors: validation.ErrMultipleSources(commands.MavenFlagWildcard, commands.LocalPathAndSource, flags.ImageFlagName, flags.GitFlagWildcard),
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
		{
			Name: "param yaml",
			Validatable: &commands.WorkloadOptions{
				Namespace:  "default",
				Name:       "my-resource",
				ParamsYaml: []string{"ports_json={\"name\": \"smtp\", \"port\": 1026}", "ports_nesting_yaml=- deployment:\n    name: smtp\n    port: 1026"},
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run",
			Validatable: &commands.WorkloadOptions{
				Namespace:  "default",
				Name:       "my-resource",
				ParamsYaml: []string{"ports_nesting_yaml=- deployment:\n    name: smtp\n    port: 1026", "ports_json={\"name\": \"smtp\", \"port\": 1026"},
			},
			ExpectFieldErrors: validation.ErrInvalidValue("ports_json={\"name\": \"smtp\", \"port\": 1026", flags.ParamYamlFlagName+"[1]"),
		},
		{
			Name: "registry username and pass",
			Validatable: &commands.WorkloadOptions{
				Namespace:        "default",
				Name:             "my-resource",
				RegistryUsername: "username",
				RegistryPassword: "password",
				SourceImage:      "repo.example/image:tag",
				LocalPath:        localRepo,
			},
			ShouldValidate: true,
		},
		{
			Name: "all registry opts",
			Validatable: &commands.WorkloadOptions{
				Namespace:        "default",
				Name:             "my-resource",
				RegistryUsername: "username",
				RegistryPassword: "password",
				CACertPaths:      []string{caCertPath},
				SourceImage:      "repo.example/image:tag",
				LocalPath:        localRepo,
			},
			ShouldValidate: true,
		},
		{
			Name: "ca cert",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				CACertPaths: []string{caCertPath},
				SourceImage: "repo.example/image:tag",
				LocalPath:   localRepo,
			},
			ShouldValidate: true,
		},
		{
			Name: "various ca certs",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				CACertPaths: []string{caCertPath, caCertPath2, caCertPath3},
				SourceImage: "repo.example/image:tag",
				LocalPath:   localRepo,
			},
			ShouldValidate: true,
		},
		{
			Name: "ca cert with no local path",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				CACertPaths: []string{caCertPath},
				SourceImage: "repo.example/image:tag",
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.LocalPathFlagName),
		},
		{
			Name: "ca cert with no source image",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				LocalPath:   localRepo,
				CACertPaths: []string{caCertPath},
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.SourceImageFlagName),
		},
		{
			Name: "ca cert with no local path and no source image",
			Validatable: &commands.WorkloadOptions{
				Namespace:   "default",
				Name:        "my-resource",
				CACertPaths: []string{caCertPath},
			},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(flags.SourceImageFlagName),
				validation.ErrMissingField(flags.LocalPathFlagName),
			),
		},
		{
			Name: "registry username and pass with no source image",
			Validatable: &commands.WorkloadOptions{
				Namespace:        "default",
				Name:             "my-resource",
				LocalPath:        localRepo,
				RegistryUsername: "username",
				RegistryPassword: "password",
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.SourceImageFlagName),
		},
		{
			Name: "registry username and pass with no local path",
			Validatable: &commands.WorkloadOptions{
				Namespace:        "default",
				Name:             "my-resource",
				RegistryUsername: "username",
				RegistryPassword: "password",
				SourceImage:      "repo.example/image:tag",
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.LocalPathFlagName),
		},
		{
			Name: "registry username and pass with no local path and no source image",
			Validatable: &commands.WorkloadOptions{
				Namespace:        "default",
				Name:             "my-resource",
				RegistryUsername: "username",
				RegistryPassword: "password",
			},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(flags.SourceImageFlagName),
				validation.ErrMissingField(flags.LocalPathFlagName),
			),
		},
		{
			Name: "registry token with no source image",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RegistryToken: "my-token",
				LocalPath:     localRepo,
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.SourceImageFlagName),
		},
		{
			Name: "registry token with no local path",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RegistryToken: "my-token",
				SourceImage:   "repo.example/image:tag",
			},
			ExpectFieldErrors: validation.ErrMissingField(flags.LocalPathFlagName),
		},
		{
			Name: "registry token with no source image and no local path",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RegistryToken: "my-token",
			},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(flags.SourceImageFlagName),
				validation.ErrMissingField(flags.LocalPathFlagName),
			),
		},
		{
			Name: "registry token",
			Validatable: &commands.WorkloadOptions{
				Namespace:     "default",
				Name:          "my-resource",
				RegistryToken: "my-token",
				SourceImage:   "repo.example/image:tag",
				LocalPath:     localRepo,
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

func TestOutputWorkload(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		input       *cartov1alpha1.Workload
		expected    string
		shouldError bool
	}{{
		name: "print output with yaml",
		args: []string{flags.OutputFlagName, printer.OutputFormatYaml},
		input: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
				UID:                        types.UID("uid-xyz"),
				ResourceVersion:            "999",
				Generation:                 1,
				CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
				Finalizers:                 []string{"my.finalizer"},
				DeletionGracePeriodSeconds: &[]int64{5}[0],
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "tanzu"},
				},
				SelfLink: "/default/my-workload",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "workload-owner",
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:    cartov1alpha1.WorkloadConditionReady,
						Status:  metav1.ConditionTrue,
						Reason:  "No printing status",
						Message: "a hopefully informative message about what went wrong",
						LastTransitionTime: metav1.Time{
							Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
						},
					},
				},
			},
		},
		expected: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  annotations:
    name: value
  creationTimestamp: "2021-09-10T15:00:00Z"
  deletionGracePeriodSeconds: 5
  deletionTimestamp: "2021-09-10T15:00:00Z"
  finalizers:
  - my.finalizer
  generation: 1
  labels:
    name: value
  name: my-workload
  namespace: default
  ownerReferences:
  - apiVersion: v1
    kind: Pod
    name: workload-owner
    uid: ""
  resourceVersion: "999"
  selfLink: /default/my-workload
  uid: uid-xyz
spec: {}
status:
  conditions:
  - lastTransitionTime: "2019-06-29T01:44:05Z"
    message: a hopefully informative message about what went wrong
    reason: No printing status
    status: "True"
    type: Ready
  supplyChainRef: {}
`,
	}, {
		name: "print output with json",
		args: []string{flags.OutputFlagName, printer.OutputFormatJson},
		input: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
				UID:                        types.UID("uid-xyz"),
				ResourceVersion:            "999",
				Generation:                 1,
				CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
				Finalizers:                 []string{"my.finalizer"},
				DeletionGracePeriodSeconds: &[]int64{5}[0],
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "tanzu"},
				},
				SelfLink: "/default/my-workload",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "workload-owner",
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:    cartov1alpha1.WorkloadConditionReady,
						Status:  metav1.ConditionTrue,
						Reason:  "No printing status",
						Message: "a hopefully informative message about what went wrong",
						LastTransitionTime: metav1.Time{
							Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
						},
					},
				},
			},
		},
		expected: `
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"annotations": {
			"name": "value"
		},
		"creationTimestamp": "2021-09-10T15:00:00Z",
		"deletionGracePeriodSeconds": 5,
		"deletionTimestamp": "2021-09-10T15:00:00Z",
		"finalizers": [
			"my.finalizer"
		],
		"generation": 1,
		"labels": {
			"name": "value"
		},
		"name": "my-workload",
		"namespace": "default",
		"ownerReferences": [
			{
				"apiVersion": "v1",
				"kind": "Pod",
				"name": "workload-owner",
				"uid": ""
			}
		],
		"resourceVersion": "999",
		"selfLink": "/default/my-workload",
		"uid": "uid-xyz"
	},
	"spec": {},
	"status": {
		"conditions": [
			{
				"lastTransitionTime": "2019-06-29T01:44:05Z",
				"message": "a hopefully informative message about what went wrong",
				"reason": "No printing status",
				"status": "True",
				"type": "Ready"
			}
		],
		"supplyChainRef": {}
	}
}
`,
	}, {
		name:        "not valid output",
		args:        []string{flags.OutputFlagName, "myBadFormat"},
		input:       &cartov1alpha1.Workload{},
		shouldError: true,
	}}

	for _, test := range tests {
		scheme := k8sruntime.NewScheme()
		_ = cartov1alpha1.AddToScheme(scheme)
		c := cli.NewDefaultConfig("test", scheme)
		output := &bytes.Buffer{}
		c.Stdout = output
		c.Stderr = output

		cmd := &cobra.Command{}
		ctx := cli.WithCommand(context.Background(), cmd)

		opts := &commands.WorkloadOptions{}
		opts.DefineFlags(ctx, c, cmd)
		cmd.ParseFlags(test.args)

		actual := test.input.DeepCopy()
		err := opts.OutputWorkload(c, actual)
		if err != nil && !test.shouldError {
			t.Errorf("OutputWorkload() errored %v", err)
		}
		if err == nil && test.shouldError {
			t.Errorf("OutputWorkload() expected error")
		}
		if test.shouldError {
			return
		}

		if diff := cmp.Diff(strings.TrimSpace(test.expected), strings.TrimSpace(output.String())); diff != "" {
			t.Errorf("OutputWorkload() (-want, +got) = %s", diff)
		}
	}
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

	scheme := k8sruntime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)
	c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

	tests := []struct {
		name     string
		args     []string
		current  *cartov1alpha1.Workload
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
						"FOO":                                 "foo",
						"BAR":                                 "bar",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"NEW":                                 "value",
						"FOO":                                 "bar",
						"apps.tanzu.vmware.com/workload-type": "web",
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
						"FOO":                                 "foo",
						"BAR":                                 "bar",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"NEW":                                 "value",
						"FOO":                                 "bar",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
			},
		},
		{
			name: "workload with labels add/delete",
			args: []string{flags.GitRepoFlagName, gitRepo, flags.GitBranchFlagName, gitBranch, flags.LabelFlagName, "app.kubernetes.io/part-of=my-app", flags.LabelFlagName, "app.kubernetes.io/part-of-"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
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
		{
			name: "add/update annotation",
			args: []string{flags.AnnotationFlagName, "NEW=value", flags.AnnotationFlagName, "FOO=bar", flags.AnnotationFlagName, "removeme-"},
			input: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
			name: "add maven with param yaml",
			args: []string{flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples"}`},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
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
		{
			name: "add maven with flags",
			args: []string{flags.MavenArtifactFlagName, "spring-petclinic", flags.MavenVersionFlagName, "2.6.0", flags.MavenGroupFlagName, "org.springframework.samples"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
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
		{
			name: "update params",
			args: []string{flags.ParamFlagName, "foo=bar", flags.ParamFlagName, "removeme-"},
			input: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
						apis.AppPartOfLabelName:    appName,
						apis.WorkloadTypeLabelName: "web",
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
			name: "workload debug deactivated",
			args: []string{fmt.Sprintf("%s=%s", flags.DebugFlagName, "false")},
			input: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
						apis.WorkloadTypeLabelName: "web",
					},
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
			name: "workload with live-update deactivated",
			args: []string{fmt.Sprintf("%s=%s", flags.LiveUpdateFlagName, "false")},
			input: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
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
						Subpath: "./cmd",
					},
				},
			},
		},
		{
			name: "set git flags to empty",
			args: []string{flags.GitBranchFlagName, "", flags.GitCommitFlagName, ""},
			input: &cartov1alpha1.Workload{
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
								Tag:    gitTag,
								Commit: gitCommit,
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
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: gitRepo,
							Ref: cartov1alpha1.GitRef{
								Tag: gitTag,
							},
						},
					},
				},
			},
		},
		{
			name: "set git repo to empty",
			args: []string{flags.GitRepoFlagName, ""},
			input: &cartov1alpha1.Workload{
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
								Tag:    gitTag,
								Commit: gitCommit,
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
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{},
			},
		},
		{
			name: "subPath update with image source",
			args: []string{flags.SubPathFlagName, subPath},
			input: &cartov1alpha1.Workload{
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
			expected: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
						apis.AppPartOfLabelName:    workloadName,
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
						apis.AppPartOfLabelName:    workloadName,
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
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"cache":{"namespace":"my-cache-ns"}}}}`,
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
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
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
						apis.AppPartOfLabelName:    workloadName,
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
						apis.AppPartOfLabelName:    workloadName,
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
			name: "update service reference without namespaces",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-ns:my-prod-db"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
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
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
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
			name: "remove service references namespaces",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-db"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-delete-me"}}}}`,
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
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
					},
					Annotations: map[string]string{},
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
			name: "remove service references namespaces from multiple services",
			args: []string{flags.ServiceRefFlagName, "database=services.tanzu.vmware.com/v1alpha1:MySQL:my-prod-db"},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind": "ServiceClaimsExtension","apiVersion": "supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec": {"serviceClaims": {"database": {"namespace": "my-prod-delete-me","name": "my-prod-db"},"cache": {"namespace": "my-prod-cache-ns","name": "my-prod-cache"}}}}`,
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
						{
							Name: "cache",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "MySQL",
								Name:       "my-prod-cache",
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
						apis.AppPartOfLabelName:    workloadName,
						apis.WorkloadTypeLabelName: "web",
					},
					Annotations: map[string]string{
						apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"cache":{"name":"my-prod-cache","namespace":"my-prod-cache-ns"}}}}`,
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
						{
							Name: "cache",
							Ref: &cartov1alpha1.WorkloadServiceClaimReference{
								APIVersion: "services.tanzu.vmware.com/v1alpha1",
								Kind:       "MySQL",
								Name:       "my-prod-cache",
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
						apis.WorkloadTypeLabelName: "web",
					},
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
		{
			name: "workload from lsp",
			args: []string{flags.LocalPathFlagName, "path"},
			current: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
					Annotations: map[string]string{
						apis.LocalSourceProxyAnnotationName: "my-lsp-annotation",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: "any-image",
					},
				},
			},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: "local-source-proxy.tap-local-source-system.svc.cluster.local/source:default-my-workload",
					},
				},
			},
		},
		{
			name: "update workload with source image already defined",
			args: []string{flags.LocalPathFlagName, "path"},
			current: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: "any-image",
					},
				},
			},
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{},
			},
			expected: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: "any-image",
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
		opts.ApplyOptionsToWorkload(ctx, test.current, actual)
		t.Run(test.name, func(t *testing.T) {
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("ApplyOptionsToWorkload() (-want, +got) = %s", diff)
			}
		})
	}
}

func TestManageLocalSourceProxyAnnotation(t *testing.T) {
	sourceImage := "localsource/hello:source"
	tests := []struct {
		name             string
		args             []string
		currentWorkload  *cartov1alpha1.Workload
		workload         *cartov1alpha1.Workload
		expectedWorkload *cartov1alpha1.Workload
		fileWorkload     *cartov1alpha1.Workload
	}{{
		name: "add annotation",
		args: []string{flags.LocalPathFlagName, localSource},
		workload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
		},
	}, {
		name: "do not add annotation coming from file",
		args: []string{},
		fileWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
	}, {
		name: "do not remove annotation if workload in file does not contain source",
		args: []string{flags.LocalPathFlagName, localSource},
		fileWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"my-label": "my-image@sha:123abc",
				},
			},
		},
		currentWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
	}, {
		name: "remove annotation when using source image",
		args: []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage},
		currentWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{},
	}, {
		name: "remove annotation when switching to git",
		args: []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage},
		currentWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Git: &cartov1alpha1.GitSource{
						URL: "my-repo.git",
						Ref: cartov1alpha1.GitRef{
							Branch: "main",
						},
					},
				},
			},
		},
	}, {
		name: "remove annotation when switching to image",
		args: []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage},
		currentWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Image: "my-image",
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Image: "my-image",
			},
		},
	}, {
		name: "update annotation",
		args: []string{flags.LocalPathFlagName, localSource},
		currentWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-local-source-annotation",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:456efg",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:456efg",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:456efg",
				},
			},
		},
	}, {
		name: "do not change annotation when setting other flags",
		args: []string{flags.LocalPathFlagName, localSource, flags.TypeFlagName, "my-type"},
		currentWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		workload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
		expectedWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: "my-image@sha:123abc",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image@sha:123abc",
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := k8sruntime.NewScheme()
			c := cli.NewDefaultConfig("test", scheme)

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)
			opts := &commands.WorkloadOptions{}
			opts.LoadDefaults(c)
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)

			opts.ManageLocalSourceProxyAnnotation(test.fileWorkload, test.currentWorkload, test.workload)

			if test.expectedWorkload.IsAnnotationExists(apis.LocalSourceProxyAnnotationName) && !test.workload.IsAnnotationExists(apis.LocalSourceProxyAnnotationName) {
				t.Errorf("expected local source proxy annotation to exist in workload")
			}
			if !test.expectedWorkload.IsAnnotationExists(apis.LocalSourceProxyAnnotationName) && test.workload.IsAnnotationExists(apis.LocalSourceProxyAnnotationName) {
				t.Errorf("local source proxy annotation not expected to be in workload")
			}
			if test.expectedWorkload.IsAnnotationExists(apis.LocalSourceProxyAnnotationName) && test.workload.IsAnnotationExists(apis.LocalSourceProxyAnnotationName) {
				workloadAnnotation := test.workload.Annotations[apis.LocalSourceProxyAnnotationName]
				expectedAnnotation := test.expectedWorkload.Annotations[apis.LocalSourceProxyAnnotationName]

				if workloadAnnotation != expectedAnnotation {
					t.Errorf("workload annotation expected to be %q, got %q", expectedAnnotation, workloadAnnotation)
				}
			}
		})
	}
}

func TestWorkloadOptionsPublishLocalSourcePrivateRegistry(t *testing.T) {
	reg, err := ggcrregistry.TLS("localhost")
	utilruntime.Must(err)
	defer reg.Close()
	u, err := url.Parse(reg.URL)
	utilruntime.Must(err)
	registryHost := u.Host

	cert, err := os.CreateTemp("", "customCA")
	if err != nil {
		t.Fatalf("Unable to create temp dir %v", err)
	}
	defer os.RemoveAll(cert.Name())

	if err := pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: reg.Certificate().Raw}); err != nil {
		t.Fatalf("Unable to parse certificate %v", err)
	}

	sourceImage := fmt.Sprintf("%s/hello:source", registryHost)

	tests := []struct {
		name           string
		args           []string
		shouldPrint    bool
		input          string
		expected       string
		shouldError    bool
		expectedOutput string
	}{{
		name:        "local source to private registry",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.RegistryCertFlagName, cert.Name(), flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "local source to private registry with username and pass",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.RegistryCertFlagName, cert.Name(), flags.RegistryUsernameFlagName, "admin", flags.RegistryPasswordFlagName, "password", flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "local source to private registry with token",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.RegistryCertFlagName, cert.Name(), flags.RegistryTokenFlagName, "myToken123", flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:           "local source to private registry without prompts",
		args:           []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.RegistryCertFlagName, cert.Name(), flags.YesFlagName},
		input:          fmt.Sprintf("%s/hello:source", registryHost),
		expected:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: "",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := k8sruntime.NewScheme()
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)
			ctx = logger.StashSourceImageLogger(ctx, logger.NewNoopLogger())
			opts := &commands.WorkloadOptions{}
			opts.LoadDefaults(c)
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)

			ctx = logger.StashProgressBarLogger(ctx, fake.NewNoopProgressBar())
			workload := &cartov1alpha1.Workload{
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: test.input,
					},
				},
			}

			err := opts.PublishLocalSource(ctx, c, nil, workload, test.shouldPrint)
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

func TestWorkloadOptionsPublishLocalSource(t *testing.T) {
	reg, err := ggcrregistry.TLS("localhost")
	utilruntime.Must(err)
	defer reg.Close()
	u, err := url.Parse(reg.URL)
	utilruntime.Must(err)
	registryHost := u.Host
	helloJarFilePath := filepath.Join("testdata", "hello.go.jar")
	expectedImageDigest := "fedc574423e7aa2ecdd2ffb3381214e3c288db871ab9a3758f77489d6a777a1d"
	if runtime.GOOS == "windows" {
		expectedImageDigest = "4b931bb7ef0a7780a3fc58364aaf8634cf4885af6359fb692461f0247c8a9f34"
	}

	sourceImage := fmt.Sprintf("%s/hello:source", registryHost)

	tests := []struct {
		name             string
		args             []string
		input            string
		expected         string
		shouldError      bool
		shouldPrint      bool
		expectedOutput   string
		skip             bool
		existingWorkload *cartov1alpha1.Workload
	}{{
		name:        "local source with excluded files",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, filepath.Join("testdata", "local-source-exclude-files"), flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, expectedImageDigest),
		expectedOutput: `
The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
Publishing source in ` + fmt.Sprintf("%q", filepath.Join("testdata", "local-source-exclude-files")) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "local source include tanzu ignore with windows path",
		shouldPrint: true,
		skip:        runtime.GOOS != "windows",
		args:        []string{flags.LocalPathFlagName, filepath.Join("testdata", "local-source-exclude-files-windows"), flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "8ce661d3fc7f94de72d76ec32f3ab6befc159fc263977e5b80564bf9e97a4509"),
		expectedOutput: `
The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
Publishing source in ` + fmt.Sprintf("%q", filepath.Join("testdata", "local-source-exclude-files-windows")) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "local source",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "jar file",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "invalid file",
		args:        []string{flags.LocalPathFlagName, filepath.Join("testdata", "invalid.zip"), flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", registryHost),
		shouldError: true,
	}, {
		name:        "with digest",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "0000000000000000000000000000000000000000000000000000000000000000"),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "when workload already has resolved image with digest",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "0000000000000000000000000000000000000000000000000000000000000000"),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
				},
			},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + registryHost + `/hello:source"...
No source code is changed
`,
	}, {
		name:        "when workload already has resolved image with digest and no source",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "0000000000000000000000000000000000000000000000000000000000000000"),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "update workload created in public repo when changing local path and not using source image",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
				},
			},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "from git source to source image",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName, flags.SourceImageFlagName, sourceImage},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Git: &cartov1alpha1.GitSource{
						Ref: cartov1alpha1.GitRef{
							Branch: "main",
						},
						URL: "my-repo.git",
					},
				},
			},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:        "from image to source image",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName, flags.SourceImageFlagName, sourceImage},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
		expected:    fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Image: "my-image",
			},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + registryHost + `/hello:source"...
📥 Published source
`,
	}, {
		name:           "local source without prompts",
		args:           []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:          fmt.Sprintf("%s/hello:source", registryHost),
		expected:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: "",
	}, {
		name:           "jar file without prompts",
		args:           []string{flags.LocalPathFlagName, helloJarFilePath, flags.SourceImageFlagName, sourceImage, flags.YesFlagName},
		input:          fmt.Sprintf("%s/hello:source", registryHost),
		expected:       fmt.Sprintf("%s/hello:source@sha256:%s", registryHost, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		expectedOutput: "",
	}, {
		name:           "no local path",
		args:           []string{},
		input:          fmt.Sprintf("%s/hello:source", registryHost),
		expected:       fmt.Sprintf("%s/hello:source", registryHost),
		expectedOutput: "",
	}, {
		name:        "publish local source with error",
		args:        []string{flags.LocalPathFlagName, localSource, flags.SourceImageFlagName, "a", flags.YesFlagName},
		input:       "a",
		shouldError: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			scheme := k8sruntime.NewScheme()
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)
			ctx = source.StashContainerRemoteTransport(ctx, reg.Client().Transport)
			ctx = logger.StashSourceImageLogger(ctx, logger.NewNoopLogger())
			opts := &commands.WorkloadOptions{}
			opts.LoadDefaults(c)
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)

			ctx = logger.StashProgressBarLogger(ctx, fake.NewNoopProgressBar())
			workload := &cartov1alpha1.Workload{
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: test.input,
					},
				},
			}

			err := opts.PublishLocalSource(ctx, c, test.existingWorkload, workload, test.shouldPrint)
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

func TestWorkloadOptionsPublishLocalSourceProxy(t *testing.T) {
	expectedImageDigest := "fedc574423e7aa2ecdd2ffb3381214e3c288db871ab9a3758f77489d6a777a1d"
	if runtime.GOOS == "windows" {
		expectedImageDigest = "4b931bb7ef0a7780a3fc58364aaf8634cf4885af6359fb692461f0247c8a9f34"
	}
	workload := &cartov1alpha1.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	fakeWrapper := fake_source.GetFakeWrapper(http.Header{
		"Content-Type":          []string{"text/html", "application/json", "application/octet-stream"},
		"Docker-Content-Digest": []string{"sha256:" + expectedImageDigest},
	})
	if fakeWrapper.Repository == "" {
		fakeWrapper.Repository = "my-test-repo"
	}

	helloJarFilePath := filepath.Join("testdata", "hello.go.jar")

	tests := []struct {
		name             string
		args             []string
		input            string
		expected         string
		shouldError      bool
		shouldPrint      bool
		expectedOutput   string
		skip             bool
		existingWorkload *cartov1alpha1.Workload
	}{{
		name:        "local source with excluded files",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, filepath.Join("testdata", "local-source-exclude-files"), flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expected:    fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, expectedImageDigest),
		expectedOutput: `
The files and/or directories listed in the .tanzuignore file are being excluded from the uploaded source code.
Publishing source in ` + fmt.Sprintf("%q", filepath.Join("testdata", "local-source-exclude-files")) + ` to "` + source.GetLocalImageRepo() + `/` + source.ImageTag + `:` + workload.Namespace + `-` + workload.Name + `"...
📥 Published source
`,
	}, {
		name:        "local source",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expected:    fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + source.GetLocalImageRepo() + `/` + source.ImageTag + `:` + workload.Namespace + `-` + workload.Name + `"...
📥 Published source
`,
	}, {
		name:        "jar file",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expected:    fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + source.GetLocalImageRepo() + `/` + source.ImageTag + `:` + workload.Namespace + `-` + workload.Name + `"...
📥 Published source
`,
	}, {
		name:        "invalid file",
		args:        []string{flags.LocalPathFlagName, filepath.Join("testdata", "invalid.zip"), flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		shouldError: true,
	}, {
		name:        "with digest",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, localSource, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", source.GetLocalImageRepo(), "0000000000000000000000000000000000000000000000000000000000000000"),
		expected:    fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", localSource) + ` to "` + source.GetLocalImageRepo() + `/` + source.ImageTag + `:` + workload.Namespace + `-` + workload.Name + `"...
📥 Published source
`,
	}, {
		name:        "when workload already has resolved image with digest",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", source.GetLocalImageRepo(), "0000000000000000000000000000000000000000000000000000000000000000"),
		expected:    fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
				},
			},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + source.GetLocalImageRepo() + `/` + source.ImageTag + `:` + workload.Namespace + `-` + workload.Name + `"...
No source code is changed
`,
	}, {
		name:        "when workload already has resolved image with digest and no source",
		shouldPrint: true,
		args:        []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName},
		input:       fmt.Sprintf("%s/hello:source@sha256:%s", source.GetLocalImageRepo(), "0000000000000000000000000000000000000000000000000000000000000000"),
		expected:    fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		existingWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.LocalSourceProxyAnnotationName: fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "111d543b7736846f502387eed53be08c5ceb0a6010faaaf043409702074cf652"),
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{},
		},
		expectedOutput: `
Publishing source in ` + fmt.Sprintf("%q", helloJarFilePath) + ` to "` + source.GetLocalImageRepo() + `/` + source.ImageTag + `:` + workload.Namespace + `-` + workload.Name + `"...
📥 Published source
`,
	}, {
		name:           "local source without prompts",
		args:           []string{flags.LocalPathFlagName, localSource, flags.YesFlagName},
		input:          fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expected:       fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "978be33a7f0cbe89bf48fbb438846047a28e1298d6d10d0de2d64bdc102a9e69"),
		expectedOutput: "",
	}, {
		name:           "jar file without prompts",
		args:           []string{flags.LocalPathFlagName, helloJarFilePath, flags.YesFlagName},
		input:          fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expected:       fmt.Sprintf("%s:%s-%s@sha256:%s", fakeWrapper.Repository, workload.Namespace, workload.Name, "f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be"),
		expectedOutput: "",
	}, {
		name:           "no local path",
		args:           []string{},
		input:          fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expected:       fmt.Sprintf("%s/hello:source", source.GetLocalImageRepo()),
		expectedOutput: "",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			scheme := k8sruntime.NewScheme()
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)
			ctx = source.StashContainerWrapper(ctx, fakeWrapper)
			ctx = logger.StashSourceImageLogger(ctx, logger.NewNoopLogger())
			opts := &commands.WorkloadOptions{}
			opts.LoadDefaults(c)
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)

			ctx = logger.StashProgressBarLogger(ctx, fake.NewNoopProgressBar())

			workload.Annotations = map[string]string{
				apis.LocalSourceProxyAnnotationName: test.input,
			}
			workload.Spec = cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: test.input,
				},
			}

			err := opts.PublishLocalSource(ctx, c, test.existingWorkload, workload, test.shouldPrint)
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
		noColor        bool
		expectedOutput string
		withReactors   []clitesting.ReactionFunc
	}{
		{
			name: "Create workload successfully",
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu:bionic",
				},
			},
			shouldError: false,
			expectedOutput: `
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
     10 + |  image: ubuntu:bionic
👍 Created workload "my-workload"`,
		},
		{
			name:    "Create workload with no color",
			noColor: true,
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec:
     10 + |  image: ubuntu:bionic
Created workload "my-workload"`,
		},
		{
			name: "Create workload from maven",
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
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
			shouldError: false,
			expectedOutput: `
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
👍 Created workload "my-workload"`,
		},
		{
			name: "Create workload without source successfully",
			input: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
				},
			},
			shouldError: false,
			expectedOutput: `
🔎 Create workload:
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  labels:
      6 + |    apps.tanzu.vmware.com/workload-type: web
      7 + |  name: my-workload
      8 + |  namespace: default
      9 + |spec: {}
❗ NOTICE: no source code or image has been specified for this workload.
👍 Created workload "my-workload"`,
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
			scheme := k8sruntime.NewScheme()
			_ = cartov1alpha1.AddToScheme(scheme)
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			c.NoColor = test.noColor
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

	scheme := k8sruntime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)
	c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

	tests := []struct {
		name           string
		args           []string
		givenWorkload  *cartov1alpha1.Workload
		shouldError    bool
		expectedOutput string
		withReactors   []clitesting.ReactionFunc
		noColor        bool
	}{
		{
			name:    "Update workload successfully with no emojis",
			noColor: true,
			args:    []string{flags.LabelFlagName, "NEW=value", flags.YesFlagName},
			givenWorkload: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"FOO": "bar",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu",
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
  9, 10   |spec:
 10, 11   |  image: ubuntu
Updated workload "my-workload"
`,
		},
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
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu",
				},
			},
			shouldError: false,
			expectedOutput: `
🔎 Update workload:
...
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6,  6   |    FOO: bar
      7 + |    NEW: value
  7,  8   |  name: my-workload
  8,  9   |  namespace: default
  9, 10   |spec:
 10, 11   |  image: ubuntu
👍 Updated workload "my-workload"
`,
		},
		{
			name: "Update workload without source successfully",
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
🔎 Update workload:
...
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6,  6   |    FOO: bar
      7 + |    NEW: value
  7,  8   |  name: my-workload
  8,  9   |  namespace: default
  9, 10   |spec: {}
❗ NOTICE: no source code or image has been specified for this workload.
👍 Updated workload "my-workload"
`,
		},
		{
			name: "Update workload to add maven param",
			args: []string{flags.ParamYamlFlagName, `maven={"artifactId": "spring-petclinic", "version": "2.6.0", "groupId": "org.springframework.samples"}`, flags.YesFlagName},
			givenWorkload: &cartov1alpha1.Workload{
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
							Name:  "removeme",
							Value: apiextensionsv1.JSON{Raw: []byte(`"bye"`)},
						},
					},
				},
			},
			shouldError: false,
			expectedOutput: `
🔎 Update workload:
...
  9,  9   |spec:
 10, 10   |  params:
 11, 11   |  - name: removeme
 12, 12   |    value: bye
     13 + |  - name: maven
     14 + |    value:
     15 + |      artifactId: spring-petclinic
     16 + |      groupId: org.springframework.samples
     17 + |      version: 2.6.0
👍 Updated workload "my-workload"`,
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
					Labels: map[string]string{
						apis.WorkloadTypeLabelName: "web",
					},
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
			scheme := k8sruntime.NewScheme()
			_ = cartov1alpha1.AddToScheme(scheme)
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			c.NoColor = test.noColor
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
			opts.ApplyOptionsToWorkload(ctx, currentWorkload, workload)
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
	scheme := k8sruntime.NewScheme()
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
			name:  "loads workload from url",
			file:  "https://raw.githubusercontent.com/vmware-tanzu/apps-cli-plugin/main/pkg/commands/testdata/workload.yaml",
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
			name:        "error loading non-accepted url file",
			file:        "ftp://raw.githubusercontent.com/vmware-tanzu/apps-cli-plugin/main/pkg/commands/testdata/workload.yaml",
			stdin:       c.Stdin,
			shouldError: true,
		},
		{
			name:        "error loading non-workload file",
			file:        "https://raw.githubusercontent.com/vmware-tanzu/apps-cli-plugin/main/testing/e2e/testdata/prereq/cluster-supply-chain.yaml",
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
