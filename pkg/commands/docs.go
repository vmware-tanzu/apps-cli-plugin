/*
Copyright 2021 VMware, Inc.

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
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

type DocsOptions struct {
	Directory string
}

var (
	_ validation.Validatable = (*DocsOptions)(nil)
)

func (opts *DocsOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Directory == "" {
		errs = errs.Also(validation.ErrMissingField("--directory"))
	}

	return errs
}

func NewDocsCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &DocsOptions{}

	cmd := &cobra.Command{
		Use:     "docs",
		Short:   "generate docs in Markdown for this CLI",
		Example: fmt.Sprintf("%s docs", c.Name),
		Hidden:  true,
		PreRunE: cli.ValidateE(ctx, opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(opts.Directory, 0744); err != nil {
				return err
			}

			if noColorFlag := cmd.Root().Flag(cli.StripDash(flags.NoColorFlagName)); noColorFlag != nil {
				// force default to false for doc generation no matter the environment
				noColorFlag.DefValue = "false"
			}

			root := &cobra.Command{
				Use:               "tanzu",
				DisableAutoGenTag: true,
			}
			root.AddCommand(cmd.Root())

			// hack to rewrite the CommandPath content to add args
			cli.Visit(root, func(cmd *cobra.Command) error {
				if !cmd.HasSubCommands() {
					cmd.Use = cmd.Use + cli.FormatArgs(cmd)
				}
				return nil
			})

			if err := doc.GenMarkdownTree(root, opts.Directory); err != nil {
				return err
			}

			// remove synthetic root command
			if err := os.Remove(path.Join(opts.Directory, "tanzu.md")); err != nil {
				return err
			}
			return filepath.Walk(opts.Directory, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				input, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				inlines := strings.Split(string(input), "\n")
				outlines := []string{}
				for _, line := range inlines {
					if !strings.HasPrefix(line, "* [tanzu](tanzu.md)") {
						outlines = append(outlines, line)
					}
				}
				return ioutil.WriteFile(path, []byte(strings.Join(outlines, "\n")), 0644)
			})
		},
	}

	cmd.Flags().StringVarP(&opts.Directory, "directory", "d", "docs", "the output `directory` for the docs")

	return cmd
}
