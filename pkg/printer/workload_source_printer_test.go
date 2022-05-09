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
sub-path:   my-subpath
branch:     my-branch
tag:        my-tag
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

func TestWorkloadResourcesPrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *cartov1alpha1.Workload
		expectedOutput string
	}{{
		name: "various resources",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ResourceReady,
							Status: "True",
						},
						{
							Type:   cartov1alpha1.ResourceSubmitted,
							Status: "True",
						},
					},
				}, {
					Name: "deliverable",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ResourceReady,
							Status: "Unknown",
						},
						{
							Type:   cartov1alpha1.ResourceSubmitted,
							Status: "Unknown",
						},
					},
				}, {
					Name: "image-builder",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ResourceReady,
							Status: "False",
						},
						{
							Type:   cartov1alpha1.ResourceSubmitted,
							Status: "False",
						},
					},
				}},
			},
		},
		expectedOutput: `
RESOURCE          READY     TIME
source-provider   True      <unknown>
deliverable       Unknown   <unknown>
image-builder     False     <unknown>
`,
	}, {
		name: "no resources",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
		},
		expectedOutput: `
RESOURCE   READY   TIME
`,
	}, {
		name: "no ready condition inside resource",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ResourceSubmitted,
							Status: "True",
						},
					},
				}, {
					Name: "deliverable",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ResourceSubmitted,
							Status: "Unknown",
						},
					},
				}, {
					Name: "image-builder",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ResourceReady,
							Status: "False",
						},
						{
							Type:   cartov1alpha1.ResourceSubmitted,
							Status: "False",
						},
					},
				}},
			},
		},
		expectedOutput: `
RESOURCE          READY   TIME
source-provider           
deliverable               
image-builder     False   <unknown>
`,
	}, {
		name: "resource without conditions",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
				}, {
					Name: "deliverable",
				}},
			},
		},
		expectedOutput: `
RESOURCE          READY   TIME
source-provider           
deliverable               
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.WorkloadResourcesPrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
