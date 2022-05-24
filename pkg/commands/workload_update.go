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
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/logs"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/wait"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/watch"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type WorkloadUpdateOptions struct {
	WorkloadOptions
}

var (
	_ validation.Validatable = (*WorkloadUpdateOptions)(nil)
	_ cli.Executable         = (*WorkloadUpdateOptions)(nil)
	_ cli.DryRunable         = (*WorkloadUpdateOptions)(nil)
)

func (opts *WorkloadUpdateOptions) Validate(ctx context.Context) validation.FieldErrors {
	return opts.WorkloadOptions.Validate(ctx)
}

func (opts *WorkloadUpdateOptions) Exec(ctx context.Context, c *cli.Config) error {
	fileWorkload := &cartov1alpha1.Workload{}
	if opts.FilePath != "" {
		if err := opts.WorkloadOptions.LoadInputWorkload(c.Stdin, fileWorkload); err != nil {
			return err
		}

		if opts.Name == "" {
			opts.Name = fileWorkload.Name
		}
		if fileWorkload.Namespace != "" && !cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.NamespaceFlagName)) {
			opts.Namespace = fileWorkload.Namespace
		}
	}

	// validate that a namespace and name are provided
	errs := validation.FieldErrors{}
	if opts.Name == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NameArgumentName))
	}
	if opts.Namespace == "" {
		errs = errs.Also(validation.ErrMissingField(flags.NamespaceFlagName))
	}
	if err := errs.ToAggregate(); err != nil {
		return err
	}

	workload := &cartov1alpha1.Workload{}
	err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: opts.Name}, workload)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}
		nsGet := &corev1.Namespace{}
		if getErr := c.Get(ctx, types.NamespacedName{Name: opts.Namespace}, nsGet); getErr != nil && apierrs.IsNotFound(getErr) {
			c.Eprintf("%s %s\n", printer.Serrorf("Error:"), fmt.Sprintf("namespace %q not found, it may not exist or user does not have permissions to read it.", opts.Namespace))
			return cli.SilenceError(getErr)
		}
		c.Errorf("Workload %q not found\n", fmt.Sprintf("%s/%s", opts.Namespace, opts.Name))
		return cli.SilenceError(err)
	}
	currentWorkload := workload.DeepCopy()
	if opts.FilePath != "" {
		var serviceAccountCopy string
		// avoid passing a nil pointer to MergeServiceAccountName func
		if fileWorkload.Spec.ServiceAccountName != nil {
			serviceAccountCopy = *fileWorkload.Spec.ServiceAccountName
		}

		workload.Spec.MergeServiceAccountName(serviceAccountCopy)
	}
	workload.Merge(fileWorkload)

	opts.ApplyOptionsToWorkload(ctx, workload)

	// validate complex flag interactions with existing state
	errs = workload.Validate()
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

	// If user answers yes to survey prompt about publishing source, continue with workload update
	if okToPush, err := opts.PublishLocalSource(ctx, c, workload); err != nil {
		return err
	} else if !okToPush {
		return nil
	}

	okToUpdate, err := opts.Update(ctx, c, currentWorkload, workload)
	if err != nil {
		return err
	}

	anyTail := opts.Tail || opts.TailTimestamps
	if (opts.Yes || okToUpdate) && (opts.Wait || anyTail) {
		c.Infof("Waiting for workload %q to become ready...\n", opts.Name)

		workers := []wait.Worker{
			func(ctx context.Context) error {
				clientWithWatch, err := watch.GetWatcher(ctx, c)
				if err != nil {
					panic(err)
				}
				return wait.UntilCondition(ctx, clientWithWatch, types.NamespacedName{Name: workload.Name, Namespace: workload.Namespace}, &cartov1alpha1.WorkloadList{}, cartov1alpha1.WorkloadReadyConditionFunc)
			},
		}

		if anyTail {
			workers = append(workers, func(ctx context.Context) error {
				selector, err := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workload.Name))
				if err != nil {
					panic(err)
				}
				containers := []string{}
				return logs.Tail(ctx, c, opts.Namespace, selector, containers, time.Second, opts.TailTimestamps)
			})
		}

		if err := wait.Race(ctx, opts.WaitTimeout, workers); err != nil {
			if err == context.DeadlineExceeded {
				c.Printf("%s timeout after %s waiting for %q to become ready\n", printer.Serrorf("Error:"), opts.WaitTimeout, workload.Name)
				c.Infof("To view status run: tanzu apps workload get %s %s %s\n", workload.Name, flags.NamespaceFlagName, opts.Namespace)
				return cli.SilenceError(err)
			}
			c.Eprintf("%s %s\n", printer.Serrorf("Error:"), err)
			return cli.SilenceError(err)
		}
		c.Infof("Workload %q is ready\n", workload.Name)
	}
	return nil
}

func (opts *WorkloadUpdateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewWorkloadUpdateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadUpdateOptions{}
	opts.LoadDefaults(c)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update configuration of an existing workload",
		Long: strings.TrimSpace(`
Update configuration of an existing workload.

Workload configuration options include:
- source code to build
- runtime resource limits
- environment variables
- services to bind
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload update my-workload %s=false", c.Name, flags.DebugFlagName),
			fmt.Sprintf("%s workload update my-workload %s .", c.Name, flags.LocalPathFlagName),
			fmt.Sprintf("%s workload update my-workload %s %s", c.Name, flags.EnvFlagName, "key=value"),
			fmt.Sprintf("%s workload update my-workload %s %s", c.Name, flags.BuildEnvFlagName, "key=value"),
			fmt.Sprintf("%s workload update %s workload.yaml", c.Name, flags.FilePathFlagName),
		}, "\n"),
		PreRunE:           cli.ValidateE(ctx, opts),
		RunE:              cli.ExecE(ctx, c, opts),
		ValidArgsFunction: completion.SuggestWorkloadNames(ctx, c),
	}

	cli.Args(cmd,
		cli.OptionalNameArg(&opts.Name),
	)

	// Define common flags
	opts.DefineFlags(ctx, c, cmd)

	return cmd
}
