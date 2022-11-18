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
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type WorkloadApplyOptions struct {
	WorkloadOptions
	UpdateStrategy string
}

var (
	_ validation.Validatable = (*WorkloadApplyOptions)(nil)
	_ cli.Executable         = (*WorkloadApplyOptions)(nil)
	_ cli.DryRunable         = (*WorkloadApplyOptions)(nil)
)

const (
	mergeUpdateStrategy   = "merge"
	replaceUpdateStrategy = "replace"
)

func (opts *WorkloadApplyOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}
	errs = errs.Also(opts.WorkloadOptions.Validate(ctx))

	if opts.UpdateStrategy != "" && cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.UpdateStrategyFlagName)) {
		if opts.FilePath == "" {
			errs = errs.Also(validation.ErrMissingField(flags.FilePathFlagName))
		}
		errs = errs.Also(validation.Enum(opts.UpdateStrategy, flags.UpdateStrategyFlagName, []string{mergeUpdateStrategy, replaceUpdateStrategy}))
	}

	return errs
}

func (opts *WorkloadApplyOptions) Exec(ctx context.Context, c *cli.Config) error {
	var createError error
	var updateError error
	okToCreate := false
	okToUpdate := false

	fileWorkload := &cartov1alpha1.Workload{}
	if opts.FilePath != "" {
		c.Emoji(cli.Exclamation, fmt.Sprintf("WARNING: Configuration file update strategy is changing. By default, provided configuration files will replace rather than merge existing configuration. The change will take place in the January 2024 TAP release (use %q to control strategy explicitly).\n\n", flags.UpdateStrategyFlagName))
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
	var currentWorkload *cartov1alpha1.Workload
	err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: opts.Name}, workload)
	if err == nil {
		currentWorkload = workload.DeepCopy()
	} else {
		if !apierrs.IsNotFound(err) {
			return err
		}
		if apierrs.IsNotFound(err) {
			if nsErr := validateNamespace(ctx, c, opts.Namespace); nsErr != nil {
				return nsErr
			}
		}
	}

	if opts.UpdateStrategy == mergeUpdateStrategy {
		if opts.FilePath != "" {
			var serviceAccountCopy string
			// avoid passing a nil pointer to MergeServiceAccountName func
			if fileWorkload.Spec.ServiceAccountName != nil {
				serviceAccountCopy = *fileWorkload.Spec.ServiceAccountName
			}

			workload.Spec.MergeServiceAccountName(serviceAccountCopy)
		}
		workload.Merge(fileWorkload)
	}

	if opts.UpdateStrategy == replaceUpdateStrategy {
		// assign all the file workload fields to the workload in the cluster
		workload = fileWorkload

		// if there is a workload in the cluster with all metadata populated
		// re assign the system populated fields so we won't find an error because of some missing fields
		if currentWorkload != nil && !reflect.DeepEqual(currentWorkload.ObjectMeta, (metav1.ObjectMeta{})) {
			wldMeta := workload.GetObjectMeta()
			currentWldMeta := currentWorkload.GetObjectMeta()

			// assign the system populated fields to the new workload
			wldMeta.SetResourceVersion(currentWldMeta.GetResourceVersion())
			wldMeta.SetUID(currentWldMeta.GetUID())
			wldMeta.SetGeneration(currentWldMeta.GetGeneration())
			wldMeta.SetCreationTimestamp(currentWldMeta.GetCreationTimestamp())
			wldMeta.SetDeletionTimestamp(currentWldMeta.GetDeletionTimestamp())
		}
	}

	workload.Name = opts.Name
	workload.Namespace = opts.Namespace
	ctx = opts.ApplyOptionsToWorkload(ctx, workload)

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

	// If user answers yes to survey prompt about publishing source, continue with creation or update
	if okToPush, err := opts.PublishLocalSource(ctx, c, currentWorkload, workload); err != nil {
		return err
	} else if !okToPush {
		return nil
	}

	// If there is no workload, create a new one
	if currentWorkload == nil {
		okToCreate, createError = opts.Create(ctx, c, workload)
		if createError != nil {
			return createError
		}
	} else {
		okToUpdate, updateError = opts.Update(ctx, c, currentWorkload, workload)
		if updateError != nil {
			return updateError
		}
	}

	if okToCreate || okToUpdate {
		c.Printf("\n")
		DisplayCommandNextSteps(c, workload)
		c.Printf("\n")
	}

	anyTail := opts.Tail || opts.TailTimestamps
	if (okToCreate || okToUpdate) && (opts.Wait || anyTail) {
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
				return logs.Tail(ctx, c, opts.Namespace, selector, containers, time.Minute, opts.TailTimestamps)
			})
		}

		if err := wait.Race(ctx, opts.WaitTimeout, workers); err != nil {
			if err == context.DeadlineExceeded {
				c.Printf("%s timeout after %s waiting for %q to become ready\n", printer.Serrorf("Error:"), opts.WaitTimeout, workload.Name)
				return cli.SilenceError(err)
			}
			c.Eprintf("%s %s\n", printer.Serrorf("Error:"), err)
			return cli.SilenceError(err)
		}
		c.Infof("Workload %q is ready\n", workload.Name)
	}
	return nil
}

func (opts *WorkloadApplyOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewWorkloadApplyCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadApplyOptions{}
	opts.LoadDefaults(c)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply configuration to a new or existing workload",
		Long: strings.TrimSpace(`
Apply configuration to a new or existing workload. If the resource does not exist, it will be created.

Workload configuration options include:
- source code to build
- runtime resource limits
- environment variables
- services to bind
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload apply %s workload.yaml", c.Name, flags.FilePathFlagName),
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
	cmd.Flags().StringVar(&opts.UpdateStrategy, cli.StripDash(flags.UpdateStrategyFlagName), mergeUpdateStrategy, "specify configuration file update strategy (supported strategies: merge, replace)")
	cmd.RegisterFlagCompletionFunc(cli.StripDash(flags.UpdateStrategyFlagName), func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{replaceUpdateStrategy, mergeUpdateStrategy}, cobra.ShellCompDirectiveNoFileComp
	})

	// Bind flags to environment variables
	opts.DefineEnvVars(ctx, c, cmd)

	return cmd
}
