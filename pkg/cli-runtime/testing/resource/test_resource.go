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

package testing

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

var (
	_ webhook.Defaulter         = &TestResource{}
	_ client.Object             = &TestResource{}
	_ validation.FieldValidator = &TestResource{}
)

// +kubebuilder:object:root=true
// +genclient

type TestResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TestResourceSpec `json:"spec"`
	//Status TestResourceStatus `json:"status"`
}

// +kubebuilder:object:generate=true
type TestResourceSpec struct {
	Fields map[string]string `json:"fields,omitempty"`
}

func (r *TestResource) Default() {
	if r.Spec.Fields == nil {
		r.Spec.Fields = map[string]string{}
	}
	r.Spec.Fields["Defaulter"] = "ran"
}

func (r *TestResource) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	if r.Spec.Fields != nil {
		if _, ok := r.Spec.Fields["invalid"]; ok {
			errs = errs.Also(validation.ErrInvalidValue(r.Spec.Fields["invalid"], "spec.fields.invalid"))
		}
	}

	return errs
}

// +kubebuilder:object:root=true
type TestResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TestResource `json:"items"`
}

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "testing.reconciler.runtime", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// compatibility with k8s.io/code-generator
var SchemeGroupVersion = GroupVersion

func init() {
	SchemeBuilder.Register(&TestResource{}, &TestResourceList{})
}
