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

// +versionName=v1alpha1
// +groupName=carto.run
// +kubebuilder:object:generate=true

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionReady                  = "Ready"
	WorkloadSupplyChainReady        = "SupplyChainReady"
	WorkloadResourceSubmitted       = "ResourcesSubmitted"
	ResourcesHealthy                = "ResourcesHealthy"
	WorkloadDeliverableResourceKind = "Deliverable"
)

const (
	ReadySupplyChainReason                               = "Ready"
	WorkloadLabelsMissingSupplyChainReason               = "WorkloadLabelsMissing"
	NotFoundSupplyChainReadyReason                       = "SupplyChainNotFound"
	MultipleMatchesSupplyChainReadyReason                = "MultipleSupplyChainMatches"
	ServiceAccountSecretErrorResourcesSubmittedReason    = "ServiceAccountSecretError"
	ResourceRealizerBuilderErrorResourcesSubmittedReason = "ResourceRealizerBuilderError"
)

const (
	ConditionResourceReady     = "Ready"
	ConditionResourceSubmitted = "ResourceSubmitted"
	ConditionResourceHealthy   = "Healthy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="all"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=`.spec.source['git.url','image']`
// +kubebuilder:printcolumn:name="SupplyChain",type="string",JSONPath=".status.supplyChainRef.name"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=`.status.conditions[?(@.type=='Ready')].reason`

type Workload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WorkloadSpec   `json:"spec"`
	Status            WorkloadStatus `json:"status,omitempty"`
}

type WorkloadServiceClaim struct {
	Name string                         `json:"name"`
	Ref  *WorkloadServiceClaimReference `json:"ref,omitempty"`
}

type WorkloadServiceClaimReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type WorkloadSpec struct {
	Params []Param         `json:"params,omitempty"`
	Source *Source         `json:"source,omitempty"`
	Build  *WorkloadBuild  `json:"build,omitempty"`
	Env    []corev1.EnvVar `json:"env,omitempty"`
	// Image is a pre-built image in a registry. It is an alternative to defining source
	// code.
	Image              string                       `json:"image,omitempty"`
	Resources          *corev1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccountName *string                      `json:"serviceAccountName,omitempty"`
	ServiceClaims      []WorkloadServiceClaim       `json:"serviceClaims,omitempty"`
}

type WorkloadBuild struct {
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type WorkloadStatus struct {
	// ObservedGeneration refers to the metadata.Generation of the spec that resulted in
	// the current `status`.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions describing this resource's reconcile state. The top level condition is
	// of type `Ready`, and follows these Kubernetes conventions:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// SupplyChainRef is the Supply Chain resource that was used when this status was set.
	SupplyChainRef ObjectReference `json:"supplyChainRef,omitempty"`

	// Resources contain references to the objects created by the Supply Chain and the templates used to create them.
	// It also contains Inputs and Outputs that were passed between the templates as the Supply Chain was processed.
	Resources []RealizedResource `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true

type WorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workload `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Workload{},
		&WorkloadList{},
	)
}
