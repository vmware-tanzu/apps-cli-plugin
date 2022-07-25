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
	"testing"
	"time"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

func TestTimestampSince(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	now := time.Now()

	tests := []struct {
		name   string
		input  metav1.Time
		output string
	}{{
		name:   "empty",
		output: printer.Swarnf("<unknown>"),
	}, {
		name:   "now",
		input:  metav1.Time{Time: now},
		output: "0s",
	}, {
		name:   "1 minute ago",
		input:  metav1.Time{Time: now.Add(-1 * time.Minute)},
		output: "60s",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, printer.TimestampSince(test.input, now); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}

func TestEmptyString(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name   string
		input  string
		output string
	}{{
		name:   "empty",
		output: printer.Sfaintf("<empty>"),
	}, {
		name:   "not empty",
		input:  "hello",
		output: "hello",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, printer.EmptyString(test.input); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}

func TestConditionStatus(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name   string
		input  *metav1.Condition
		output string
	}{{
		name:   "empty",
		output: printer.Swarnf("<unknown>"),
	}, {
		name: "status true",
		input: &metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionTrue,
		},
		output: printer.Ssuccessf("Ready"),
	}, {
		name: "status false",
		input: &metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionFalse,
			Reason: "uh-oh",
		},
		output: printer.Serrorf("uh-oh"),
	}, {
		name: "status false, no reason",
		input: &metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionFalse,
		},
		output: printer.Serrorf("not-Ready"),
	}, {
		name: "status unknown",
		input: &metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionUnknown,
		},
		output: printer.Sinfof("Unknown"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, printer.ConditionStatus(test.input); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}

func TestColorConditionStatus(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name   string
		input  string
		output string
	}{{
		name:   "empty",
		output: "",
	}, {
		name:   "status true",
		input:  "True",
		output: printer.Ssuccessf("True"),
	}, {
		name:   "status false",
		input:  "False",
		output: printer.Serrorf("False"),
	}, {
		name:   "status unknown",
		input:  "Unknown",
		output: printer.Sinfof("Unknown"),
	}, {
		name:   "status unknown",
		input:  "SomeOtherStatus",
		output: printer.Sinfof("SomeOtherStatus"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, printer.ColorConditionStatus(test.input); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}

func TestLabels(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name   string
		input  map[string]string
		output string
	}{{
		name:   "empty",
		output: printer.Sfaintf("<empty>"),
	}, {
		name:   "not empty",
		input:  map[string]string{"hello": "hi", "foo": "bar"},
		output: "foo=bar,hello=hi",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, printer.Labels(test.input); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}
