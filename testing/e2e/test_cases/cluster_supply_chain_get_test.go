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
	"testing"

	"github.com/vmware-tanzu/apps-cli-plugin/testing/e2e/helpers"
	it "github.com/vmware-tanzu/apps-cli-plugin/testing/e2e/suite"
)

func TestGetClusterSupplyChain(t *testing.T) {
	testSuite := it.CommandLineIntegrationTestSuite{
		{
			Name:                      "Get the existing supply chain",
			Command:                   *it.NewTanzuAppsCommandLine("cluster-supply-chain", "get", "ci-test"),
			ExpectedCommandLineOutput: helpers.GetFileAsString(t, filepath.Join(consoleOutBasePath, "get-csc", "test-get-ci-test-csc.txt")),
		},
	}
	testSuite.Run(t)
}
