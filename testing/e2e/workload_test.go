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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/creack/pty"
	"github.com/google/go-cmp/cmp"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	it "github.com/vmware-tanzu/apps-cli-plugin/testing/suite"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespaceFlag    = "--namespace=" + it.TestingNamespace
	workloadTypeMeta = metav1.TypeMeta{
		Kind:       cartov1alpha1.WorkloadKind,
		APIVersion: cartov1alpha1.SchemeGroupVersion.String(),
	}
)

var (
	regexPod           = regexp.MustCompile("\\s*NAME\\s+READY\\s+STATUS\\s+RESTARTS\\s+AGE\\s*")
	regexProgress      = regexp.MustCompile(`[0-9]*\.*[0-9]*\s[a-zA-Z]+\s/\s[0-9]*[\.]*[0-9]*\s[a-zA-Z]+\s\[-+>*_*\]\s[0-9]*\.[0-9]+%.*`)
	serviceAccountName = "my-service-account"
)

func TestCreateFromGitWithAnnotations(t *testing.T) {
	testSuite := it.CommandLineIntegrationTestSuite{
		{
			Name:         "Create workload and show emojis",
			Skip:         runtime.GOOS == "windows",
			WorkloadName: "test-create-workload-show-emojis",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-workload-show-emojis",
					"--annotation=min-instances=2", "--annotation=max-instances=4",
					"--app=test-create-workload-show-emojis",
					"--git-tag=tap-1.2",
					"--git-repo=https://github.com/sample-accelerators/spring-petclinic",
					namespaceFlag,
					"--type=web",
				)
				c.SurveyAnswer("y")
				return c
			}(),
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				var err error
				if ctx, err = createPtyTerminal(ctx); err != nil {
					t.Fatalf("error while opening pty %v", err)
				}

				return ctx, nil
			},
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-workload-show-emojis",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of":           "test-create-workload-show-emojis",
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
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if !it.EmojisExistInOutput(output, it.ApplyEmojis) {
					t.Fatalf("expected emojis in create output %v", output)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				os.Stdout = ctx.Value("stdout").(*os.File)
				return nil
			},
		},
		{
			Name:         "Create workload with valid name from url filepath",
			WorkloadName: "spring-petclinic",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "spring-petclinic",
					"--file=https://raw.githubusercontent.com/vmware-tanzu/apps-cli-plugin/main/pkg/commands/testdata/workload.yaml",
					namespaceFlag,
					"--type=web",
				)
				c.SurveyAnswer("y")
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spring-petclinic",
					Namespace: it.TestingNamespace,
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
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: "https://github.com/spring-projects/spring-petclinic.git",
							Ref: cartov1alpha1.GitRef{
								Branch: "main",
							},
						},
					},
				},
			},
		},
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
				)
				c.SurveyAnswer("y")
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
							Raw: []byte(`{"artifactId":"spring-petclinic","groupId":"org.springframework.samples","version":"v2.6.0"}`),
						},
					}},
				},
			},
		},
		{
			Name:         "Create workload with valid name from local source code",
			WorkloadName: "test-create-local-registry",
			RequireEnvs:  []string{"BUNDLE"},
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
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				if v, ok := os.LookupEnv("CERT_DIR"); ok {
					if err := it.NewCommandLine("sudo", "cp", v+"/ca.pem", "/usr/local/share/ca-certificates/ca.crt").Exec(); err != nil {
						t.Fatalf("unexpected error while trying to copy registry certs %v ", err)
					}
					if err := it.NewCommandLine("sudo", "update-ca-certificates").Exec(); err != nil {
						t.Fatalf("unexpected error while trying to update registry certs %v ", err)
					}
				}
				return ctx, nil
			},
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
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if _, err := exec.LookPath("imgpkg"); err != nil {
					t.Fatalf("expected imgpkg in PATH: %v", err)
				}
				dir, _ := ioutil.TempDir("", "")
				defer os.RemoveAll(dir)

				if regexProgress.FindString(output) == "" {
					t.Fatalf("expected progressbar in output %v", output)
				}
				if err := it.NewCommandLine("imgpkg", "pull", "-i", os.Getenv("BUNDLE"), "-o", dir).Exec(); err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				// compare files
				got, err := os.ReadFile(filepath.Join(dir, "hello.go"))
				if err != nil {
					t.Fatalf("unexpected error reading file %v ", err)
				}

				excepted, err := os.ReadFile("./testdata/hello.go")
				if err != nil {
					t.Fatalf("unexpected error reading file %v ", err)
				}

				if diff := cmp.Diff(string(excepted), string(got)); diff != "" {
					t.Fatalf("(-expected, +actual)\n%s", diff)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				if _, ok := os.LookupEnv("CERT_DIR"); ok {
					if err := it.NewCommandLine("sudo", "rm", "-f", "/usr/local/share/ca-certificates/ca.crt").Exec(); err != nil {
						t.Fatalf("unexpected error while trying to delete registry certs %v ", err)
					}
					if err := it.NewCommandLine("sudo", "update-ca-certificates").Exec(); err != nil {
						t.Fatalf("unexpected error while trying to update registry certs %v ", err)
					}
				}
				return nil
			},
		},
		{
			Name:         "Create workload with valid name from local source code with private repo",
			WorkloadName: "test-create-local-registry-priv",
			RequireEnvs:  []string{"BUNDLE", "REGISTRY_USERNAME", "REGISTRY_PASSWORD", "CERT_DIR"},
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-local-registry-priv",
					"--local-path=./testdata/hello.go.jar",
					"--source-image", os.Getenv("BUNDLE"),
					"--registry-username", os.Getenv("REGISTRY_USERNAME"),
					"--registry-password", os.Getenv("REGISTRY_PASSWORD"),
					"--registry-ca-cert", os.Getenv("CERT_DIR")+"/ca.pem",
					namespaceFlag,
					"--type=web",
					"--yes",
				)
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-local-registry-priv",
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
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if _, err := exec.LookPath("imgpkg"); err != nil {
					t.Fatalf("expected imgpkg in PATH: %v", err)
				}
				dir, _ := ioutil.TempDir("", "")
				defer os.RemoveAll(dir)

				if regexProgress.FindString(output) == "" {
					t.Fatalf("expected progressbar in output %v", output)
				}
				ic := it.NewCommandLine("imgpkg", "pull", "--registry-ca-cert-path", os.Getenv("CERT_DIR")+"/ca.pem", "-i", os.Getenv("BUNDLE"), "-o", dir)
				ic.AddEnvVars(
					"IMGPKG_USERNAME="+os.Getenv("REGISTRY_USERNAME"),
					"IMGPKG_PASSWORD="+os.Getenv("REGISTRY_PASSWORD"),
				)

				if err := ic.Exec(); err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				// compare files
				got, err := os.ReadFile(filepath.Join(dir, "hello.go"))
				if err != nil {
					t.Fatalf("unexpected error reading file %v ", err)
				}

				excepted, err := os.ReadFile("./testdata/hello.go")
				if err != nil {
					t.Fatalf("unexpected error reading file %v ", err)
				}

				if diff := cmp.Diff(string(excepted), string(got)); diff != "" {
					t.Fatalf("(-expected, +actual)\n%s", diff)
				}
			},
		},
		{
			Name:         "Create workload with valid name from local source code values from environment variables with private repo",
			WorkloadName: "test-create-local-registry-venv",
			RequireEnvs:  []string{"BUNDLE", "REGISTRY_USERNAME", "REGISTRY_PASSWORD", "CERT_DIR"},
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-local-registry-venv",
					"--local-path=./testdata/hello.go.jar",
					namespaceFlag,
					"--source-image", os.Getenv("BUNDLE")+"-env",
					"--yes",
				)
				c.AddEnvVars(
					"TANZU_APPS_TYPE=web",
					"TANZU_APPS_REGISTRY_USERNAME="+os.Getenv("REGISTRY_USERNAME"),
					"TANZU_APPS_REGISTRY_PASSWORD="+os.Getenv("REGISTRY_PASSWORD"),
					"TANZU_APPS_REGISTRY_CA_CERT="+os.Getenv("CERT_DIR")+"/ca.pem",
				)
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-local-registry-venv",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Source: &cartov1alpha1.Source{
						Image: fmt.Sprintf("%v@sha256:f8a4db186af07dbc720730ebb71a07bf5e9407edc150eb22c1aa915af4f242be", os.Getenv("BUNDLE")+"-env"),
					},
				},
			},
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if _, err := exec.LookPath("imgpkg"); err != nil {
					t.Fatalf("expected imgpkg in PATH: %v", err)
				}
				dir, _ := ioutil.TempDir("", "")
				defer os.RemoveAll(dir)

				if regexProgress.FindString(output) == "" {
					t.Fatalf("expected progressbar in output %v", output)
				}
				ic := it.NewCommandLine("imgpkg", "pull", "--registry-ca-cert-path", os.Getenv("CERT_DIR")+"/ca.pem", "-i", os.Getenv("BUNDLE")+"-env", "-o", dir)
				ic.AddEnvVars(
					"IMGPKG_USERNAME="+os.Getenv("REGISTRY_USERNAME"),
					"IMGPKG_PASSWORD="+os.Getenv("REGISTRY_PASSWORD"),
				)

				if err := ic.Exec(); err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				// compare files
				got, err := os.ReadFile(filepath.Join(dir, "hello.go"))
				if err != nil {
					t.Fatalf("unexpected error reading file %v ", err)
				}

				excepted, err := os.ReadFile("./testdata/hello.go")
				if err != nil {
					t.Fatalf("unexpected error reading file %v ", err)
				}

				if diff := cmp.Diff(string(excepted), string(got)); diff != "" {
					t.Fatalf("(-expected, +actual)\n%s", diff)
				}
			},
		},
		{
			Name:         "Update the created workload and show emojis",
			Skip:         runtime.GOOS == "windows",
			WorkloadName: "test-create-workload-show-emojis",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "apply", "test-create-workload-show-emojis", namespaceFlag, "--annotation=min-instances=3", "--annotation=max-instances=5", "-y"),
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				var err error
				if ctx, err = createPtyTerminal(ctx); err != nil {
					t.Fatalf("error while opening pty %v", err)
				}

				return ctx, nil
			},
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-workload-show-emojis",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of":           "test-create-workload-show-emojis",
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
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if !it.EmojisExistInOutput(output, it.ApplyEmojis) {
					t.Fatalf("expected emojis in update output %v", output)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				os.Stdout = ctx.Value("stdout").(*os.File)
				return nil
			},
		},
		{
			Name:         "Update the created workload with `replace` update strategy",
			WorkloadName: "spring-petclinic",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "apply", "spring-petclinic",
					"--file=testdata/replace-workload.yaml",
					namespaceFlag,
					"--annotation=min-instances=3",
					"--annotation=max-instances=5",
					"--request-memory=2Gi",
					"--limit-cpu=400m",
					"--type=my-type",
					"--update-strategy=replace",
				)
				c.SurveyAnswer("y")
				return c
			}(),
			ExpectedObject: &cartov1alpha1.Workload{
				TypeMeta: workloadTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Namespace: it.TestingNamespace,
					Name:      "spring-petclinic",
					Labels: map[string]string{
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
							corev1.ResourceCPU: resource.MustParse("400m"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("2Gi"),
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
							Value: v1.JSON{Raw: []byte(`[{"image":"mysql:5.7","name":"mysql"},{"image":"postgres:9.6","name":"postgres"}]`)},
						},
						{
							Name: "annotations",
							Value: v1.JSON{
								Raw: []byte(`{"max-instances":"5","min-instances":"3"}`),
							},
						},
					},
				},
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
			Name:         "Get the updated workload with color",
			Skip:         runtime.GOOS == "windows",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "get", "test-create-git-annotations-workload", namespaceFlag),
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				var err error
				if ctx, err = createPtyTerminal(ctx); err != nil {
					t.Fatalf("error while opening pty %v", err)
				}

				return ctx, nil
			},
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if regexPod.FindString(output) == "" {
					t.Fatalf("expected Pod results in output %v", output)
				}
				if !it.EmojisExistInOutput(output, it.GetEmojis) {
					t.Fatalf("expected emojis in get output %v", output)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				os.Stdout = ctx.Value("stdout").(*os.File)
				return nil
			},
		},
		{
			Name:         "Get the updated workload with no color flag",
			Skip:         runtime.GOOS == "windows",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "get", "test-create-git-annotations-workload", namespaceFlag, "--no-color"),
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				var err error
				if ctx, err = createPtyTerminal(ctx); err != nil {
					t.Fatalf("error while opening pty %v", err)
				}

				return ctx, nil
			},
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if regexPod.FindString(output) == "" {
					t.Fatalf("expected Pod results in output %v", output)
				}
				if it.EmojisExistInOutput(output, it.GetEmojis) {
					t.Fatalf("did not expect emojis in get output with no color flag %v", output)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				os.Stdout = ctx.Value("stdout").(*os.File)
				return nil
			},
		},
		{
			Name:         "Get the updated workload with no color envvar",
			Skip:         runtime.GOOS == "windows",
			WorkloadName: "test-create-git-annotations-workload",
			Command: func() it.CommandLine {
				c := it.NewTanzuAppsCommandLine(
					"workload", "get", "test-create-git-annotations-workload", namespaceFlag)
				c.AddEnvVars("NO_COLOR=true")
				return *c
			}(),
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				var err error
				if ctx, err = createPtyTerminal(ctx); err != nil {
					t.Fatalf("error while opening pty %v", err)
				}

				return ctx, nil
			},
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if regexPod.FindString(output) == "" {
					t.Fatalf("expected Pod results in output %v", output)
				}
				if it.EmojisExistInOutput(output, it.GetEmojis) {
					t.Fatalf("did not expect emojis in get output with no color envvar %v", output)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				os.Stdout = ctx.Value("stdout").(*os.File)
				return nil
			},
		},
		{
			Name:         "Delete the created workload and show emojis",
			WorkloadName: "test-create-workload-show-emojis",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "delete", "test-create-workload-show-emojis", namespaceFlag, "-y"),
			Prepare: func(ctx context.Context, t *testing.T) (context.Context, error) {
				var err error
				if ctx, err = createPtyTerminal(ctx); err != nil {
					t.Fatalf("error while opening pty %v", err)
				}

				return ctx, nil
			},
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				if !it.EmojisExistInOutput(output, it.DeleteEmojis) {
					t.Fatalf("expected emojis in delete output %v", output)
				}
			},
			CleanUp: func(ctx context.Context, t *testing.T) error {
				os.Stdout = ctx.Value("stdout").(*os.File)
				return nil
			},
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

func createPtyTerminal(ctx context.Context) (context.Context, error) {
	r, w, err := pty.Open()
	if err != nil {
		return ctx, err
	}
	originalStdout := os.Stdout
	os.Stdout = w

	ctx = context.WithValue(ctx, "reader", r)
	ctx = context.WithValue(ctx, "writer", w)
	ctx = context.WithValue(ctx, "stdout", originalStdout)

	return ctx, err
}
