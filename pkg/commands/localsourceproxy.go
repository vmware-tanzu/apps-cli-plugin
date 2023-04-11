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

package commands

import (
	"context"

	"github.com/spf13/cobra"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func NewLocalSourceProxyCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "local-source-proxy",
		Hidden:  true,
		Short:   "View status and health details for Local Source Proxy",
		Aliases: []string{"lsp"},
	}

	cmd.AddCommand(NewLocalSourceProxyHealthCommand(ctx, c))

	return cmd
}
