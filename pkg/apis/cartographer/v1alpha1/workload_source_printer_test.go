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

package v1alpha1

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWorkloadSourceImagePrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	strImageProvider := "image-provider"
	strImage := "image"
	preview := "example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b"

	tests := []struct {
		name           string
		testWorkload   *Workload
		expectedOutput string
	}{{
		name: "built from image",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: WorkloadSpec{
				Image: "my-image",
			},
			Status: WorkloadStatus{
				Resources: []RealizedResource{
					{
						Name: strImageProvider,
						Outputs: []Output{
							{
								Name:    strImage,
								Preview: preview,
							},
						},
					},
				},
			},
		},
		expectedOutput: `
Source:    Image
Image:     example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b
Updated:   <unknown>
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := WorkloadSourceImagePrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestWorkloadSourceLocalImagePrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	strImageBuilder := "image-builder"
	strImage := "image"
	preview := "example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b"

	tests := []struct {
		name           string
		testWorkload   *Workload
		expectedOutput string
	}{{
		name: "built from image",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Image: "my-image",
				},
			},
			Status: WorkloadStatus{
				Resources: []RealizedResource{
					{
						Name: strImageBuilder,
						Outputs: []Output{
							{
								Name:    strImage,
								Preview: preview,
							},
						},
					},
				},
			},
		},
		expectedOutput: `
Source:         Local source code
Source-image:   example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b
Updated:        <unknown>
`,
	}, {
		name: "built from image with sub-path",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "my-image",
					Subpath: "my-subpath",
				},
			},
			Status: WorkloadStatus{
				Resources: []RealizedResource{
					{
						Name: strImageBuilder,
						Outputs: []Output{
							{
								Name:    strImage,
								Preview: preview,
							},
						},
					},
				},
			},
		},
		expectedOutput: `
Source:         Local source code
Sub-path:       my-subpath
Source-image:   example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b
Updated:        <unknown>
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := WorkloadSourceLocalImagePrinter(output, test.testWorkload); err != nil {
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
	strImageBuilder := "image-builder"
	strImage := "image"
	preview := "example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b"

	tests := []struct {
		name           string
		testWorkload   *Workload
		expectedOutput string
	}{{
		name: "built from git all specs",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: WorkloadStatus{
				Resources: []RealizedResource{
					{
						Name: strImageBuilder,
						Outputs: []Output{
							{
								Name:    strImage,
								Preview: preview,
							},
						},
					},
				},
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Subpath: "subpath",
					Git: &GitSource{
						URL: "https://example.com/my-repo",
						Ref: GitRef{
							Branch: "my-branch",
							Tag:    "my-tag",
							Commit: "my-commit",
						},
					},
				},
			},
		},
		expectedOutput: `
Source:        https://example.com/my-repo
Sub-path:      subpath
Branch:        my-branch
Tag:           my-tag
Commit:        my-commit
Built image:   example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b
Updated:       <unknown>
`,
	}, {
		name: "built from git some specs",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: WorkloadStatus{
				Resources: []RealizedResource{
					{
						Name: "image-builder",
						Outputs: []Output{
							{
								Name:    "image",
								Preview: preview,
							},
						},
					},
				},
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "https://example.com/my-repo",
						Ref: GitRef{
							Branch: "my-branch",
						},
					},
				},
			},
		},
		expectedOutput: `
Source:        https://example.com/my-repo
Branch:        my-branch
Built image:   example.com/example/example@sha256:c844da7c2890ffe02b6c2dfe6489f3f5ae31a116277d1b511ae0a0a953306f0b
Updated:       <unknown>
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := WorkloadSourceGitPrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
