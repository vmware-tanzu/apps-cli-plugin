// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TemplateParams []TemplateParam

type TemplateParam struct {
	Name         string               `json:"name"`
	DefaultValue apiextensionsv1.JSON `json:"default"`
}

type Param struct {
	Name  string               `json:"name"`
	Value apiextensionsv1.JSON `json:"value"`
}

type DelegatableParam struct {
	Name         string                `json:"name"`
	Value        *apiextensionsv1.JSON `json:"value,omitempty"`
	DefaultValue *apiextensionsv1.JSON `json:"default,omitempty"`
}

type ResourceReference struct {
	Name     string `json:"name"`
	Resource string `json:"resource"`
}

type Source struct {
	Git *GitSource `json:"git,omitempty"`
	// Image is an OCI image is a registry that contains source code
	Image   string `json:"image,omitempty"`
	Subpath string `json:"subPath,omitempty"`
}

type GitSource struct {
	URL string `json:"url,omitempty"`
	Ref GitRef `json:"ref,omitempty"`
}

type GitRef struct {
	Branch string `json:"branch,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Commit string `json:"commit,omitempty"`
}

type ObjectReference struct {
	Kind       string `json:"kind,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type ServiceAccountRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type RealizedResource struct {
	// Name is the name of the resource in the blueprint
	Name string `json:"name"`

	// StampedRef is a reference to the object that was created by the resource
	StampedRef *corev1.ObjectReference `json:"stampedRef,omitempty"`

	// TemplateRef is a reference to the template used to create the object in StampedRef
	TemplateRef *corev1.ObjectReference `json:"templateRef,omitempty"`

	// Inputs are references to resources that were used to template the object in StampedRef
	Inputs []Input `json:"inputs,omitempty"`

	// Outputs are values from the object in StampedRef that can be consumed by other resources
	Outputs []Output `json:"outputs,omitempty"`

	// Conditions describing this resource's reconcile state. The top level condition is
	// of type `Ready`, and follows these Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type Input struct {
	// Name is the name of the resource in the blueprint whose output the resource consumes as an input
	Name string `json:"name"`
}

type Output struct {
	// Name is the output type generated from the resource [url, revision, image or config]
	Name string `json:"name"`

	// Preview is a preview of the value of the output
	Preview string `json:"preview"`

	// Digest is a sha256 of the full value of the output
	Digest string `json:"digest"`

	// LastTransitionTime is a timestamp of the last time the value changed
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

type OwnerStatus struct {
	// ObservedGeneration refers to the metadata.Generation of the spec that resulted in
	// the current `status`.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions describing this resource's reconcile state. The top level condition is
	// of type `Ready`, and follows these Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
