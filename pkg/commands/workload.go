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
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/parsers"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

const AnnotationReservedKey = "annotations"

func NewWorkloadCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workload",
		Short: "Workload lifecycle management",
		Long: strings.TrimSpace(`
A workload may run as a knative service, kubernetes deployment, or other runtime. Workloads can be grouped together with other related resources such as storage or credential objects as a logical application for easier management.

Workload configuration includes:
- source code to build
- runtime resource limits
- environment variables
- services to bind
`),
		Aliases: []string{"workloads"},
	}

	cmd.AddCommand(NewWorkloadListCommand(ctx, c))
	cmd.AddCommand(NewWorkloadGetCommand(ctx, c))
	cmd.AddCommand(NewWorkloadTailCommand(ctx, c))
	cmd.AddCommand(NewWorkloadCreateCommand(ctx, c))
	cmd.AddCommand(NewWorkloadUpdateCommand(ctx, c))
	cmd.AddCommand(NewWorkloadApplyCommand(ctx, c))
	cmd.AddCommand(NewWorkloadDeleteCommand(ctx, c))

	return cmd
}

type WorkloadOptions struct {
	Namespace string
	Name      string

	App         string
	Type        string
	Labels      []string
	Annotations []string
	Params      []string
	Debug       bool
	LiveUpdate  bool

	FilePath    string
	GitRepo     string
	GitCommit   string
	GitBranch   string
	GitTag      string
	SourceImage string
	LocalPath   string
	Image       string
	SubPath     string

	BuildEnv    []string
	Env         []string
	ServiceRefs []string

	LimitCPU    string
	LimitMemory string

	RequestCPU    string
	RequestMemory string

	Wait           bool
	WaitTimeout    time.Duration
	Tail           bool
	TailTimestamps bool
	DryRun         bool
	Yes            bool
}

var _ validation.Validatable = (*WorkloadUpdateOptions)(nil)

func (opts *WorkloadOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	errs = errs.Also(validation.K8sName(opts.Namespace, flags.NamespaceFlagName))
	if opts.FilePath == "" {
		errs = errs.Also(validation.K8sName(opts.Name, cli.NameArgumentName))
	}
	errs = errs.Also(validation.DeletableKeyValues(opts.Labels, flags.LabelFlagName))
	errs = errs.Also(validation.DeletableKeyValues(opts.Annotations, flags.AnnotationFlagName))
	errs = errs.Also(validation.DeletableKeyValues(opts.Params, flags.ParamFlagName))
	errs = errs.Also(validation.DeletableEnvVars(opts.Env, flags.EnvFlagName))
	errs = errs.Also(validation.DeletableEnvVars(opts.BuildEnv, flags.BuildEnvFlagName))
	errs = errs.Also(validation.DeletableKeyObjectReferences(opts.ServiceRefs, flags.ServiceRefFlagName))

	if opts.LimitCPU != "" {
		errs = errs.Also(validation.Quantity(opts.LimitCPU, flags.LimitCPUFlagName))
	}
	if opts.LimitMemory != "" {
		errs = errs.Also(validation.Quantity(opts.LimitMemory, flags.LimitMemoryFlagName))
	}

	if opts.RequestCPU != "" {
		errs = errs.Also(validation.Quantity(opts.RequestCPU, flags.RequestCPUFlagName))
	}
	if opts.RequestMemory != "" {
		errs = errs.Also(validation.Quantity(opts.RequestMemory, flags.RequestMemoryFlagName))
	}

	if opts.RequestCPU != "" && opts.LimitCPU != "" {
		errs = errs.Also(validation.CompareQuantity(opts.LimitCPU, opts.RequestCPU, flags.RequestCPUFlagName))
	}
	if opts.LimitMemory != "" && opts.RequestMemory != "" {
		errs = errs.Also(validation.CompareQuantity(opts.LimitMemory, opts.RequestMemory, flags.RequestMemoryFlagName))
	}

	// source options are mutually exclusive
	source := []string{}
	if opts.GitBranch != "" || opts.GitCommit != "" || opts.GitRepo != "" || opts.GitTag != "" {
		source = append(source, flags.GitFlagWildcard)
	}
	if opts.SourceImage != "" {
		source = append(source, flags.SourceImageFlagName)
	}
	if opts.Image != "" {
		source = append(source, flags.ImageFlagName)
	}
	if len(source) > 1 {
		errs = errs.Also(validation.ErrMultipleOneOf(source...))
	}

	return errs
}

func (opts *WorkloadOptions) ApplyOptionsToWorkload(ctx context.Context, workload *cartov1alpha1.Workload) {
	for _, label := range opts.Labels {
		parts := parsers.DeletableKeyValue(label)
		if len(parts) == 1 {
			delete(workload.Labels, parts[0])
		} else {
			workload.MergeLabels(parts[0], parts[1])
		}
	}
	for _, annotation := range opts.Annotations {
		kv := parsers.DeletableKeyValue(annotation)
		if len(kv) == 1 {
			workload.Spec.RemoveAnnotationParams(kv[0])
		} else {
			workload.Spec.MergeAnnotationParams(kv[0], kv[1])
		}
	}

	for _, p := range opts.Params {
		kv := parsers.DeletableKeyValue(p)
		if len(kv) == 1 {
			workload.Spec.RemoveParam(kv[0])
		} else {
			workload.Spec.MergeParams(kv[0], kv[1])
		}
	}

	if opts.App != "" {
		workload.MergeLabels(apis.AppPartOfLabelName, opts.App)
	}

	if opts.Type != "" {
		workload.MergeLabels(apis.WorkloadTypeLabelName, opts.Type)
	}

	if opts.Debug {
		workload.Spec.MergeParams("debug", "true")
	} else if cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.DebugFlagName)) {
		// debug was actively disabled
		workload.Spec.RemoveParam("debug")
	}

	if opts.LiveUpdate {
		workload.Spec.MergeParams("live-update", "true")
	} else if cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.LiveUpdateFlagName)) {
		// live-update was actively disabled
		workload.Spec.RemoveParam("live-update")
	}

	if opts.GitRepo != "" || opts.GitBranch != "" || opts.GitCommit != "" || opts.GitTag != "" {
		workload.Spec.MergeGit(cartov1alpha1.GitSource{
			URL: opts.GitRepo,
			Ref: cartov1alpha1.GitRef{
				Branch: opts.GitBranch,
				Commit: opts.GitCommit,
				Tag:    opts.GitTag,
			},
		})
	}

	if opts.SourceImage != "" {
		workload.Spec.MergeSourceImage(opts.SourceImage)
	}

	if cli.CommandFromContext(ctx).Flags().Changed(cli.StripDash(flags.SubPathFlagName)) {
		workload.Spec.MergeSubPath(opts.SubPath)
	}

	if opts.Image != "" {
		workload.Spec.MergeImage(opts.Image)
	}

	for _, ev := range opts.Env {
		env, delete := parsers.DeletableEnvVar(ev)
		if delete {
			workload.Spec.RemoveEnv(env.Name)
		} else {
			workload.Spec.MergeEnv(env)
		}
	}

	for _, ev := range opts.BuildEnv {
		env, delete := parsers.DeletableEnvVar(ev)
		if delete {
			workload.Spec.RemoveBuildEnv(env.Name)
		} else {
			workload.Spec.MergeBuildEnv(env)
		}
	}

	for _, ref := range opts.ServiceRefs {
		parts := parsers.DeletableKeyValue(ref)
		serviceRefKey := parts[0]
		if len(parts) == 1 {
			workload.Spec.DeleteServiceClaim(serviceRefKey)
			workload.DeleteServiceClaimAnnotation(serviceRefKey)
		} else {
			deleteValue := parts[1]
			workload.Spec.MergeServiceClaim(cartov1alpha1.NewServiceClaim(serviceRefKey, parsers.ObjectReference(deleteValue)))
			serviceClaimAnnotationValue := parsers.ObjectReferenceAnnotation(deleteValue)
			if serviceClaimAnnotationValue != nil {
				workload.MergeServiceClaimAnnotation(serviceRefKey, serviceClaimAnnotationValue)
			} else {
				workload.DeleteServiceClaimAnnotation(serviceRefKey)
			}
		}
	}

	if opts.LimitCPU != "" {
		workload.Spec.MergeResources(&corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				// parse errors are handled by the opt validation
				corev1.ResourceCPU: resource.MustParse(opts.LimitCPU),
			},
		})
	}

	if opts.LimitMemory != "" {
		workload.Spec.MergeResources(&corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				// parse errors are handled by the opt validation
				corev1.ResourceMemory: resource.MustParse(opts.LimitMemory),
			},
		})
	}

	if opts.RequestCPU != "" {
		workload.Spec.MergeResources(&corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				// parse errors are handled by the opt validation
				corev1.ResourceCPU: resource.MustParse(opts.RequestCPU),
			},
		})
	}

	if opts.RequestMemory != "" {
		workload.Spec.MergeResources(&corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				// parse errors are handled by the opt validation
				corev1.ResourceMemory: resource.MustParse(opts.RequestMemory),
			},
		})
	}
}

// PublishLocalSource packages the specified source code in the --local-path flag and creates an image
// that will be eventually published to the registry specified in the --source-image flag.
// Returns a boolean that indicates if user does actually want to publish the image and an error in case of failure
func (opts *WorkloadOptions) PublishLocalSource(ctx context.Context, c *cli.Config, workload *cartov1alpha1.Workload) (bool, error) {
	if opts.LocalPath == "" {
		return true, nil
	}

	taggedImage := strings.Split(workload.Spec.Source.Image, "@sha")[0]
	okToPush := opts.checkToPublishLocalSource(taggedImage, c, workload)
	if !okToPush {
		return okToPush, nil
	}

	c.Infof("Publishing source in %q to %q...\n", opts.LocalPath, taggedImage)
	digestedImage, err := source.ImgpkgPush(ctx, opts.LocalPath, taggedImage)
	if err != nil {
		return okToPush, err
	}
	workload.Spec.Source.Image = digestedImage
	c.Successf("Published source\n")

	return okToPush, nil
}

func (opts *WorkloadOptions) checkToPublishLocalSource(taggedImage string, c *cli.Config, workload *cartov1alpha1.Workload) bool {
	okToPush := true
	if !opts.Yes {
		err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Publish source in %q to %q? It may be visible to others who can pull images from that repository", opts.LocalPath, taggedImage),
		}, &okToPush, printer.WithSurveyStdio(c.Stdin, c.Stdout, c.Stderr))
		if err != nil || !okToPush {
			c.Infof("Skipping workload %q\n", workload.Name)
			return false
		}
	}
	return okToPush
}

func (opts *WorkloadOptions) Update(ctx context.Context, c *cli.Config, currentWorkload *cartov1alpha1.Workload, workload *cartov1alpha1.Workload) (bool, error) {
	okToUpdate := false

	if msgs := workload.DeprecationWarnings(); len(msgs) != 0 {
		for _, msg := range msgs {
			c.Infof("WARNING: %s\n", msg)
		}
	}

	difference, noChange, err := printer.ResourceDiff(currentWorkload, workload, c.Scheme)
	if err != nil {
		return okToUpdate, err
	}

	if noChange {
		c.Infof("Workload is unchanged, skipping update\n")
		return okToUpdate, nil
	}
	c.Printf("Update workload:\n")
	c.Printf("%s\n", difference)

	if !opts.Yes {
		if opts.FilePath == "-" {
			c.Errorf("Skipping workload, cannot confirm intent. Run command with %s flag to confirm intent when providing input from stdin\n", flags.YesFlagName)
			return okToUpdate, nil
		} else {
			err := survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf("Really update the workload %q?", workload.Name),
			}, &okToUpdate, printer.WithSurveyStdio(c.Stdin, c.Stdout, c.Stderr))

			if err != nil || !okToUpdate {
				c.Infof("Skipping workload %q\n", workload.Name)
				return okToUpdate, nil
			}
		}
	}

	if err := c.Update(ctx, workload); err != nil {
		okToUpdate = false
		if apierrs.IsConflict(err) {
			c.Printf("%s conflict updating workload, the object was modified by another user; please run the update command again\n", printer.Serrorf("Error:"))
			return okToUpdate, cli.SilenceError(err)
		}
		return okToUpdate, err
	}

	c.Successf("Updated workload %q\n", workload.Name)
	return okToUpdate, nil
}

func (opts *WorkloadOptions) Create(ctx context.Context, c *cli.Config, workload *cartov1alpha1.Workload) (bool, error) {
	okToCreate := false

	if msgs := workload.DeprecationWarnings(); len(msgs) != 0 {
		for _, msg := range msgs {
			c.Infof("WARNING: %s\n", msg)
		}
	}

	diff, _, err := printer.ResourceDiff(nil, workload, c.Scheme)
	if err != nil {
		return okToCreate, err
	}

	c.Printf("Create workload:\n")
	c.Printf("%s\n", diff)

	if !opts.Yes {
		if opts.FilePath == "-" {
			c.Errorf("Skipping workload, cannot confirm intent. Run command with %s flag to confirm intent when providing input from stdin\n", flags.YesFlagName)
			return okToCreate, nil
		} else {
			err := survey.AskOne(&survey.Confirm{
				Message: "Do you want to create this workload?",
			}, &okToCreate, printer.WithSurveyStdio(c.Stdin, c.Stdout, c.Stderr))

			if err != nil || !okToCreate {
				c.Infof("Skipping workload %q\n", workload.Name)
				return okToCreate, nil
			}
		}
	}

	if err := c.Create(ctx, workload); err != nil {
		return okToCreate, err
	}
	c.Successf("Created workload %q\n", workload.Name)

	return okToCreate, nil
}

func (opts *WorkloadOptions) LoadInputWorkload(input io.Reader, workload *cartov1alpha1.Workload) error {
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

func (opts *WorkloadOptions) DefineFlags(ctx context.Context, c *cli.Config, cmd *cobra.Command) {
	cli.NamespaceFlag(ctx, cmd, c, &opts.Namespace)
	cmd.Flags().StringVarP(&opts.FilePath, cli.StripDash(flags.FilePathFlagName), "f", "", "`file path` containing the description of a single workload, other flags are layered on top of this resource. Use value \"-\" to read from stdin")
	cmd.Flags().StringVar(&opts.App, cli.StripDash(flags.AppFlagName), "", "application `name` the workload is a part of")
	cmd.Flags().StringVar(&opts.Type, cli.StripDash(flags.TypeFlagName), "", "distinguish workload `type`")
	cmd.RegisterFlagCompletionFunc(cli.StripDash(flags.TypeFlagName), func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"web"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringSliceVar(&opts.Labels, cli.StripDash(flags.LabelFlagName), []string{}, "label is represented as a `\"key=value\" pair`, or \"key-\" to remove. This flag may be specified multiple times")
	cmd.Flags().StringSliceVar(&opts.Annotations, cli.StripDash(flags.AnnotationFlagName), []string{}, "annotation is represented as a `\"key=value\" pair`, or \"key-\" to remove. This flag may be specified multiple times")
	cmd.Flags().StringArrayVar(&opts.Params, cli.StripDash(flags.ParamFlagName), []string{}, "additional parameters represented as a `\"key=value\" pair`, or \"key-\" to remove. This flag may be specified multiple times")
	cmd.Flags().BoolVar(&opts.Debug, cli.StripDash(flags.DebugFlagName), false, "put the workload in debug mode, "+flags.DebugFlagName+"=false to disable")
	cmd.Flags().BoolVar(&opts.LiveUpdate, cli.StripDash(flags.LiveUpdateFlagName), false, "put the workload in live update mode, "+flags.LiveUpdateFlagName+"=false to disable")
	cmd.Flags().StringVar(&opts.GitRepo, cli.StripDash(flags.GitRepoFlagName), "", "git `url` to remote source code")
	cmd.Flags().StringVar(&opts.GitBranch, cli.StripDash(flags.GitBranchFlagName), "", "`branch` within the git repo to checkout")
	cmd.Flags().StringVar(&opts.GitCommit, cli.StripDash(flags.GitCommitFlagName), "", "commit `SHA` within the git repo to checkout")
	cmd.Flags().StringVar(&opts.GitTag, cli.StripDash(flags.GitTagFlagName), "", "`tag` within the git repo to checkout")
	cmd.Flags().StringVar(&opts.SourceImage, cli.StripDash(flags.SourceImageFlagName), "", "destination `image` repository where source code is staged before being built")
	cmd.Flags().StringVar(&opts.SubPath, cli.StripDash(flags.SubPathFlagName), "", "relative `path` within source directory containing workload source code. To unset, pass empty string")
	cmd.Flags().StringVar(&opts.LocalPath, cli.StripDash(flags.LocalPathFlagName), "", "`path` on the local file system to a directory of source code to build for the workload")
	cmd.MarkFlagDirname(cli.StripDash(flags.LocalPathFlagName))
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(flags.ImageFlagName), "", "pre-built `image`, skips the source resolution and build phases of the supply chain")
	cmd.Flags().StringArrayVar(&opts.Env, cli.StripDash(flags.EnvFlagName), []string{}, "environment variables represented as a `\"key=value\" pair`, or \"key-\" to remove. This flag may be specified multiple times")
	cmd.Flags().StringArrayVar(&opts.BuildEnv, cli.StripDash(flags.BuildEnvFlagName), []string{}, "build environment variables represented as a `\"key=value\" pair`, or \"key-\" to remove. This flag may be specified multiple times")
	cmd.Flags().StringArrayVar(&opts.ServiceRefs, cli.StripDash(flags.ServiceRefFlagName), []string{}, "`object reference` for a service to bind to the workload \"database=rabbitmq.com/v1beta1:RabbitmqCluster:my-broker\", or \"database-\" to delete. This flag may be specified multiple times.")
	cmd.Flags().StringVar(&opts.LimitCPU, cli.StripDash(flags.LimitCPUFlagName), "", "the maximum amount of cpu allowed, in CPU `cores` (500m = .5 cores)")
	cmd.Flags().StringVar(&opts.LimitMemory, cli.StripDash(flags.LimitMemoryFlagName), "", "the maximum amount of memory allowed, in `bytes` (500Mi = 500MiB = 500 * 1024 * 1024)")
	cmd.Flags().StringVar(&opts.RequestCPU, cli.StripDash(flags.RequestCPUFlagName), "", "the minimum amount of cpu required, in CPU `cores` (500m = .5 cores)")
	cmd.Flags().StringVar(&opts.RequestMemory, cli.StripDash(flags.RequestMemoryFlagName), "", "the minimum amount of memory required, in `bytes` (500Mi = 500MiB = 500 * 1024 * 1024)")
	cmd.Flags().BoolVar(&opts.Wait, cli.StripDash(flags.WaitFlagName), false, "waits for workload to become ready")
	cmd.Flags().DurationVar(&opts.WaitTimeout, cli.StripDash(flags.WaitTimeoutFlagName), 10*time.Minute, "timeout for workload to become ready when waiting")
	cmd.RegisterFlagCompletionFunc(cli.StripDash(flags.WaitTimeoutFlagName), completion.SuggestDurationUnits(ctx, completion.CommonDurationUnits))
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(flags.TailFlagName), false, "show logs while waiting for workload to become ready")
	cmd.Flags().BoolVar(&opts.TailTimestamps, cli.StripDash(flags.TailTimestampFlagName), false, "show logs and add timestamp to each log line while waiting for workload to become ready")
	cmd.MarkFlagFilename(cli.StripDash(flags.FilePathFlagName), ".yaml", ".yml")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(flags.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")
	cmd.Flags().BoolVarP(&opts.Yes, cli.StripDash(flags.YesFlagName), "y", false, "accept all prompts")
}
