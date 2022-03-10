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
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetGroupVersionKind(t *testing.T) {
	tests := []struct {
		name     string
		resource interface {
			GetGroupVersionKind() schema.GroupVersionKind
		}
		want schema.GroupVersionKind
	}{{
		name:     "ClusterSupplyChain",
		resource: &ClusterSupplyChain{},
		want: schema.GroupVersionKind{
			Group:   "carto.run",
			Version: "v1alpha1",
			Kind:    "ClusterSupplyChain",
		},
	}, {
		name:     "Workload",
		resource: &Workload{},
		want: schema.GroupVersionKind{
			Group:   "carto.run",
			Version: "v1alpha1",
			Kind:    "Workload",
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.resource.GetGroupVersionKind()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("GetGroupVersionKind() (-want, +got) = %v", diff)
			}
		})
	}
}
