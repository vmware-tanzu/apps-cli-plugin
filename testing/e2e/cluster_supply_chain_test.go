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

package integration_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	it "github.com/vmware-tanzu/apps-cli-plugin/testing/suite"
)

func TestClusterSupplyChain(t *testing.T) {
	testSuite := it.CommandLineIntegrationTestSuite{
		{
			Name:    "List the existing supply chains",
			Command: *it.NewTanzuAppsCommandLine("cluster-supply-chain", "list"),
			Verify: func(t *testing.T, output string, err error) {
				match, err := regexp.MatchString(`.*.\nci-test[ ]{3,}[<>a-zA-Z]{5,}[ ]{3,}[0-9]{1,}s\n.*`, output)
				if !match {
					t.Error("Expected 'ci-test   <status>   0s' to be present in the output")
					t.FailNow()
				}
				if err != nil {
					t.Error("Error while validating the output", err)
					t.FailNow()
				}
				const outFoot = `
To view details: "tanzu apps cluster-supply-chain get <name>"

`
				if !strings.HasSuffix(output, outFoot) {
					t.Errorf("Expected %s to be present in the output", outFoot)
					t.FailNow()
				}

			},
		},
		{
			Name:    "Get the existing supply chain",
			Command: *it.NewTanzuAppsCommandLine("cluster-supply-chain", "get", "ci-test"),
			Verify: func(t *testing.T, output string, err error) {
				const outHead = "---\n# ci-test:"
				if !strings.HasPrefix(output, outHead) {
					t.Errorf("Expected %s to be present in the output", outHead)
					t.FailNow()
				}
				valContent := it.GetFileAsString(t, filepath.Join(it.ConsoleOutBasePath, "get-csc", "test-get-ci-test-csc.txt"))
				if diff := cmp.Diff(valContent, output[len(output)-len(valContent):]); diff != "" {
					t.Errorf("%s(Get Supply Chain Selectors)\n(-expected, +actual)\n%s", t.Name(), diff)
					t.FailNow()
				}

			},
		},
	}
	testSuite.Run(t)
}
