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

package v1alpha1_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	servicev1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/services/v1alpha1"
)

func TestServiceClaimWorkloadConfigAnnotation(t *testing.T) {
	dbServiceClaim := make(servicev1alpha1.ServiceClaims)
	dbServiceClaim["database"] = map[string]string{"namespace": "services"}

	actualRc := &servicev1alpha1.ServiceClaimWorkloadConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "supplychain.apps.x-tanzu.vmware.com/v1alpha1",
			Kind:       "somekind",
		},
		Spec: servicev1alpha1.ServiceClaimWorkloadConfigSpec{
			ServiceClaims: dbServiceClaim,
		},
	}

	expectedAnnotation := `{"kind":"somekind","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"database":{"namespace":"services"}}}}`
	if diff := cmp.Diff(expectedAnnotation, actualRc.Annotation()); diff != "" {
		t.Errorf("Annotation() (-want, +got) = %v", diff)
	}
}

func TestAddServiceClaim(t *testing.T) {
	expectedServicesClaim := make(servicev1alpha1.ServiceClaims)
	expectedServicesClaim["mongo-db"] = map[string]string{"namespace": "mongo-db-svc"}
	actual := &servicev1alpha1.ServiceClaimWorkloadConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "supplychain.apps.x-tanzu.vmware.com/v1alpha1",
			Kind:       "ServiceClaimsExtension",
		},
	}

	actual.AddServiceClaim("mongo-db", map[string]string{"namespace": "mongo-db-svc"})

	if diff := cmp.Diff(expectedServicesClaim, actual.Spec.ServiceClaims); diff != "" {
		t.Errorf("AddServiceClaim() (-want, +got) = %v", diff)
	}

	actual.AddServiceClaim("database", map[string]string{"namespace": "db-services"})
	expectedServicesClaim["database"] = map[string]string{"namespace": "db-services"}
	if diff := cmp.Diff(expectedServicesClaim, actual.Spec.ServiceClaims); diff != "" {
		t.Errorf("AddServiceClaim() (-want, +got) = %v", diff)
	}
}
func TestNewServiceClaimWorkloadConfigFromAnnotation(t *testing.T) {
	tests := []struct {
		name              string
		wantServiceClaims servicev1alpha1.ServiceClaims
		annotation        string
		shouldErr         bool
	}{{
		name: "valid annotation",
		wantServiceClaims: map[string]interface{}{
			"cache":    map[string]interface{}{"namespace": "cache-services"},
			"database": map[string]interface{}{"namespace": "db-services"},
		},
		annotation: `{"kind":"ServiceClaimsExtension","apiVersion":"supplychain.apps.x-tanzu.vmware.com/v1alpha1","spec":{"serviceClaims":{"cache":{"namespace":"cache-services"},"database":{"namespace":"db-services"}}}}`,
	}, {
		name:              "invalid annotation",
		wantServiceClaims: map[string]interface{}{},
		annotation:        "invalid stuff",
		shouldErr:         true,
	}, {
		name:              "empty",
		wantServiceClaims: map[string]interface{}{},
		annotation:        "",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := servicev1alpha1.NewServiceClaimWorkloadConfigFromAnnotation(test.annotation)
			if (err == nil) == test.shouldErr {
				t.Errorf("NewServiceClaimWorkloadConfigFromAnnotation() shouldErr %t %v", test.shouldErr, err)
			} else if test.shouldErr {
				return
			}
			if diff := cmp.Diff(test.wantServiceClaims, got.Spec.ServiceClaims); diff != "" {
				t.Errorf("NewServiceClaimWorkloadConfigFromAnnotation() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestNewServiceClaimWorkloadConfig(t *testing.T) {
	actualServiceClaimConfig := servicev1alpha1.NewServiceClaimWorkloadConfig()
	actualServiceClaimConfig.AddServiceClaim("cache", map[string]string{"namespace": "cache-services"})

	wantServiceConfig := &servicev1alpha1.ServiceClaimWorkloadConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: servicev1alpha1.ServiceClaimAPIVersion,
			Kind:       servicev1alpha1.ServiceClaimKind,
		},
		Spec: servicev1alpha1.ServiceClaimWorkloadConfigSpec{
			ServiceClaims: servicev1alpha1.ServiceClaims{
				"cache": map[string]string{"namespace": "cache-services"},
			},
		},
	}
	if diff := cmp.Diff(wantServiceConfig, actualServiceClaimConfig); diff != "" {
		t.Errorf("NewServiceClaimWorkloadConfigFromAnnotation() (-want, +got) = %v", diff)
	}
}
