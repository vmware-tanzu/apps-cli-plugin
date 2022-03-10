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

package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

func TestObjectReference(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		value    string
	}{{
		name:     "valid",
		expected: validation.FieldErrors{},
		value:    "example.com/v1alpha1:FooBar:blah",
	}, {
		name:     "valid core/v1",
		expected: validation.FieldErrors{},
		value:    "v1:ConfigMap:blah",
	}, {
		name:     "valid with namespace",
		expected: validation.FieldErrors{},
		value:    "example.com/v1alpha1:FooBar:blah:blah-ns",
	}, {
		name:     "empty",
		expected: validation.ErrInvalidValue("", clitesting.TestField),
		value:    "",
	}, {
		name:     "invalid",
		expected: validation.ErrInvalidValue("/", clitesting.TestField),
		value:    "/",
	}, {
		name:     "missing api version",
		expected: validation.ErrInvalidValue(":FooBar:blah", clitesting.TestField),
		value:    ":FooBar:blah",
	}, {
		name:     "missing api group",
		expected: validation.ErrInvalidValue("v1alpha1:FooBar:blah", clitesting.TestField),
		value:    "v1alpha1:FooBar:blah",
	}, {
		name:     "missing kind",
		expected: validation.ErrInvalidValue("example.com/v1alpha1::blah", clitesting.TestField),
		value:    "example.com/v1alpha1::blah",
	}, {
		name:     "invalid name",
		expected: validation.ErrInvalidValue("-", clitesting.TestField),
		value:    "v1:ConfigMap:-",
	}, {
		name:     "invalid namespace",
		expected: validation.ErrInvalidValue("blah-(&", clitesting.TestField),
		value:    "example.com/v1alpha1:FooBar:blah:blah-(&",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.ObjectReference(test.value, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestDeletableObjectReference(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		value    string
	}{{
		name:     "valid delete",
		expected: validation.FieldErrors{},
		value:    "example.com/v1alpha1:FooBar:blah-",
	}, {
		name:     "valid",
		expected: validation.FieldErrors{},
		value:    "example.com/v1alpha1:FooBar:blah",
	}, {
		name:     "valid core/v1",
		expected: validation.FieldErrors{},
		value:    "v1:ConfigMap:blah",
	}, {
		name:     "empty",
		expected: validation.ErrInvalidValue("", clitesting.TestField),
		value:    "",
	}, {
		name:     "invalid",
		expected: validation.ErrInvalidValue("/", clitesting.TestField),
		value:    "/",
	}, {
		name:     "invalid delete",
		expected: validation.ErrInvalidValue("/", clitesting.TestField),
		value:    "/-",
	}, {
		name:     "missing api version",
		expected: validation.ErrInvalidValue(":FooBar:blah", clitesting.TestField),
		value:    ":FooBar:blah",
	}, {
		name:     "missing api group",
		expected: validation.ErrInvalidValue("v1alpha1:FooBar:blah", clitesting.TestField),
		value:    "v1alpha1:FooBar:blah",
	}, {
		name:     "missing kind",
		expected: validation.ErrInvalidValue("example.com/v1alpha1::blah", clitesting.TestField),
		value:    "example.com/v1alpha1::blah",
	}, {
		name:     "invalid name",
		expected: validation.ErrInvalidValue("", clitesting.TestField),
		value:    "v1:ConfigMap:-",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.DeletableObjectReference(test.value, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestObjectReferences(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		values   []string
	}{{
		name:     "valid, empty",
		expected: validation.FieldErrors{},
		values:   []string{},
	}, {
		name:     "valid, not empty",
		expected: validation.FieldErrors{},
		values:   []string{"example.com/v1alpha1:FooBar:blah"},
	}, {
		name:     "invalid",
		expected: validation.ErrInvalidValue("", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 0),
		values:   []string{""},
	}, {
		name: "multiple invalid",
		expected: validation.FieldErrors{}.Also(
			validation.ErrInvalidValue("", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 0),
			validation.ErrInvalidValue("", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 1),
		),
		values: []string{"", ""},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.ObjectReferences(test.values, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestDeletableObjectReferences(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		values   []string
	}{{
		name:     "valid, empty",
		expected: validation.FieldErrors{},
		values:   []string{},
	}, {
		name:     "valid, not empty",
		expected: validation.FieldErrors{},
		values:   []string{"example.com/v1alpha1:FooBar:blah"},
	}, {
		name:     "valid, delete",
		expected: validation.FieldErrors{},
		values:   []string{"example.com/v1alpha1:FooBar:blah-"},
	}, {
		name:     "invalid",
		expected: validation.ErrInvalidValue("", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 0),
		values:   []string{""},
	}, {
		name: "multiple invalid",
		expected: validation.FieldErrors{}.Also(
			validation.ErrInvalidValue("", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 0),
			validation.ErrInvalidValue("", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 1),
		),
		values: []string{"", ""},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.DeletableObjectReferences(test.values, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
