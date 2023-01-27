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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/logs"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/wait"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type WorkloadCreateOptions struct {
	WorkloadOptions
}

var (
	_ validation.Validatable = (*WorkloadCreateOptions)(nil)
	_ cli.Executable         = (*WorkloadCreateOptions)(nil)
	_ cli.DryRunable         = (*WorkloadCreateOptions)(nil)
)

func (opts *WorkloadCreateOptions) Validate(ctx context.Context) validation.FieldErrors {
	return opts.WorkloadOptions.Validate(ctx)
}

func (opts *WorkloadCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	workload := &cartov1alpha1.Workload{}

	if opts.FilePath != "" {
		if err := opts.WorkloadOptions.LoadInputWorkload(c.Stdin, workload); err != nil {
			return err
		}
	}

	if opts.Name != "" {
		workload.Name = opts.Name
	}
	if workload.Namespace == "" || cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.NamespaceFlagName)) {
		workload.Namespace = opts.Namespace
	}

	existingWorkload := &cartov1alpha1.Workload{}

	if err := c.Get(ctx, client.ObjectKey{Namespace: workload.Namespace, Name: workload.Name}, existingWorkload); err != nil {
		// return err, except when not found
		if !apierrs.IsNotFound(err) {
			return err
		} else if apierrs.IsNotFound(err) {
			if nsErr := validateNamespace(ctx, c, opts.Namespace); nsErr != nil {
				return err
			}
		}
	}

	// check if the workload exists
	if existingWorkload != nil {
		if existingWorkload.Name == workload.Name && existingWorkload.Namespace == workload.Namespace {
			c.Printf("%s workload %q already exists\n", printer.Serrorf("Error:"), fmt.Sprintf("%s/%s", workload.Namespace, workload.Name))
			return cli.SilenceError(errors.New(""))
		}
	}

	ctx = opts.ApplyOptionsToWorkload(ctx, workload)

	// validate complex flag interactions with existing state
	errs := workload.Validate()
	// local path requires a source image
	if opts.LocalPath != "" && (workload.Spec.Source == nil || workload.Spec.Source.Image == "") {
		errs = errs.Also(
			validation.ErrMissingField(flags.SourceImageFlagName),
		)
	}
	if err := errs.ToAggregate(); err != nil {
		// show command usage before error
		cli.CommandFromContext(ctx).SilenceUsage = false
		return err
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, workload, workload.GetGroupVersionKind())
		return nil
	}

	var okToCreate bool
	shouldPrint := opts.Output == "" || (opts.Output != "" && !opts.Yes)
	if shouldPrint {
		// If user answers yes to survey prompt about publishing source, continue with workload creation
		if okToPush, err := opts.PublishLocalSource(ctx, c, nil, workload, shouldPrint); err != nil {
			return err
		} else if !okToPush {
			return nil
		}

		var err error
		okToCreate, err = opts.Create(ctx, c, workload)
		if err != nil {
			return err
		}

		if okToCreate {
			c.Printf("\n")
			DisplayCommandNextSteps(c, workload)
			c.Printf("\n")
		}
	} else if opts.Output != "" && opts.Yes {
		// since there are no prompts, set okToApply to true (accepted through --yes)
		okToCreate = true
		if _, err := opts.PublishLocalSource(ctx, c, workload, workload, shouldPrint); err != nil {
			return err
		}
		if err := c.Create(ctx, workload); err != nil {
			return err
		}
	}

	var workers []wait.Worker
	if opts.Output != "" && okToCreate {
		workers = opts.WaitToBeReady(c, workload)
		if err := opts.WaitError(ctx, c, workload, workers); err != nil {
			return err
		}
		// once the workload is ready, get it as is in the cluster
		if err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: opts.Name}, workload); err != nil {
			return err
		}
		if err := opts.OutputWorkload(c, workload); err != nil {
			return err
		}

		return nil
	}

	anyTail := opts.Tail || opts.TailTimestamps
	if okToCreate && (opts.Wait || anyTail) {
		c.Infof("Waiting for workload %q to become ready...\n", opts.Name)

		workers = opts.WaitToBeReady(c, workload)

		if anyTail {
			workers = append(workers, func(ctx context.Context) error {
				selector, err := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workload.Name))
				if err != nil {
					panic(err)
				}
				containers := []string{}
				return logs.Tail(ctx, c, opts.Namespace, selector, containers, time.Minute, opts.TailTimestamps)
			})
		}

		if err := opts.WaitError(ctx, c, workload, workers); err != nil {
			return err
		}

		c.Infof("Workload %q is ready\n\n", workload.Name)
	}

	return nil
}

func (opts *WorkloadCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewWorkloadCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadCreateOptions{}
	opts.LoadDefaults(c)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a workload with specified configuration",
		Long: strings.TrimSpace(`
Create a workload with specified configuration.

Workload configuration options include:
- source code to build
- runtime resource limits
- environment variables
- services to bind
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload create my-workload %s https://example.com/my-workload.git %s my-branch", c.Name, flags.GitRepoFlagName, flags.GitBranchFlagName),
			fmt.Sprintf("%s workload create my-workload %s . %s registry.example/repository:tag", c.Name, flags.LocalPathFlagName, flags.SourceImageFlagName),
			fmt.Sprintf("%s workload create %s workload.yaml", c.Name, flags.FilePathFlagName),
		}, "\n"),
		PreRunE: cli.ValidateE(ctx, opts),
		RunE:    cli.ExecE(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.OptionalNameArg(&opts.Name),
	)

	// Define common flags
	opts.DefineFlags(ctx, c, cmd)

	// Bind flags to environment variables
	opts.DefineEnvVars(ctx, c, cmd)

	return cmd
}
