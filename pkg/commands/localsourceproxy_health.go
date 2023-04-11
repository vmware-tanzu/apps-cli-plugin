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
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands/lsp"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type LocalSourceProxyHealthOptions struct {
	Output string
}

var (
	_ validation.Validatable = (*LocalSourceProxyHealthOptions)(nil)
	_ cli.Executable         = (*LocalSourceProxyHealthOptions)(nil)
)

func (opts *LocalSourceProxyHealthOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}
	if opts.Output != "" {
		errs = errs.Also(validation.Enum(opts.Output, flags.OutputFlagName, []string{printer.OutputFormatJson, printer.OutputFormatYaml, printer.OutputFormatYml}))
	}
	return errs
}

func (opts *LocalSourceProxyHealthOptions) Exec(ctx context.Context, c *cli.Config) error {
	if s, err := lsp.GetStatus(ctx, c); err != nil {
		return err
	} else {
		printer.PrintLocalSourceProxyStatus(c.Stdout, opts.Output, s)
	}
	return nil
}

func NewLocalSourceProxyHealthCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &LocalSourceProxyHealthOptions{}

	cmd := &cobra.Command{
		Use:   "health",
		Short: "View status and health details for Local Source Proxy",
		Long:  strings.TrimSpace(`View status and health details for Local Source Proxy`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s local-source-proxy health --output json", c.Name),
		}, "\n"),
		PreRunE: cli.ValidateE(ctx, opts),
		RunE:    cli.ExecE(ctx, c, opts),
	}

	cmd.Flags().StringVarP(&opts.Output, cli.StripDash(flags.OutputFlagName), "o", "", "output the Local Source Proxy status formatted. Supported formats: \"json\", \"yaml\", \"yml\"")

	return cmd
}
