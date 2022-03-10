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
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ServiceClaimAPIVersion = "supplychain.apps.x-tanzu.vmware.com/v1alpha1"
const ServiceClaimKind = "ServiceClaimsExtension"

type ServiceClaims map[string]interface{}

type ServiceClaimWorkloadConfig struct {
	metav1.TypeMeta `json:",inline"`
	Spec            ServiceClaimWorkloadConfigSpec `json:"spec"`
}
type ServiceClaimWorkloadConfigSpec struct {
	ServiceClaims ServiceClaims `json:"serviceClaims"`
}

func (sc *ServiceClaimWorkloadConfig) Annotation() string {
	annotation := []byte{}
	if len(sc.Spec.ServiceClaims) > 0 {
		// TODO: capture error
		annotation, _ = json.Marshal(&sc)
	}
	return string(annotation)
}

func NewServiceClaimWorkloadConfig() *ServiceClaimWorkloadConfig {
	return &ServiceClaimWorkloadConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ServiceClaimAPIVersion,
			Kind:       ServiceClaimKind,
		},
		Spec: ServiceClaimWorkloadConfigSpec{
			ServiceClaims: make(ServiceClaims),
		},
	}
}

func NewServiceClaimWorkloadConfigFromAnnotation(annotationValue string) (*ServiceClaimWorkloadConfig, error) {
	serviceClaimConfig := &ServiceClaimWorkloadConfig{}
	err := json.Unmarshal([]byte(annotationValue), &serviceClaimConfig)
	return serviceClaimConfig, err
}

func (sc *ServiceClaimWorkloadConfig) AddServiceClaim(name string, value interface{}) {
	if sc.Spec.ServiceClaims == nil {
		sc.Spec.ServiceClaims = make(ServiceClaims)
	}
	sc.Spec.ServiceClaims[name] = value
}
