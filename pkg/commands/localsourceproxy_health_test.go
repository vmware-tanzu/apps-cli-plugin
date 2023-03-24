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

import (
	"context"
	"reflect"
	"testing"

	"github.com/spf13/cobra"

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
	type args struct {
		ctx context.Context
		c   *cli.Config
	}
	tests := []struct {
		name string
		args args
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commands.NewLocalSourceProxyHealthCommand(tt.args.ctx, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLocalSourceProxyHealthCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
