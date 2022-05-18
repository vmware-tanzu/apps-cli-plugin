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

package parsers_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/parsers"
)

func TestJsonYamlToObject(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expectedError bool
		expected      interface{}
	}{{
		name:     "valid json",
		value:    "{\"key\":\"value\"}",
		expected: map[string]interface{}{"key": "value"},
	}, {
		name:     "valid yaml",
		value:    "\"key\": \"value\"",
		expected: map[string]interface{}{"key": "value"},
	}, {
		name:          "invalid json",
		value:         "{\"key\":\"value\"",
		expectedError: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parsers.JsonYamlToObject(test.value)
			if test.expectedError && err == nil {
				t.Error("JsonYamlObject() = expected error")
			} else if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("JsonYamlObject() = (-expected, +actual): %s", diff)
			}
		})
	}
}
