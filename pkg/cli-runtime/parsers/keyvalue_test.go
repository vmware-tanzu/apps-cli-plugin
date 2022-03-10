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

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/parsers"
)

func TestKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		value    string
	}{{
		name:     "valid",
		value:    "MY_VAR=my-value",
		expected: []string{"MY_VAR", "my-value"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := parsers.KeyValue(test.value)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestDeletableKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		value    string
	}{{
		name:     "set",
		value:    "MY_VAR=my-value",
		expected: []string{"MY_VAR", "my-value"},
	}, {
		name:     "delete",
		value:    "MY_VAR-",
		expected: []string{"MY_VAR"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := parsers.DeletableKeyValue(test.value)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
