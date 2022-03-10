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

package completion_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
)

func TestSuggestDurationUnits(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedSuggestions []string
		units               []string
	}{
		{
			name:                "no input",
			input:               "",
			expectedSuggestions: []string{},
			units:               []string{"s", "m", "h"},
		},
		{
			name:                "int input",
			input:               "1",
			units:               []string{"s", "m", "h"},
			expectedSuggestions: []string{"1h", "1m", "1s"},
		},
		{
			name:                "float input",
			input:               "1.5",
			units:               []string{"s", "m", "h"},
			expectedSuggestions: []string{"1.5h", "1.5m", "1.5s"},
		},
		{
			name:                "valid input with int unit",
			input:               "1h",
			expectedSuggestions: []string{"1h", "1m", "1s"},
			units:               []string{"s", "m", "h"},
		},
		{
			name:                "valid input with float unit",
			input:               "1.6m",
			expectedSuggestions: []string{"1.6h", "1.6m", "1.6s"},
			units:               []string{"s", "m", "h"},
		},
		{
			name:                "valid input with multiple time units",
			input:               "1h2m",
			units:               []string{"s", "m", "h"},
			expectedSuggestions: []string{"1h2m", "1h2s"},
		},
		{
			name:                "multiple units with h",
			input:               "1h3",
			expectedSuggestions: []string{"1h3m", "1h3s"},
			units:               []string{"s", "m", "h"},
		},
		{
			name:                "invalid time unit",
			input:               "2foo3",
			expectedSuggestions: []string{"2foo3h", "2foo3m", "2foo3s"},
			units:               []string{"s", "m", "h"},
		},
		{
			name:                "time unit without integer",
			input:               "m",
			expectedSuggestions: []string{"m", "ms", "ns", "s"},
			units:               []string{"ms", "ns", "s", "m"},
		},
		{
			name:                "no lower time unit",
			input:               "3s2",
			expectedSuggestions: []string{},
			units:               []string{"s", "m", "h"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := &cobra.Command{}
			suggestions, directive := completion.SuggestDurationUnits(ctx, test.units)(cmd, []string{}, test.input)
			if !cmp.Equal(test.expectedSuggestions, suggestions) {
				t.Errorf("SuggestDurationUnits() suggestions want %v, got %v", test.expectedSuggestions, suggestions)
			}
			if want, got := cobra.ShellCompDirectiveNoFileComp, directive; want != got {
				t.Errorf("SuggestDurationUnits() ShellCompDirective: want %d, got %d", want, got)
			}
		})
	}
}
