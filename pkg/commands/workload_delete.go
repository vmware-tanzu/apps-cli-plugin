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
	"io"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/wait"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type WorkloadDeleteOptions struct {
	Namespace string
	Names     []string
	All       bool

	FilePath string

	Wait        bool
	WaitTimeout time.Duration
	Yes         bool
}

var (
	_ validation.Validatable = (*WorkloadDeleteOptions)(nil)
	_ cli.Executable         = (*WorkloadDeleteOptions)(nil)
)

func (opts *WorkloadDeleteOptions) Validate(_ context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(validation.ErrMissingField(flags.NamespaceFlagName))
	}

	if opts.All && len(opts.Names) != 0 {
		errs = errs.Also(validation.ErrMultipleOneOf(flags.AllFlagName, cli.NamesArgumentName))
	}

	if opts.All && opts.FilePath != "" {
		errs = errs.Also(validation.ErrMultipleOneOf(flags.AllFlagName, flags.FilePathFlagName))
	}

	if opts.FilePath == "" && !opts.All && len(opts.Names) == 0 {
		errs = errs.Also(validation.ErrMissingOneOf(flags.AllFlagName, cli.NamesArgumentName, flags.FilePathFlagName))
	}

	return errs
}

func (opts *WorkloadDeleteOptions) Exec(ctx context.Context, c *cli.Config) error {
	workload := &cartov1alpha1.Workload{}
	names := opts.Names

	if opts.FilePath != "" {
		fileWorkload := &cartov1alpha1.Workload{}

		if err := opts.loadInputWorkload(c.Stdin, fileWorkload); err != nil {
			return err
		}

		if fileWorkload.Namespace != "" && !cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.NamespaceFlagName)) {
			opts.Namespace = fileWorkload.Namespace
		}
		if fileWorkload.Name != "" {
			names = append(names, fileWorkload.Name)
		}
	}

	if opts.All {
		if !opts.Yes {
			if opts.FilePath == "-" {
				c.Errorf("Skipping workload, cannot confirm intent. Run command with %s flag to confirm intent when providing input from stdin\n", flags.YesFlagName)
				return nil
			} else {
				okToDeleteAll := false
				err := survey.AskOne(&survey.Confirm{
					Message: fmt.Sprintf("Really delete all workloads in the namespace %q?", opts.Namespace),
				}, &okToDeleteAll, printer.WithSurveyStdio(c.Stdin, c.Stdout, c.Stderr))

				if err != nil || !okToDeleteAll {
					c.Infof("Skipping workloads in namespace %q\n", opts.Namespace)
					return nil
				}
			}
		}
		err := c.DeleteAllOf(ctx, workload, client.InNamespace(opts.Namespace))
		if err != nil {
			return err
		}
		c.Successf("Deleted workloads in namespace %q\n", opts.Namespace)
		return nil
	}

	for _, name := range names {
		if err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: name}, workload); err != nil {
			if apierrs.IsNotFound(err) {
				c.Infof("Workload %q does not exist\n", name)
				continue
			}
			return err
		}
		if !opts.Yes {
			if opts.FilePath == "-" {
				c.Errorf("Skipping workload, cannot confirm intent. Run command with %s flag to confirm intent when providing input from stdin\n", flags.YesFlagName)
				return nil
			} else {
				okToDelete := false
				err := survey.AskOne(&survey.Confirm{
					Message: fmt.Sprintf("Really delete the workload %q?", name),
				}, &okToDelete, printer.WithSurveyStdio(c.Stdin, c.Stdout, c.Stderr))

				if err != nil || !okToDelete {
					c.Infof("Skipping workload %q\n", name)
					continue
				}
			}
		}
		if err := c.Delete(ctx, workload); err != nil {
			return err
		}
		c.Successf("Deleted workload %q\n", name)
		if opts.Wait {
			c.Infof("Waiting for workload %q to be deleted...\n", name)
			workers := []wait.Worker{
				func(ctx context.Context) error {
					return wait.UntilDelete(ctx, c.Client, workload)
				},
			}
			if err := wait.Race(ctx, opts.WaitTimeout, workers); err != nil {
				if err == context.DeadlineExceeded {
					c.Printf("%s timeout after %s waiting for %q to be deleted\n", printer.Serrorf("Error:"), opts.WaitTimeout, name)
					c.Infof("To view status run: tanzu apps workload get %s %s %s\n", name, flags.NamespaceFlagName, opts.Namespace)
					return cli.SilenceError(err)
				}
				c.Eprintf("%s %s\n", printer.Serrorf("Error:"), err)
				return cli.SilenceError(err)
			}
			c.Infof("Workload %q was deleted\n", name)
		}
	}

	return nil
}

func (opts *WorkloadDeleteOptions) loadInputWorkload(input io.Reader, workload *cartov1alpha1.Workload) error {
	var in io.Reader

	f, err := os.Open(opts.FilePath)
	in = f
	if f == nil && opts.FilePath == "-" {
		in = input
	} else if err != nil {
		return fmt.Errorf("unable to open file %q: %w", opts.FilePath, err)
	}
	defer f.Close()

	if err := workload.Load(in); err != nil {
		return fmt.Errorf("unable to load file %q: %w", opts.FilePath, err)
	}
	return nil
}

func NewWorkloadDeleteCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete workload(s)",
		Long: strings.TrimSpace(`
Delete one or more workloads by name or all workloads within a namespace.

Deleting a workload prevents new builds while preserving built images in the
registry.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload delete my-workload", c.Name),
			fmt.Sprintf("%s workload delete %s", c.Name, flags.AllFlagName),
		}, "\n"),
		PreRunE:           cli.ValidateE(ctx, opts),
		RunE:              cli.ExecE(ctx, c, opts),
		ValidArgsFunction: completion.SuggestWorkloadNames(ctx, c),
	}

	cli.Args(cmd,
		cli.NamesArg(&opts.Names),
	)

	cli.NamespaceFlag(ctx, cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.All, cli.StripDash(flags.AllFlagName), false, "delete all workloads within the namespace")
	cmd.Flags().BoolVar(&opts.Wait, cli.StripDash(flags.WaitFlagName), false, "waits for workload to be deleted")
	cmd.Flags().DurationVar(&opts.WaitTimeout, cli.StripDash(flags.WaitTimeoutFlagName), 1*time.Minute, "timeout for workload to be deleted when waiting")
	cmd.RegisterFlagCompletionFunc(cli.StripDash(flags.WaitTimeoutFlagName), completion.SuggestDurationUnits(ctx, completion.CommonDurationUnits))
	cmd.Flags().BoolVarP(&opts.Yes, cli.StripDash(flags.YesFlagName), "y", false, "accept all prompts")
	cmd.Flags().StringVarP(&opts.FilePath, cli.StripDash(flags.FilePathFlagName), "f", "", "`file path` containing the description of a single workload, other flags are layered on top of this resource. Use value \"-\" to read from stdin")

	return cmd
}
