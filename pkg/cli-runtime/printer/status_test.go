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

package printer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

func TestResourceStatus(t *testing.T) {
	tests := []struct {
		name      string
		condition *metav1.Condition
		output    string
	}{{
		name: "nil",
		output: `
# test: <unknown>
`,
	}, {
		name:      "empty",
		condition: &metav1.Condition{},
		output: `
# test: <unknown>
---
lastTransitionTime: null
message: ""
reason: ""
status: ""
type: ""
`,
	}, {
		name: "unknown",
		condition: &metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionUnknown,
			Reason:  "HangOn",
			Message: "a hopefully informative message about what's in flight",
			LastTransitionTime: metav1.Time{
				Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
			},
		},
		output: `
# test: Unknown
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what's in flight
reason: HangOn
status: Unknown
type: Ready
`,
	}, {
		name: "ready",
		condition: &metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionTrue,
			LastTransitionTime: metav1.Time{
				Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
			},
		},
		output: `
# test: Ready
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: ""
reason: ""
status: "True"
type: Ready
`,
	}, {
		name: "failure",
		condition: &metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "OopsieDoodle",
			Message: "a hopefully informative message about what went wrong",
			LastTransitionTime: metav1.Time{
				Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
			},
		},
		output: `
# test: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := printer.ResourceStatus("test", test.condition)
			expected, actual := strings.TrimSpace(test.output), strings.TrimSpace(output)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestFindCondition(t *testing.T) {
	tests := []struct {
		name          string
		conditions    []metav1.Condition
		conditionType string
		output        *metav1.Condition
	}{{
		name:          "missing",
		conditions:    []metav1.Condition{},
		conditionType: "Ready",
		output:        nil,
	}, {
		name: "found",
		conditions: []metav1.Condition{
			{
				Type:   "NotReady",
				Status: metav1.ConditionTrue,
			},
			{
				Type:   "Ready",
				Status: metav1.ConditionTrue,
			},
		},
		conditionType: "Ready",
		output: &metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionTrue,
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := printer.FindCondition(test.conditions, test.conditionType)
			expected, actual := test.output, output
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
