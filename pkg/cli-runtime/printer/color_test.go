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

	"github.com/fatih/color"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

func TestScolorf(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name    string
		format  string
		args    []interface{}
		printer func(format string, a ...interface{}) string
		output  string
	}{{
		name:    "Sfaintf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: printer.Sfaintf,
		output:  printer.FaintColor.Sprint("hello"),
	}, {
		name:    "Sinfof",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: printer.Sinfof,
		output:  printer.InfoColor.Sprint("hello"),
	}, {
		name:    "Ssuccessf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: printer.Ssuccessf,
		output:  printer.SuccessColor.Sprint("hello"),
	}, {
		name:    "Swarnf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: printer.Swarnf,
		output:  printer.WarnColor.Sprint("hello"),
	}, {
		name:    "Serrorf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: printer.Serrorf,
		output:  printer.ErrorColor.Sprint("hello"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, test.printer(test.format, test.args...); expected != actual {
				t.Errorf("Expected output to be %q, actually %q", expected, actual)
			}
		})
	}
}
