//go:build integration
// +build integration

/*
Copyright 2022 VMware, Inc.

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

package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	it "github.com/vmware-tanzu/apps-cli-plugin/testing/suite"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespaceFlag    = "--namespace=" + it.TestingNamespace
	workloadTypeMeta = metav1.TypeMeta{
		Kind:       cartov1alpha1.WorkloadKind,
		APIVersion: cartov1alpha1.SchemeGroupVersion.String(),
	}
)

func TestCreateFromGitWithAnnotations(t *testing.T) {
	testSuite := it.CommandLineIntegrationTestSuite{
		{
			Name:         "Create workload with valid name from a git repo and annotations",
			WorkloadName: "test-create-git-annotations-workload",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-git-annotations-workload",
					"--annotation=min-instances=2", "--annotation=max-instances=4",
					"--app=test-create-git-annotations-workload",
					"--git-tag=tap-1.2",
					"--git-repo=https://github.com/sample-accelerators/spring-petclinic",
					namespaceFlag,
					"--type=web",
					"--yes",
				)
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-git-annotations-workload",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of":           "test-create-git-annotations-workload",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{{
						Name: "annotations",
						Value: v1.JSON{
							Raw: []byte(`{"max-instances":"4","min-instances":"2"}`),
						},
					}},
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
		{
			Name:         "Create workload with maven flags",
			WorkloadName: "test-create-git-annotations-workload",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-maven-workload",
					"--maven-artifact=spring-petclinic",
					"--maven-version=v2.6.0",
					"--maven-group=org.springframework.samples",
					namespaceFlag,
					"--type=web",
					"--yes",
				)
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-maven-workload",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{{
						Name: "maven",
						Value: v1.JSON{
							Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","type":null,"version":"v2.6.0"}`),
						},
					}},
				},
			},
		},
		{
			Name:         "Create workload with valid name from local source code",
			WorkloadName: "test-create-local-registry",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-local-registry",
					"--local-path=./testdata/hello.go.jar",
					"--source-image", os.Getenv("BUNDLE"),
					namespaceFlag,
					"--type=web",
					"--yes",
				)
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-local-registry",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: fmt.Sprintf("%v@sha256:f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be", os.Getenv("BUNDLE")),
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if _, err := exec.LookPath("imgpkg"); err != nil {
					t.Errorf("expected imgpkg in PATH: %v", err)
					t.FailNow()
				}
				dir, _ := ioutil.TempDir("", "")
				defer os.RemoveAll(dir)

				if err := it.NewCommandLine("imgpkg", "pull", "-i", os.Getenv("BUNDLE"), "-o", dir).Exec(); err != nil {
					t.Errorf("unexpected error %v ", err)
					t.FailNow()
				}
				// compare files
				got, err := os.ReadFile(filepath.Join(dir, "hello.go"))
				if err != nil {
					t.Errorf("unexpected error reading file %v ", err)
					t.FailNow()
				}

				excepted, err := os.ReadFile("./testdata/hello.go")
				if err != nil {
					t.Errorf("unexpected error reading file %v ", err)
					t.FailNow()
				}

				if diff := cmp.Diff(string(excepted), string(got)); diff != "" {
					t.Errorf("(-expected, +actual)\n%s", diff)
					t.FailNow()
				}
			},
		},
		{
			Name:         "Create workload with valid name from local source code values from environment variables",
			WorkloadName: "test-create-local-registry-venv",
			Command: func() it.CommandLine {
				os.Setenv("TANZU_APPS_SOURCE_IMAGE", os.Getenv("BUNDLE"))
				c := *it.NewTanzuAppsCommandLine(
					"workload", "test-create-local-registry-venv",
					"--local-path=./testdata/hello.go.jar",
					namespaceFlag,
					"--yes",
				)
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-local-registry-venv",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: fmt.Sprintf("%v@sha256:f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be", os.Getenv("BUNDLE")),
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if _, err := exec.LookPath("imgpkg"); err != nil {
					t.Errorf("expected imgpkg in PATH: %v", err)
					t.FailNow()
				}
				dir, _ := ioutil.TempDir("", "")
				defer os.RemoveAll(dir)

				if err := it.NewCommandLine("imgpkg", "pull", "-i", os.Getenv("BUNDLE"), "-o", dir).Exec(); err != nil {
					t.Errorf("unexpected error %v ", err)
					t.FailNow()
				}
				// compare files
				got, err := os.ReadFile(filepath.Join(dir, "hello.go"))
				if err != nil {
					t.Errorf("unexpected error reading file %v ", err)
					t.FailNow()
				}

				excepted, err := os.ReadFile("./testdata/hello.go")
				if err != nil {
					t.Errorf("unexpected error reading file %v ", err)
					t.FailNow()
				}

				if diff := cmp.Diff(string(excepted), string(got)); diff != "" {
					t.Errorf("(-expected, +actual)\n%s", diff)
					t.FailNow()
				}
			},
		},
		{
			Name:         "Update the created workload",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "apply", "test-create-git-annotations-workload", namespaceFlag, "--annotation=min-instances=3", "--annotation=max-instances=5", "-y"),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-git-annotations-workload",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of":           "test-create-git-annotations-workload",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{{
						Name: "annotations",
						Value: v1.JSON{
							Raw: []byte(`{"max-instances":"5","min-instances":"3"}`),
						},
					}},
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
		{
			Name:         "Get the updated workload",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "get", "test-create-git-annotations-workload", namespaceFlag),
		},
		{
			Name:         "Delete the created workload",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "delete", "test-create-git-annotations-workload", namespaceFlag, "-y"),
		},
		{
			Name: "Get the deleted workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "get", "test-create-git-annotations-workload", namespaceFlag),
			ShouldError: true,
		},
	}
	testSuite.Run(t)
}
