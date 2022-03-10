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

package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func TestDryRunResource(t *testing.T) {
	stdout := &bytes.Buffer{}
	ctx := cli.WithStdout(context.Background(), stdout)
	resource := &rtesting.TestResource{}

	cli.DryRunResource(ctx, resource, rtesting.GroupVersion.WithKind("TestResource"))

	expected := strings.TrimSpace(`
---
apiVersion: testing.reconciler.runtime/v1
kind: TestResource
metadata:
  creationTimestamp: null
spec:
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
status: {}
`)
	actual := strings.TrimSpace(stdout.String())
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Unexpected stdout (-expected, +actual): %s", diff)
	}

}
