/*
Copyright 2019 The Knative Authors

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

// +versionName=v1
// +groupName=serving.knative.dev
// +kubebuilder:object:generate=true

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ServiceConditionReady = "Ready"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Service struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Status ServiceStatus `json:"status,omitempty"`
}

// ServiceStatus represents the Status stanza of the Service resource.
type ServiceStatus struct {
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// URL holds the url that will distribute traffic over the provided traffic targets.
	// It generally has the form http[s]://{route-name}.{route-namespace}.{cluster-level-suffix}
	// +optional
	URL string `json:"url,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceList is a list of Service resources
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Service{},
		&ServiceList{},
	)
}
