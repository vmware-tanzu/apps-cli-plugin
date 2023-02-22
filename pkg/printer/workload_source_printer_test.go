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

package printer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestWorkloadSourceImagePrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *cartov1alpha1.Workload
		expectedOutput string
	}{{
		name: "built from image",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Image: "my-image",
			},
		},
		expectedOutput: `
   type:    image
   image:   my-image
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.WorkloadSourceImagePrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestWorkloadLocalSourceImagePrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *cartov1alpha1.Workload
		expectedOutput string
	}{{
		name: "built from image",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Image: "my-image",
				},
			},
		},
		expectedOutput: `
   type:    source image
   image:   my-image
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.WorkloadLocalSourceImagePrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestWorkloadSourceGitPrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *cartov1alpha1.Workload
		expectedOutput string
	}{{
		name: "built from git all specs",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Subpath: "my-subpath",
					Git: &cartov1alpha1.GitSource{
						URL: "https://example.com/my-repo",
						Ref: cartov1alpha1.GitRef{
							Branch: "my-branch",
							Tag:    "my-tag",
							Commit: "my-commit",
						},
					},
				},
			},
		},
		expectedOutput: `
   type:       git
   url:        https://example.com/my-repo
   branch:     my-branch
   tag:        my-tag
   sub-path:   my-subpath
   commit:     my-commit
`,
	}, {
		name: "built from git some specs",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Git: &cartov1alpha1.GitSource{
						URL: "https://example.com/my-repo",
						Ref: cartov1alpha1.GitRef{
							Branch: "my-branch",
						},
					},
				},
			},
		},
		expectedOutput: `
   type:     git
   url:      https://example.com/my-repo
   branch:   my-branch
`,
	}, {
		name: "built from git with revision",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Subpath: "my-subpath",
					Git: &cartov1alpha1.GitSource{
						URL: "https://example.com/my-repo",
						Ref: cartov1alpha1.GitRef{
							Branch: "my-branch",
							Tag:    "my-tag",
						},
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{
					{
						Outputs: []cartov1alpha1.Output{
							{
								Name:    "revision",
								Preview: "my-branch/my-commit",
							},
						},
					},
				},
			},
		},
		expectedOutput: `
   type:       git
   url:        https://example.com/my-repo
   branch:     my-branch
   tag:        my-tag
   sub-path:   my-subpath
   revision:   my-branch/my-commit
`,
	}, {
		name: "do not display revision if there is commit",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Subpath: "my-subpath",
					Git: &cartov1alpha1.GitSource{
						URL: "https://example.com/my-repo",
						Ref: cartov1alpha1.GitRef{
							Branch: "my-branch",
							Tag:    "my-tag",
							Commit: "my-commit",
						},
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{
					{
						Outputs: []cartov1alpha1.Output{
							{
								Name:    "revision",
								Preview: "my-branch/my-commit",
							},
						},
					},
				},
			},
		},
		expectedOutput: `
   type:       git
   url:        https://example.com/my-repo
   branch:     my-branch
   tag:        my-tag
   sub-path:   my-subpath
   commit:     my-commit
`,
	}, {
		name: "output with no revision",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Subpath: "my-subpath",
					Git: &cartov1alpha1.GitSource{
						URL: "https://example.com/my-repo",
						Ref: cartov1alpha1.GitRef{
							Branch: "my-branch",
							Tag:    "my-tag",
							Commit: "my-commit",
						},
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{
					{
						Outputs: []cartov1alpha1.Output{
							{
								Name:    "url",
								Preview: "my-source-controller/my-url@sha:mysha12345",
							},
						},
					},
				},
			},
		},
		expectedOutput: `
   type:       git
   url:        https://example.com/my-repo
   branch:     my-branch
   tag:        my-tag
   sub-path:   my-subpath
   commit:     my-commit
`,
	}, {
		name: "empty status",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Subpath: "my-subpath",
					Git: &cartov1alpha1.GitSource{
						URL: "https://example.com/my-repo",
						Ref: cartov1alpha1.GitRef{
							Branch: "my-branch",
							Tag:    "my-tag",
							Commit: "my-commit",
						},
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{},
		},
		expectedOutput: `
   type:       git
   url:        https://example.com/my-repo
   branch:     my-branch
   tag:        my-tag
   sub-path:   my-subpath
   commit:     my-commit
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.WorkloadSourceGitPrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
