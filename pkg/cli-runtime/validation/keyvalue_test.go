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

func TestKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		value    string
	}{{
		name:     "valid",
		expected: validation.FieldErrors{},
		value:    "MY_VAR=my-value",
	}, {
		name:     "empty",
		expected: validation.ErrInvalidValue("", clitesting.TestField),
		value:    "",
	}, {
		name:     "missing name",
		expected: validation.ErrInvalidValue("=my-value", clitesting.TestField),
		value:    "=my-value",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.KeyValue(test.value, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestDeletableKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		value    string
	}{{
		name:     "empty",
		expected: validation.ErrInvalidValue("", clitesting.TestField),
		value:    "",
	}, {
		name:     "invalid",
		expected: validation.ErrInvalidValue("MY_VAR", clitesting.TestField),
		value:    "MY_VAR",
	}, {
		name:     "set",
		expected: validation.FieldErrors{},
		value:    "MY_VAR=my-value",
	}, {
		name:     "delete",
		expected: validation.FieldErrors{},
		value:    "MY_VAR-",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.DeletableKeyValue(test.value, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestKeyValues(t *testing.T) {
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
		values:   []string{"MY_VAR=my-value"},
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
			actual := validation.KeyValues(test.values, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestDeletableKeyValues(t *testing.T) {
	tests := []struct {
		name     string
		expected validation.FieldErrors
		values   []string
	}{{
		name:     "valid, empty",
		expected: validation.FieldErrors{},
		values:   []string{},
	}, {
		name:     "valid, set",
		expected: validation.FieldErrors{},
		values:   []string{"MY_VAR=my-value"},
	}, {
		name:     "valid, delete",
		expected: validation.FieldErrors{},
		values:   []string{"MY_VAR-"},
	}, {
		name:     "invalid",
		expected: validation.ErrInvalidValue("MY_VAR", validation.CurrentField).ViaFieldIndex(clitesting.TestField, 0),
		values:   []string{"MY_VAR"},
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
			actual := validation.DeletableKeyValues(test.values, clitesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
