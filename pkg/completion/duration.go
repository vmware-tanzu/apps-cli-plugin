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

package completion

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	AllDurationUnits    = []string{"h", "m", "s", "ms", "us", "ns", "µs"}
	CommonDurationUnits = []string{"h", "m", "s"}
	extractAlphabets    = regexp.MustCompile(`[a-zA-Z]+(.*?)`)
)

func SuggestDurationUnits(ctx context.Context, units []string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if toComplete == "" {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		toComplete = trimUnitSuffix(toComplete)
		lastUnit := lastDurationUnit(toComplete)
		suggestUnits := intersectionOfStrings(units, shorterTimeUnits(lastUnit))
		suggestions := []string{}
		for _, unit := range suggestUnits {
			suggestions = append(suggestions, fmt.Sprintf("%s%s", toComplete, unit))
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}
}

// return last unit time  ex (1h30m2 -> m)
func lastDurationUnit(duration string) string {
	units := extractAlphabets.FindAllString(duration, -1)
	if len(units) == 0 {
		return ""
	}
	return units[len(units)-1]
}

// trimUnitSuffix trims the original string after the last digit
func trimUnitSuffix(original string) string {
	units := extractAlphabets.FindAllString(original, -1)
	if len(units) == 0 {
		return original
	}
	return strings.TrimSuffix(original, units[len(units)-1])
}

// shorterTimeUnits gets units smaller than inputUnit index in AllDurationUnits
func shorterTimeUnits(inputUnit string) []string {
	index := -1
	for i, unit := range AllDurationUnits {
		if unit == inputUnit {
			index = i
			break
		}
	}
	return AllDurationUnits[index+1:]
}

func intersectionOfStrings(input1, input2 []string) []string {
	return sets.NewString(input1...).
		Intersection(sets.NewString(input2...)).
		List()
}
