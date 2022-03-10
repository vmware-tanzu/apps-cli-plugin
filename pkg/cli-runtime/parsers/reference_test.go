/*
Copyright 2019 VMware, Inc.

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

package parsers_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/parsers"
)

func TestObjectReference(t *testing.T) {
	tests := []struct {
		name     string
		expected corev1.ObjectReference
		value    string
	}{{
		name:  "valid",
		value: "example.com/v1alpha1:FooBar:blah",
		expected: corev1.ObjectReference{
			APIVersion: "example.com/v1alpha1",
			Kind:       "FooBar",
			Name:       "blah",
		},
	}, {
		name:  "valid with namespace",
		value: "example.com/v1alpha1:FooBar:blah-ns:blah",
		expected: corev1.ObjectReference{
			APIVersion: "example.com/v1alpha1",
			Kind:       "FooBar",
			Name:       "blah",
		},
	}, {
		name:  "valid core/v1",
		value: "v1:ConfigMap:blah",
		expected: corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Name:       "blah",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := parsers.ObjectReference(test.value)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestDeletableObjectReference(t *testing.T) {
	type res struct {
		Ref    corev1.ObjectReference
		Delete bool
	}

	tests := []struct {
		name     string
		expected res
		value    string
	}{{
		name:  "valid",
		value: "example.com/v1alpha1:FooBar:blah",
		expected: res{
			Ref: corev1.ObjectReference{
				APIVersion: "example.com/v1alpha1",
				Kind:       "FooBar",
				Name:       "blah",
			},
		},
	}, {
		name:  "valid",
		value: "example.com/v1alpha1:FooBar:blah-ns:blah",
		expected: res{
			Ref: corev1.ObjectReference{
				APIVersion: "example.com/v1alpha1",
				Kind:       "FooBar",
				Name:       "blah",
			},
		},
	}, {
		name:  "valid delete",
		value: "example.com/v1alpha1:FooBar:blah-",
		expected: res{
			Ref: corev1.ObjectReference{
				APIVersion: "example.com/v1alpha1",
				Kind:       "FooBar",
				Name:       "blah",
			},
			Delete: true,
		},
	}, {
		name:  "valid core/v1",
		value: "v1:ConfigMap:blah",
		expected: res{
			Ref: corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Name:       "blah",
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actualRef, actualDelete := parsers.DeletableObjectReference(test.value)
			actual := res{
				Ref:    actualRef,
				Delete: actualDelete,
			}
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestObjectReferenceAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]string
		value    string
	}{{
		name:     "valid",
		value:    "example.com/v1alpha1:FooBar:blah",
		expected: nil,
	}, {
		name:     "invalid ",
		value:    "example.com/v1alpha1:FooBar:blah-ns:blah",
		expected: map[string]string{"namespace": "blah-ns"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := parsers.ObjectReferenceAnnotation(test.value)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
