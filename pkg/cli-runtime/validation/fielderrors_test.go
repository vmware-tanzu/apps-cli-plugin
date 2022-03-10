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

package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestErrMissingFieldWithDetail(t *testing.T) {
	tests := []struct {
		testName string
		field    string
		msg      string
		expected validation.FieldErrors
	}{
		{
			testName: "valid",
			expected: validation.FieldErrors{k8sfield.Required(k8sfield.NewPath(flags.YesFlagName), "")},
			field:    flags.YesFlagName,
			msg:      "",
		}, {
			testName: "valid with msg",
			expected: validation.FieldErrors{k8sfield.Required(k8sfield.NewPath(flags.YesFlagName), "Missing --yes flag")},
			field:    flags.YesFlagName,
			msg:      "Missing --yes flag",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			expected := test.expected
			actual := validation.ErrMissingFieldWithDetail(test.field, test.msg)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.testName, diff)
			}
		})
	}
}
