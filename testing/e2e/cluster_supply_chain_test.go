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
	"context"
	"regexp"
	"strings"
	"testing"

	it "github.com/vmware-tanzu/apps-cli-plugin/testing/suite"
)

func TestClusterSupplyChain(t *testing.T) {
	testSuite := it.CommandLineIntegrationTestSuite{
		{
			Name:    "List the existing supply chains",
			Command: *it.NewTanzuAppsCommandLine("cluster-supply-chain", "list"),
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				match, err := regexp.MatchString(`.*.\nci-test[ ]{3,}[<>a-zA-Z]{5,}[ ]{3,}[0-9]{1,}s\n.*`, output)
				if err != nil {
					t.Error("Error while validating the output", err)
					t.FailNow()
				}
				if !match {
					t.Errorf("Expected 'ci-test   <status>   0s' to be present in\n%s", output)
					t.FailNow()
				}
				expectedFooter := `
To view details: "tanzu apps cluster-supply-chain get <name>"

`
				if !strings.HasSuffix(output, expectedFooter) {
					t.Errorf("Expected %s to be present in the output", expectedFooter)
					t.FailNow()
				}

			},
		},
		{
			Name:    "Get the existing supply chain",
			Command: *it.NewTanzuAppsCommandLine("cluster-supply-chain", "get", "ci-test"),
			Verify: func(ctx context.Context, t *testing.T, output string, err error) {
				expectedHeader := "---\n# ci-test:"
				if !strings.HasPrefix(output, expectedHeader) {
					t.Errorf("Expected %s to be present in the output", expectedHeader)
					t.FailNow()
				}
				match, err := regexp.MatchString(`.*.\n[ ]{3,}labels[ ]{3,}apps\.tanzu\.vmware\.com\/workload-type[ ]{3,}web\n.*`, output)
				if err != nil {
					t.Error("Error while validating the output", err)
					t.FailNow()
				}
				if !match {
					t.Errorf("Expected 'labels   apps.tanzu.vmware.com/workload-type   web' to be present in\n%s", output)
					t.FailNow()
				}

			},
		},
	}
	testSuite.Run(t)
}
