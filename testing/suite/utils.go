//go:build integration
// +build integration

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

package suite_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func GetFileAsString(t *testing.T, file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("Error trying to read file %q: %v", file, err)
		t.FailNow()
	}
	return string(b)
}

func GetWorkloadFromFile(t *testing.T, file string) *cartov1alpha1.Workload {
	b := GetFileAsString(t, file)
	r := strings.NewReader(b)
	workload := &cartov1alpha1.Workload{}
	workload.Load(r)
	return workload
}

func EmojisExistInOutput(output string, emojisList []cli.Icon) bool {
	decodedString := strconv.QuoteToASCII(output)
	decodedString = strings.Trim(decodedString, `"`)

	for _, e := range emojisList {
		decodedEmoji := strconv.QuoteToASCII(string(e))
		decodedEmoji = strings.Trim(decodedEmoji, `"`)
		if !strings.Contains(decodedString, decodedEmoji) {
			return false
		}
	}

	return true
}
