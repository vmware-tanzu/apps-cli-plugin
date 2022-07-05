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

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	it "github.com/vmware-tanzu/apps-cli-plugin/testing/e2e/suite"
)

var (
	consoleOutBasePath = filepath.Join("testdata", "console-out")
	expectedBasePath   = filepath.Join("testdata", "expected")
	srcBasePath        = filepath.Join("testdata", "src")
	emptyWorkload      = &cartov1alpha1.Workload{}
	emptyWorkloadList  = &cartov1alpha1.WorkloadList{}
	namespaceFlag      = "--namespace=" + it.TestingNamespace
)
