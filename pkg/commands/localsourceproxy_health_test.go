/*
Copyright 2023 VMware, Inc.

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

package commands_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestLocalSourceProxyHealthOptions_Validate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name: "valid options",
			Validatable: &commands.LocalSourceProxyHealthOptions{
				Output: "json",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid options",
			Validatable: &commands.LocalSourceProxyHealthOptions{
				Output: "text",
			},
			ExpectFieldErrors: validation.EnumInvalidValue("text", flags.OutputFlagName, []string{"json", "yaml", "yml"}),
		},
		{
			Name:           "no input",
			Validatable:    &commands.LocalSourceProxyHealthOptions{},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestNewLocalSourceProxyHealthCommand(t *testing.T) {
	scheme := runtime.NewScheme()
	var cmd *cobra.Command

	respCreator := func(status int, body string) *http.Response {
		return &http.Response{
			Status:     http.StatusText(status),
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
		}
	}

	table := clitesting.CommandTestSuite{
		{
			Name:                "empty output",
			Args:                []string{},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`)),
			ExpectOutput: `
user_has_permission: true
reachable: true
upstream_authenticated: true
overall_health: true
message: All health checks passed
`,
		},
		{
			Name:                "output yaml",
			Args:                []string{flags.OutputFlagName, "yaml"},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`)),
			ExpectOutput: `
user_has_permission: true
reachable: true
upstream_authenticated: true
overall_health: true
message: All health checks passed
`,
		},
		{
			Name:                "output yml",
			Args:                []string{flags.OutputFlagName, "yml"},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`)),
			ExpectOutput: `
user_has_permission: true
reachable: true
upstream_authenticated: true
overall_health: true
message: All health checks passed
`,
		},
		{
			Name:                "output json",
			Args:                []string{flags.OutputFlagName, "json"},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`)),
			ExpectOutput: `{
  "user_has_permission": true,
  "reachable": true,
  "upstream_authenticated": true,
  "overall_health": true,
  "message": "All health checks passed"
}`,
		},
		{
			Name:        "no transport error",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:                "error from cluster",
			Args:                []string{},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusUnauthorized, `{"message": "401 Status Unauthorized"}`)),
			ExpectOutput: `
user_has_permission: false
reachable: false
upstream_authenticated: false
overall_health: false
message: |-
  The current user does not have permission to access the local source proxy.
  Messages:
  - 401 Status Unauthorized
`,
		},
		{
			Name:                "error from registry",
			Args:                []string{},
			KubeConfigTransport: clitesting.NewFakeTransportFromResponse(respCreator(http.StatusOK, `{"statuscode": "401", "message": "401 Status user UNAUTHORIZED for registry"}`)),
			ExpectOutput: `
user_has_permission: true
reachable: true
upstream_authenticated: false
overall_health: false
message: |-
  Local source proxy was unable to authenticate against the target registry.
  Messages:
  - 401 Status user UNAUTHORIZED for registry
`,
		},
	}

	table.Run(t, scheme, func(ctx context.Context, c *cli.Config) *cobra.Command {
		cmd = commands.NewLocalSourceProxyHealthCommand(ctx, c)
		return cmd
	})
}
