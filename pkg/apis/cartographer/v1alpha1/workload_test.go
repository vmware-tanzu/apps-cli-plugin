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
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestWorkload_Load(t *testing.T) {
	defaultWorkload := &Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name: "spring-petclinic",
			Labels: map[string]string{
				"app.kubernetes.io/part-of":           "spring-petclinic",
				"apps.tanzu.vmware.com/workload-type": "web",
			},
		},
		Spec: WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "https://github.com/spring-projects/spring-petclinic.git",
					Ref: GitRef{
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
					corev1.ResourceMemory: resource.MustParse("1Gi")},
			},
		},
	}
	tests := []struct {
		name      string
		file      string
		want      *Workload
		shouldErr bool
	}{{
		name: "loads workload",
		file: "testdata/workload.yaml",
		want: defaultWorkload,
	}, {
		name:      "not a workload",
		file:      "testdata/supplychain.yaml",
		shouldErr: true,
	}, {
		name:      "malformed",
		file:      "testdata/malformed.yaml",
		shouldErr: true,
	}, {
		name:      "missing",
		file:      "testdata/missing.yaml",
		shouldErr: true,
	}, {
		name:      "multi document",
		file:      "testdata/multidocument.yaml",
		shouldErr: true,
	}, {
		name: "loads workload with first document empty",
		file: "testdata/multidocument_first_empty.yaml",
		want: defaultWorkload,
	}, {
		name: "loads workload with last document empty",
		file: "testdata/multidocument_last_empty.yaml",
		want: defaultWorkload,
	}, {
		name: "loads workload with first and last document empty",
		file: "testdata/multidocument_first_last_empty.yaml",
		want: defaultWorkload,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &Workload{}

			f, _ := os.Open(test.file)
			defer f.Close()

			err := got.Load(f)

			if (err == nil) == test.shouldErr {
				t.Errorf("Load() shouldErr %t %v", test.shouldErr, err)
			} else if test.shouldErr {
				return
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Load() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_IsAnnotationExists(t *testing.T) {
	tests := []struct {
		name       string
		seed       *Workload
		exists     bool
		annotation string
	}{
		{
			name: "annotation exists in workload",
			seed: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello": "world",
					},
				},
			},
			annotation: "hello",
			exists:     true,
		}, {
			name: "annotation does not exist in workload",
			seed: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"my-annotation": "my-value",
					},
				},
			},
			annotation: "hello",
			exists:     false,
		}, {
			name:       "no annotations",
			seed:       &Workload{},
			annotation: "hello",
			exists:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			annotationExists := test.seed.IsAnnotationExists(test.annotation)
			if diff := cmp.Diff(test.exists, annotationExists); diff != "" {
				t.Errorf("IsAnnotationExists() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_IsLabelExists(t *testing.T) {
	tests := []struct {
		name   string
		seed   *Workload
		exists bool
		label  string
	}{
		{
			name: "label exists in workload",
			seed: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"hello": "world",
					},
				},
			},
			label:  "hello",
			exists: true,
		}, {
			name: "label does not exist in workload",
			seed: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"my-annotation": "my-value",
					},
				},
			},
			label:  "hello",
			exists: false,
		}, {
			name:   "no labels",
			seed:   &Workload{},
			label:  "hello",
			exists: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			labelExists := test.seed.IsLabelExists(test.label)
			if diff := cmp.Diff(test.exists, labelExists); diff != "" {
				t.Errorf("IsLabelExists() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_DeleteAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		seed     *Workload
		expected *Workload
		toRemove string
	}{
		{
			name: "delete annotation",
			seed: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"hello":         "world",
						"my-annotation": "my-value",
					},
				},
			},
			toRemove: "hello",
			expected: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"my-annotation": "my-value",
					},
				},
			},
		}, {
			name: "annotation does not exist",
			seed: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"my-annotation": "my-value",
					},
				},
			},
			expected: &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"my-annotation": "my-value",
					},
				},
			},
			toRemove: "hello",
		}, {
			name:     "no annotations",
			seed:     &Workload{},
			expected: &Workload{},
			toRemove: "hello",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.seed.RemoveAnnotations(test.toRemove)
			if diff := cmp.Diff(test.seed, test.expected); diff != "" {
				t.Errorf("RemoveAnnotations() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_MergeServiceAccountName(t *testing.T) {
	serviceAccount := "test-service-account"
	updatedServiceAccount := "updated-service-account"
	tests := []struct {
		name   string
		seed   *Workload
		update string
		want   *Workload
	}{{
		name:   "empty",
		seed:   &Workload{},
		update: "",
		want:   &Workload{},
	}, {
		name: "update service account",
		seed: &Workload{
			Spec: WorkloadSpec{
				ServiceAccountName: &serviceAccount,
			},
		},
		update: updatedServiceAccount,
		want: &Workload{
			Spec: WorkloadSpec{
				ServiceAccountName: &updatedServiceAccount,
			},
		},
	}, {
		name: "delete service account",
		seed: &Workload{
			Spec: WorkloadSpec{
				ServiceAccountName: &serviceAccount,
			},
		},
		update: "",
		want: &Workload{
			Spec: WorkloadSpec{
				ServiceAccountName: nil,
			},
		},
	}, {
		name:   "add service account",
		seed:   &Workload{},
		update: updatedServiceAccount,
		want: &Workload{
			Spec: WorkloadSpec{
				ServiceAccountName: &updatedServiceAccount,
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed.DeepCopy()
			got.Spec.MergeServiceAccountName(test.update)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeServiceAccountName() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_ReplaceMetadata(t *testing.T) {
	tests := []struct {
		name   string
		seed   *Workload
		update *Workload
		want   *Workload
	}{{
		name:   "empty",
		seed:   &Workload{},
		update: &Workload{},
		want:   &Workload{},
	}, {
		name: "add all metadata to new workload",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
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
		},
		update: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("uid-xyz"),
				ResourceVersion:   "999",
				Generation:        1,
				CreationTimestamp: metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp: &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
			},
		},
		want: &Workload{
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
				CreationTimestamp:          metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
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
		},
	}, {
		name: "overwrite workload metadata",
		seed: &Workload{
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
				ResourceVersion:            "111",
				Generation:                 1,
				CreationTimestamp:          metav1.Date(2018, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2020, 6, 29, 01, 44, 05, 0, time.UTC)},
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
		},
		update: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("uid-xyz"),
				ResourceVersion:   "999",
				Generation:        1,
				CreationTimestamp: metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp: &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
			},
		},
		want: &Workload{
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
				CreationTimestamp:          metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
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
		},
	}, {
		name: "do not update to empty metadata",
		seed: &Workload{
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
				CreationTimestamp:          metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
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
		},
		update: &Workload{
			ObjectMeta: metav1.ObjectMeta{},
		},
		want: &Workload{
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
				CreationTimestamp:          metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
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
		},
	}, {
		name: "return only updated system populated fields",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{},
		},
		update: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("uid-xyz"),
				ResourceVersion:   "999",
				Generation:        1,
				CreationTimestamp: metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp: &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
			},
		},
		want: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("uid-xyz"),
				ResourceVersion:   "999",
				Generation:        1,
				CreationTimestamp: metav1.Date(2019, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp: &metav1.Time{Time: time.Date(2021, 6, 29, 01, 44, 05, 0, time.UTC)},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed.DeepCopy()
			got.ReplaceMetadata(test.update)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Merge() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_Validate(t *testing.T) {
	tests := []struct {
		name     string
		workload Workload
		want     validation.FieldErrors
	}{{
		name:     "empty",
		workload: Workload{},
		want: validation.FieldErrors{}.Also(
			validation.ErrInvalidValue("", cli.NameArgumentName),
			validation.ErrInvalidValue("", flags.NamespaceFlagName),
		),
	}, {
		name: "valid git using --git-branch",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@github.com/example/repo.git",
						Ref: GitRef{
							Branch: "main",
						},
					},
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "valid git using --git-tag",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@github.com/example/repo.git",
						Ref: GitRef{
							Tag: "1.0.1",
						},
					},
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "valid git using --git-commit",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@github.com/example/repo.git",
						Ref: GitRef{
							Commit: "abcd123",
						},
					},
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "valid git using --git-tag with subPath",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Subpath: "test-path",
					Git: &GitSource{
						URL: "git@github.com/example/repo.git",
						Ref: GitRef{
							Tag: "1.0.1",
						},
					},
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "missing required git fields",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{},
				},
			},
		},
		want: validation.FieldErrors{}.Also(
			validation.ErrMissingField(flags.GitRepoFlagName),
			validation.ErrMissingOneOf(flags.GitBranchFlagName, flags.GitTagFlagName, flags.GitCommitFlagName),
		),
	}, {
		name: "valid source image",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Image: "registry.example/image:tag",
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "valid source image with subpath",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "registry.example/image:tag",
					Subpath: "test-path",
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "duplicate source",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@github.com/example/repo.git",
						Ref: GitRef{
							Branch: "main",
						},
					},
					Image: "registry.example/image:tag",
				},
			},
		},
		want: validation.ErrMultipleOneOf(flags.GitFlagWildcard, flags.SourceImageFlagName),
	}, {
		name: "valid image",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "valid image with subpath",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
				Source: &Source{
					Subpath: "test-paths",
				},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "source and image",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@github.com/example/repo.git",
						Ref: GitRef{
							Branch: "main",
						},
					},
				},
				Image: "ubuntu:bionic",
			},
		},
		want: validation.ErrMultipleOneOf(flags.GitFlagWildcard, flags.SourceImageFlagName, flags.ImageFlagName),
	}, {
		name: "subPath with no source",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Source: &Source{
					Subpath: "test-paths",
				},
			},
		},
		want: validation.ErrInvalidValue("test-paths", flags.SubPathFlagName),
	}, {
		name: "valid mavensource",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Params: []Param{{
					Name:  WorkloadMavenParam,
					Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.1"}`)},
				}},
			},
		},
		want: validation.FieldErrors{},
	}, {
		name: "maven with no artifactID",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Params: []Param{{
					Name:  WorkloadMavenParam,
					Value: apiextensionsv1.JSON{Raw: []byte(`{"groupId":"bar","version":"0.1.1"}`)},
				}},
			},
		},
		want: validation.ErrMissingField(flags.MavenArtifactFlagName),
	}, {
		name: "maven with no groupID",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Params: []Param{{
					Name:  WorkloadMavenParam,
					Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","version":"0.1.1"}`)},
				}},
			},
		},
		want: validation.ErrMissingField(flags.MavenGroupFlagName),
	}, {
		name: "maven with no version",
		workload: Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
			},
			Spec: WorkloadSpec{
				Params: []Param{{
					Name:  WorkloadMavenParam,
					Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar"}`)},
				}},
			},
		},
		want: validation.ErrMissingField(flags.MavenVersionFlagName),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.workload.Validate()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Validate() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkload_Merge(t *testing.T) {
	tests := []struct {
		name   string
		seed   *Workload
		update *Workload
		want   *Workload
	}{{
		name:   "empty",
		seed:   &Workload{},
		update: &Workload{},
		want:   &Workload{},
	}, {
		name: "annotations",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"existing":  "value",
					"overwrite": "value",
				},
			},
		},
		update: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"overwrite": "new-value",
				},
			},
		},
		want: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"existing":  "value",
					"overwrite": "new-value",
				},
			},
		},
	}, {
		name: "labels",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"existing":  "value",
					"overwrite": "value",
				},
			},
		},
		update: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"overwrite": "new-value",
				},
			},
		},
		want: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"existing":  "value",
					"overwrite": "new-value",
				},
			},
		},
	}, {
		name: "params",
		seed: &Workload{
			Spec: WorkloadSpec{
				Params: []Param{
					{
						Name:  "existing",
						Value: apiextensionsv1.JSON{Raw: []byte(`true`)},
					},
					{
						Name:  "overwrite",
						Value: apiextensionsv1.JSON{Raw: []byte(`false`)},
					},
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Params: []Param{
					{
						Name:  "overwrite",
						Value: apiextensionsv1.JSON{Raw: []byte(`true`)},
					},
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Params: []Param{
					{
						Name:  "existing",
						Value: apiextensionsv1.JSON{Raw: []byte(`true`)},
					},
					{
						Name:  "overwrite",
						Value: apiextensionsv1.JSON{Raw: []byte(`true`)},
					},
				},
			},
		},
	}, {
		name: "image",
		seed: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
	}, {
		name: "image without source",
		seed: &Workload{},
		update: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
	}, {
		name: "image with subpath",
		seed: &Workload{
			Spec: WorkloadSpec{
				Image: "alpine:latest",
				Source: &Source{
					Subpath: "/sys",
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
	}, {
		name: "image with sources nill",
		seed: &Workload{
			Spec: WorkloadSpec{
				Image: "alpine:latest",
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Image:  "ubuntu:bionic",
				Source: &Source{Subpath: "my-subpath"},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
	}, {
		name: "image with sources source image and subpath",
		seed: &Workload{
			Spec: WorkloadSpec{
				Image: "alpine:latest",
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "my-subpath",
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "my-subpath",
				},
			},
		},
	}, {
		name: "image update sources source image and subpath",
		seed: &Workload{
			Spec: WorkloadSpec{
				Image: "alpine:latest",
				Source: &Source{
					Subpath: "new-subpath",
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "my-subpath",
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "my-subpath",
				},
			},
		},
	}, {
		name: "local source image with subpath update source image and subpath",
		seed: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "/opt",
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "/sys",
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image:   "ubuntu:bionic",
					Subpath: "/sys",
				},
			},
		},
	}, {
		name: "git",
		seed: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@example.com/repo.git",
					},
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Git: &GitSource{
						URL: "git@example.com/repo.git",
					},
				},
			},
		},
	}, {
		name: "source image",
		seed: &Workload{
			Spec: WorkloadSpec{
				Image: "ubuntu:bionic",
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image: "registry.example/repo:tag",
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Image: "registry.example/repo:tag",
				},
			},
		},
	}, {
		name: "subPath with no source",
		seed: &Workload{
			Spec: WorkloadSpec{},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Source: &Source{
					Subpath: "test-path",
				},
			},
		},
		want: &Workload{},
	}, {
		name: "env",
		seed: &Workload{
			Spec: WorkloadSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "EXISTING",
						Value: "true",
					},
					{
						Name:  "OVERWRITE",
						Value: "false",
					},
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "OVERWRITE",
						Value: "true",
					},
				},
			},
		},
		want: &Workload{
			Spec: WorkloadSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "EXISTING",
						Value: "true",
					},
					{
						Name:  "OVERWRITE",
						Value: "true",
					},
				},
			},
		},
	}, {
		name: "resources",
		seed: &Workload{
			Spec: WorkloadSpec{
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("500Mi"),
					},
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			}},
		want: &Workload{
			Spec: WorkloadSpec{
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			}},
	}, {
		name: "service claims",
		seed: &Workload{
			Spec: WorkloadSpec{
				ServiceClaims: []WorkloadServiceClaim{
					{
						Name: "existing",
						Ref: &WorkloadServiceClaimReference{
							APIVersion: "v1",
							Kind:       "Secret",
							Name:       "secret",
						},
					},
				},
			},
		},
		update: &Workload{
			Spec: WorkloadSpec{
				ServiceClaims: []WorkloadServiceClaim{
					{
						Name: "new",
						Ref: &WorkloadServiceClaimReference{
							APIVersion: "v1",
							Kind:       "Secret",
							Name:       "secret",
						},
					},
				},
			}},
		want: &Workload{
			Spec: WorkloadSpec{
				ServiceClaims: []WorkloadServiceClaim{
					{
						Name: "existing",
						Ref: &WorkloadServiceClaimReference{
							APIVersion: "v1",
							Kind:       "Secret",
							Name:       "secret",
						},
					}, {
						Name: "new",
						Ref: &WorkloadServiceClaimReference{
							APIVersion: "v1",
							Kind:       "Secret",
							Name:       "secret",
						},
					},
				},
			}},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed.DeepCopy()
			got.Merge(test.update)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Merge() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeParams(t *testing.T) {
	tests := []struct {
		name  string
		seed  *WorkloadSpec
		key   string
		value interface{}
		want  *WorkloadSpec
	}{{
		name:  "add",
		seed:  &WorkloadSpec{},
		key:   "foo",
		value: "bar",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "foo",
					Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
				},
			},
		},
	}, {
		name: "update",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "foo",
					Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
				},
			},
		},
		key:   "foo",
		value: "more bar",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "foo",
					Value: apiextensionsv1.JSON{Raw: []byte(`"more bar"`)},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeParams(test.key, test.value)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeParams() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_RemoveParam(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		key  string
		want *WorkloadSpec
	}{{
		name: "found",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "foo",
					Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
				},
				{
					Name:  "bleep",
					Value: apiextensionsv1.JSON{Raw: []byte(`"bloop"`)},
				},
			},
		},
		key: "bleep",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "foo",
					Value: apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.RemoveParam(test.key)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("RemoveParam() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeAnnotationParams(t *testing.T) {
	tests := []struct {
		name  string
		seed  *WorkloadSpec
		key   string
		value string
		want  *WorkloadSpec
	}{{
		name:  "add",
		seed:  &WorkloadSpec{},
		key:   "foo",
		value: "bar",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"bar"}`)},
				},
			},
		},
	}, {
		name: "add to existing annotation params",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"bar"}`)},
				},
			},
		},
		key:   "bar",
		value: "baz",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"baz","foo":"bar"}`)},
				},
			},
		},
	}, {
		name: "override",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  WorkloadAnnotationParam,
					Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"baz"}`)},
				},
			},
		},
		key:   "bar",
		value: "xyz",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  WorkloadAnnotationParam,
					Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"xyz"}`)},
				},
			},
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeAnnotationParams(test.key, test.value)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeAnnotationParams() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_RemoveAnnotationParams(t *testing.T) {
	tests := []struct {
		name  string
		seed  *WorkloadSpec
		value string
		want  *WorkloadSpec
	}{{
		name: "remove non-existing key in empty annotations",
		seed: &WorkloadSpec{
			Params: []Param{},
		},
		value: "xyz",
		want: &WorkloadSpec{
			Params: []Param{},
		},
	}, {
		name: "remove non-existing key",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"baz","foo":"bar"}`)},
				},
			},
		},
		value: "xyz",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"baz","foo":"bar"}`)},
				},
			},
		},
	}, {
		name: "remove existing key",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"bar","bar":"baz"}`)},
				},
			},
		},
		value: "foo",
		want: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"baz"}`)},
				},
			},
		},
	}, {
		name: "remove annotation param",
		seed: &WorkloadSpec{
			Params: []Param{
				{
					Name:  "annotations",
					Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"bar"}`)},
				},
			},
		},
		value: "foo",
		want: &WorkloadSpec{
			Params: []Param{},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.RemoveAnnotationParams(test.value)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("RemoveAnnotationParams() (-want: %v, +got: %v)", test.want.Params, got.Params)

			}
		})
	}
}

func TestWorkloadSpec_MergeMavenSource(t *testing.T) {
	temp := "jar"
	tests := []struct {
		name  string
		seed  *WorkloadSpec
		value MavenSource
		want  *WorkloadSpec
	}{{
		name: "add maven info",
		seed: &WorkloadSpec{},
		value: MavenSource{
			ArtifactId: "foo",
			GroupId:    "bar",
			Version:    "0.1.0",
		},
		want: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.0"}`)},
			}},
		},
	}, {
		name: "change maven info",
		seed: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"version":"0.1.0"}`)},
			}},
		},
		value: MavenSource{
			ArtifactId: "foo",
			GroupId:    "bar",
			Version:    "0.1.1",
		},
		want: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.1"}`)},
			}},
		},
	}, {
		name: "change maven type",
		seed: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"version":"0.1.0"}`)},
			}},
		},
		value: MavenSource{
			Type: &temp,
		},
		want: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"","groupId":"","version":"0.1.0","type":"jar"}`)},
			}},
		},
	}, {
		name: "no change",
		seed: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.0"}`)},
			}},
		},
		value: MavenSource{
			ArtifactId: "foo",
			GroupId:    "bar",
			Version:    "0.1.0",
		},
		want: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.0"}`)},
			}},
		},
	}, {
		name: "empty source",
		seed: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.0"}`)},
			}},
		},
		value: MavenSource{},
		want: &WorkloadSpec{
			Params: []Param{{
				Name:  "maven",
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.0"}`)},
			}},
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.seed.MergeMavenSource(test.value)
			if diff := cmp.Diff(test.want, test.seed); diff != "" {
				t.Errorf("MergeMavenSource() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeGit(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		git  GitSource
		want *WorkloadSpec
	}{{
		name: "set",
		seed: &WorkloadSpec{
			Image: "ubuntu:bionic",
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Branch: "main",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
					},
				},
			},
		},
	}, {
		name: "update url",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "a bad url to be replaced",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
					},
				},
			},
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Branch: "main",
				Tag:    "v1.0.0",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
					},
				},
			},
		},
	}, {
		name: "update tag",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
					},
				},
			},
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Tag: "v1.0.1",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Tag: "v1.0.1",
					},
				},
			},
		},
	}, {
		name: "update commit",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
						Commit: "abcd1234",
					},
				},
			},
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Branch: "main",
				Tag:    "v1.0.0",
				Commit: "efgh5678",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
						Commit: "efgh5678",
					},
				},
			},
		},
	}, {
		name: "update branch",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
						Commit: "abcd1234",
					},
				},
			},
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Branch: "my-new-branch",
				Tag:    "v1.0.0",
				Commit: "abcd1234",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "my-new-branch",
						Tag:    "v1.0.0",
						Commit: "abcd1234",
					},
				},
			},
		},
	}, {
		name: "update all git ref",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
						Commit: "abcd1234",
					},
				},
			},
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Branch: "my-new-branch",
				Tag:    "v1.1",
				Commit: "efgh5678",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "my-new-branch",
						Tag:    "v1.1",
						Commit: "efgh5678",
					},
				},
			},
		},
	}, {
		name: "update git source without deleting subpath",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "my-bad-branch",
						Tag:    "v1.0.0",
					},
				},
				Subpath: "my-subpath",
			},
		},
		git: GitSource{
			URL: "git@github.com:example/repo.git",
			Ref: GitRef{
				Branch: "main",
			},
		},
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
					},
				},
				Subpath: "my-subpath",
			},
		},
	}, {
		name: "delete source when setting repo to empty string",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
					},
				},
				Subpath: "my-subpath",
			},
		},
		git: GitSource{
			URL: "",
		},
		want: &WorkloadSpec{},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeGit(test.git)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeGit() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeSourceImage(t *testing.T) {
	tests := []struct {
		name        string
		seed        *WorkloadSpec
		sourceImage string
		want        *WorkloadSpec
	}{{
		name: "set",
		seed: &WorkloadSpec{
			Source: &Source{},
		},
		sourceImage: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
		want: &WorkloadSpec{
			Source: &Source{
				Image: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
			},
		},
	}, {
		name: "update",
		seed: &WorkloadSpec{
			Source: &Source{
				Image: "my-registry.nip.io/my-folder/my-old-image:latest@sha:my-sha1234567890",
			},
		},
		sourceImage: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
		want: &WorkloadSpec{
			Source: &Source{
				Image: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
			},
		},
	}, {
		name: "change",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Tag: "v1.0.0",
					},
				},
			},
		},
		sourceImage: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
		want: &WorkloadSpec{
			Source: &Source{
				Image: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
			},
		},
	}, {
		name: "update without deleting subpath",
		seed: &WorkloadSpec{
			Source: &Source{
				Image:   "my-registry.nip.io/my-folder/my-old-image:latest@sha:my-sha1234567890",
				Subpath: "my-subpath",
			},
		},
		sourceImage: "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
		want: &WorkloadSpec{
			Source: &Source{
				Image:   "my-registry.nip.io/my-folder/my-image:latest@sha:my-sha1234567890",
				Subpath: "my-subpath",
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeSourceImage(test.sourceImage)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeImage() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeImage(t *testing.T) {
	tests := []struct {
		name  string
		seed  *WorkloadSpec
		image string
		want  *WorkloadSpec
	}{{
		name: "set",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
					},
				},
			},
		},
		image: "ubuntu:bionic",
		want: &WorkloadSpec{
			Image: "ubuntu:bionic",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeImage(test.image)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeImage() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeSubpath(t *testing.T) {
	tests := []struct {
		name    string
		seed    *WorkloadSpec
		subPath string
		want    *WorkloadSpec
	}{{
		name: "set",
		seed: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
					},
				},
			},
		},
		subPath: "./cmd",
		want: &WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com:example/repo.git",
					Ref: GitRef{
						Branch: "main",
						Tag:    "v1.0.0",
					},
				},
				Subpath: "./cmd",
			},
		},
	}, {
		name: "unset",
		seed: &WorkloadSpec{
			Source: &Source{
				Image:   "app.registry.com:source",
				Subpath: "./cmd",
			},
		},
		subPath: "",
		want: &WorkloadSpec{
			Source: &Source{
				Image: "app.registry.com:source",
			},
		},
	}, {
		name:    "empty source",
		seed:    &WorkloadSpec{},
		subPath: "",
		want: &WorkloadSpec{
			Source: &Source{},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeSubPath(test.subPath)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeSubPath() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeEnv(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		env  corev1.EnvVar
		want *WorkloadSpec
	}{{
		name: "append",
		seed: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "FOO", Value: "foo"},
			},
		},
		env: corev1.EnvVar{Name: "BAR", Value: "bar"},
		want: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "FOO", Value: "foo"},
				{Name: "BAR", Value: "bar"},
			},
		},
	}, {
		name: "replace",
		seed: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "NAME", Value: "initial"},
			},
		},
		env: corev1.EnvVar{Name: "NAME", Value: "replace"},
		want: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "NAME", Value: "replace"},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeEnv(test.env)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeEnv() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_RemoveEnv(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		env  string
		want *WorkloadSpec
	}{{
		name: "found",
		seed: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "FOO", Value: "foo"},
				{Name: "BAR", Value: "bar"},
			},
		},
		env: "FOO",
		want: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "BAR", Value: "bar"},
			},
		},
	}, {
		name: "not found",
		seed: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "BAR", Value: "bar"},
			},
		},
		env: "FOO",
		want: &WorkloadSpec{
			Env: []corev1.EnvVar{
				{Name: "BAR", Value: "bar"},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.RemoveEnv(test.env)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("RemoveEnv() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeResources(t *testing.T) {
	tests := []struct {
		name      string
		seed      *WorkloadSpec
		resources *corev1.ResourceRequirements
		want      *WorkloadSpec
	}{{
		name: "empty",
		seed: &WorkloadSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				},
			},
		},
		resources: nil,
		want: &WorkloadSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				},
			},
		},
	}, {
		name: "add",
		seed: &WorkloadSpec{},
		resources: &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("500m"),
			},
		},
		want: &WorkloadSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				},
			},
		},
	}, {
		name: "replace",
		seed: &WorkloadSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
		},
		resources: &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("2"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
		},
		want: &WorkloadSpec{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi")},
			},
		}}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeResources(test.resources)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeResources() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeServiceClaim(t *testing.T) {
	tests := []struct {
		name         string
		seed         *WorkloadSpec
		serviceClaim WorkloadServiceClaim
		want         *WorkloadSpec
	}{{
		name: "add",
		seed: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{},
		},
		serviceClaim: WorkloadServiceClaim{
			Name: "database",
			Ref: &WorkloadServiceClaimReference{
				APIVersion: "services.tanzu.vmware.com/v1alpha1",
				Kind:       "PostgreSQL",
				Name:       "my-prod-db",
			},
		},
		want: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{
				{
					Name: "database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "my-prod-db",
					},
				},
			},
		},
	}, {
		name: "replace",
		seed: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{
				{
					Name: "database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "my-prod-db",
					},
				},
			},
		},
		serviceClaim: WorkloadServiceClaim{
			Name: "database",
			Ref: &WorkloadServiceClaimReference{
				APIVersion: "services.tanzu.vmware.com/v1alpha1",
				Kind:       "PostgreSQL",
				Name:       "prod-db",
			},
		},
		want: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{
				{
					Name: "database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "prod-db",
					},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeServiceClaim(test.serviceClaim)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeServiceClaim() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpecDeleteServiceClaim(t *testing.T) {
	tests := []struct {
		name             string
		seed             *WorkloadSpec
		serviceClaimName string
		want             *WorkloadSpec
	}{{
		name: "delete existing",
		seed: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{
				{
					Name: "database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "my-prod-db",
					},
				},
			},
		},
		serviceClaimName: "database",
		want: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{},
		},
	}, {
		name: "delete non-existing",
		seed: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{
				{
					Name: "main-database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "my-prod-db",
					},
				},
			},
		},
		serviceClaimName: "database",
		want: &WorkloadSpec{
			ServiceClaims: []WorkloadServiceClaim{
				{
					Name: "main-database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "my-prod-db",
					},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.DeleteServiceClaim(test.serviceClaimName)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("DeleteServiceClaim() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeAnnotations(t *testing.T) {
	tests := []struct {
		name            string
		seed            *Workload
		annotationKey   string
		annotationValue string
		want            map[string]string
	}{{
		name:            "create annotation",
		seed:            &Workload{},
		annotationKey:   "test.annotation.io",
		annotationValue: "bar",
		want: map[string]string{
			"test.annotation.io": "bar",
		},
	}, {
		name: "append",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"test.annotation.io": "foo",
				},
			},
		},
		annotationKey:   "test2.annotation.io",
		annotationValue: "bar",
		want: map[string]string{
			"test.annotation.io":  "foo",
			"test2.annotation.io": "bar",
		},
	}, {
		name: "replace",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"test.annotation.io": "foo",
				},
			},
		},
		annotationKey:   "test.annotation.io",
		annotationValue: "bar",
		want: map[string]string{
			"test.annotation.io": "bar",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeAnnotations(test.annotationKey, test.annotationValue)
			if diff := cmp.Diff(test.want, got.GetAnnotations()); diff != "" {
				t.Errorf("MergeEnv() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_MergeLabels(t *testing.T) {
	tests := []struct {
		name       string
		seed       *Workload
		labelKey   string
		labelValue string
		want       map[string]string
	}{{
		name:       "create label",
		seed:       &Workload{},
		labelKey:   "test.label.io",
		labelValue: "bar",
		want: map[string]string{
			"test.label.io": "bar",
		},
	}, {
		name: "append",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"test.label.io": "foo",
				},
			},
		},
		labelKey:   "test2.label.io",
		labelValue: "bar",
		want: map[string]string{
			"test.label.io":  "foo",
			"test2.label.io": "bar",
		},
	}, {
		name: "replace",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"test.label.io": "foo",
				},
			},
		},
		labelKey:   "test.label.io",
		labelValue: "bar",
		want: map[string]string{
			"test.label.io": "bar",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeLabels(test.labelKey, test.labelValue)
			if diff := cmp.Diff(test.want, got.GetLabels()); diff != "" {
				t.Errorf("MergeEnv() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadReadyConditionFunc(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	tests := []struct {
		name     string
		workload *Workload
		err      error
		expected bool
	}{{
		name: "unknown status",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
			Status: WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:   WorkloadConditionReady,
						Status: metav1.ConditionUnknown,
					},
				},
			},
		},
	}, {
		name: "false status",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
			Status: WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:    WorkloadConditionReady,
						Status:  metav1.ConditionFalse,
						Message: "something went wrong",
					},
				},
			},
		},
		expected: true,
		err:      fmt.Errorf("Failed to become ready: %s", "something went wrong"),
	}, {
		name: "true status",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
			Status: WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:   WorkloadConditionReady,
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
		expected: true,
	}, {
		name: "wrong generation",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
			Status: WorkloadStatus{
				ObservedGeneration: 10,
				Conditions: []metav1.Condition{
					{
						Type:   WorkloadConditionReady,
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
	}, {
		name: "no status",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
		},
	}, {
		name: "non ready status",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
			Status: WorkloadStatus{
				ObservedGeneration: 10,
				Conditions: []metav1.Condition{
					{
						Type:   "Success",
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
	}, {
		name: "ready with no status",
		workload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: defaultNamespace,
				Name:      workloadName,
			},
			Status: WorkloadStatus{
				ObservedGeneration: 10,
				Conditions: []metav1.Condition{
					{
						Type: WorkloadConditionReady,
					},
				},
			},
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualBool, err := WorkloadReadyConditionFunc(test.workload)

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("expected error %v, actually %v", expected, actual)
			}
			if test.expected != actualBool {
				t.Errorf("expected bool value %v, actually %v", test.expected, actualBool)
			}

		})
	}
}

func TestMergeServiceClaimAnnotation(t *testing.T) {
	tests := []struct {
		name             string
		seed             map[string]string
		serviceClaimName string
		want             map[string]string
		claimValue       map[string]string
	}{{
		name:             "add new",
		serviceClaimName: "cache",
		seed: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
		},
		claimValue: map[string]string{"namespace": "test-ns"},
		want: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"cache":{"namespace":"test-ns"},"database":{"namespace":"my-prod-ns"}}}}`,
		},
	}, {
		name:             "override existing",
		serviceClaimName: "database",
		claimValue:       map[string]string{"namespace": "my-new-prod-ns"},
		seed: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
		},
		want: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-new-prod-ns"}}}}`,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: test.seed,
				},
			}
			got.MergeServiceClaimAnnotation(test.serviceClaimName, test.claimValue)
			if diff := cmp.Diff(test.want, got.GetAnnotations()); diff != "" {
				t.Errorf("MergeServiceClaimAnnotation() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestDeleteServiceClaimAnnotation(t *testing.T) {
	tests := []struct {
		name             string
		seed             map[string]string
		serviceClaimName string
		want             map[string]string
	}{{
		name:             "delete existing",
		serviceClaimName: "database",
		seed: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
		},
		want: map[string]string{},
	}, {
		name:             "delete non-existing",
		serviceClaimName: "cache",
		seed: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
		},
		want: map[string]string{
			apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &Workload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: test.seed,
				},
			}
			got.DeleteServiceClaimAnnotation(test.serviceClaimName)
			if diff := cmp.Diff(test.want, got.GetAnnotations()); diff != "" {
				t.Errorf("DeleteServiceClaimAnnotation() (-want, +got) = %v", diff)
			}
		})
	}
}
func TestWorkloadSpec_MergeBuildEnv(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		env  corev1.EnvVar
		want *WorkloadSpec
	}{{
		name: "initialize build",
		seed: &WorkloadSpec{},
		env:  corev1.EnvVar{Name: "BAR", Value: "baz"},
		want: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "BAR", Value: "baz"},
				},
			},
		},
	}, {
		name: "append",
		seed: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "FOO", Value: "bar"},
				},
			},
		},
		env: corev1.EnvVar{Name: "BAR", Value: "baz"},
		want: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "FOO", Value: "bar"},
					{Name: "BAR", Value: "baz"},
				},
			},
		},
	}, {
		name: "replace",
		seed: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "NAME", Value: "initial"},
				},
			},
		},
		env: corev1.EnvVar{Name: "NAME", Value: "replace"},
		want: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "NAME", Value: "replace"},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.MergeBuildEnv(test.env)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MergeBuildEnv() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWorkloadSpec_RemoveBuildEnv(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		env  string
		want *WorkloadSpec
	}{{
		name: "found build env",
		seed: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "Foo", Value: "bar"},
					{Name: "BAR", Value: "baz"},
				},
			},
		},
		env: "BAR",
		want: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "Foo", Value: "bar"},
				},
			},
		},
	}, {
		name: "remove build",
		seed: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "BAR", Value: "bar"},
				},
			},
		},
		env:  "BAR",
		want: &WorkloadSpec{},
	}, {
		name: "not found build env",
		seed: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "BAR", Value: "bar"},
				},
			},
		},
		env: "FOO",
		want: &WorkloadSpec{
			Build: &WorkloadBuild{
				Env: []corev1.EnvVar{
					{Name: "BAR", Value: "bar"},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed
			got.RemoveBuildEnv(test.env)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("RemoveBuildEnv() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestDeprecationWarnings(t *testing.T) {
	tests := []struct {
		name string
		seed *Workload
		want []string
	}{{
		name: "service claim annotation set",
		seed: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					apis.ServiceClaimAnnotationName: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"my-prod-ns"}}}}`,
				},
			},
		},
		want: []string{"Cross namespace service claims are deprecated. Please use `tanzu service claim create` instead."},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotWarningMsg := test.seed.DeprecationWarnings()
			if !cmp.Equal(test.want, gotWarningMsg) {
				t.Errorf("DeprecationWarnings() (-want %v, +got %v)", test.want, gotWarningMsg)
			}
		})
	}
}

func TestIsSourceFound(t *testing.T) {
	tests := []struct {
		name string
		seed WorkloadSpec
		want bool
	}{{
		name: "git source",
		seed: WorkloadSpec{
			Source: &Source{
				Git: &GitSource{
					URL: "git@github.com/example/repo.git",
					Ref: GitRef{
						Branch: "main",
					},
				},
			},
		},
		want: true,
	}, {
		name: "source image",
		seed: WorkloadSpec{
			Source: &Source{
				Image: "app:source",
			},
		},
		want: true,
	}, {
		name: "nil source",
		seed: WorkloadSpec{
			Source: &Source{},
		},
		want: false,
	}, {
		name: "image",
		seed: WorkloadSpec{
			Image: "app:source",
		},
		want: true,
	}, {
		name: "maven source",
		seed: WorkloadSpec{
			Params: []Param{{
				Name:  WorkloadMavenParam,
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.1"}`)},
			}},
		},
		want: true,
	}, {
		name: "empty",
		seed: WorkloadSpec{},
		want: false,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed.IsSourceFound()
			if !cmp.Equal(test.want, got) {
				t.Errorf("IsSourceFound() (-want %v, +got %v)", test.want, got)
			}
		})
	}
}

func TestGetMavenSource(t *testing.T) {
	tests := []struct {
		name string
		seed *WorkloadSpec
		want *MavenSource
	}{{
		name: "maven source",
		seed: &WorkloadSpec{
			Params: []Param{{
				Name:  WorkloadMavenParam,
				Value: apiextensionsv1.JSON{Raw: []byte(`{"artifactId":"foo","groupId":"bar","version":"0.1.1"}`)},
			}},
		},
		want: &MavenSource{
			ArtifactId: "foo",
			GroupId:    "bar",
			Version:    "0.1.1",
		},
	}, {
		name: "empty",
		seed: &WorkloadSpec{},
		want: nil,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.seed.GetMavenSource()
			if !cmp.Equal(test.want, got) {
				t.Errorf("GetMavenSource() (-want %v, +got %v)", test.want, got)
			}
		})
	}
}

func TestGetNotices(t *testing.T) {
	tests := []struct {
		name         string
		seed         *Workload
		givenNotices []string
		want         []string
	}{{
		name: "empty notice",
		seed: &Workload{},
		want: []string{"no source code or image has been specified for this workload."},
	}, {
		name:         "some notice",
		seed:         &Workload{},
		givenNotices: []string{"test notice", "another msg"},
		want:         []string{"no source code or image has been specified for this workload.", "test notice", "another msg"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			for i := range test.givenNotices {
				ctx = StashWorkloadNotice(ctx, test.givenNotices[i])
			}
			got := test.seed.GetNotices(ctx)
			if !cmp.Equal(test.want, got) {
				t.Errorf("GetNotices() (-want %v, +got %v)", test.want, got)
			}
		})
	}
}
