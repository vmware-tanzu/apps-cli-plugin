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
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/logs"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

type WorkloadTailOptions struct {
	Namespace string
	Name      string

	Component  string
	Since      time.Duration
	Timestamps bool
}

var (
	_ validation.Validatable = (*WorkloadTailOptions)(nil)
	_ cli.Executable         = (*WorkloadTailOptions)(nil)
)

func (opts *WorkloadTailOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(validation.ErrMissingField(flags.NamespaceFlagName))
	}

	if opts.Name == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NameArgumentName))
	} else {
		errs = errs.Also(validation.K8sName(opts.Name, cli.NameArgumentName))
	}

	if opts.Since < 0 {
		errs = errs.Also(validation.ErrInvalidValue(opts.Since, flags.SinceFlagName))
	}

	errs = errs.Also(validation.K8sLabelValue(opts.Component, flags.ComponentFlagName))
	return errs
}

func (opts *WorkloadTailOptions) Exec(ctx context.Context, c *cli.Config) error {
	workload := &cartov1alpha1.Workload{}
	err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: opts.Name}, workload)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}
		c.Errorf("Workload %q not found\n", fmt.Sprintf("%s/%s", opts.Namespace, opts.Name))
		return cli.SilenceError(err)
	}

	labelSelector := fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workload.Name)
	if opts.Component != "" {
		labelSelector = fmt.Sprintf("%s=%s,%s=%s", cartov1alpha1.WorkloadLabelName, workload.Name, apis.ComponentLabelName, opts.Component)
	}
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return logs.Tail(ctx, c, opts.Namespace, selector, containers, opts.Since, opts.Timestamps)
}

func NewWorkloadTailCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadTailOptions{}

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Watch workload related logs",
		Long: strings.TrimSpace(`
Stream logs for a workload until canceled. To cancel, press Ctl-c in
the shell or stop the process. As new workload pods are started, the logs
are displayed. To show historical logs use ` + flags.SinceFlagName + `.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload tail my-workload", c.Name),
			fmt.Sprintf("%s workload tail my-workload %s 1h", c.Name, flags.SinceFlagName),
		}, "\n"),
		PreRunE:           cli.ValidateE(ctx, opts),
		RunE:              cli.ExecE(ctx, c, opts),
		ValidArgsFunction: completion.SuggestWorkloadNames(ctx, c),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(ctx, cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Component, cli.StripDash(flags.ComponentFlagName), "", "workload component `name` (e.g. build)")
	cmd.RegisterFlagCompletionFunc(cli.StripDash(flags.ComponentFlagName), completion.SuggestComponentNames(ctx, c))
	cmd.Flags().BoolVarP(&opts.Timestamps, cli.StripDash(flags.TimestampFlagName), "t", false, "print timestamp for each log line")
	cmd.Flags().DurationVar(&opts.Since, cli.StripDash(flags.SinceFlagName), time.Minute, "time `duration` to start reading logs from")
	cmd.RegisterFlagCompletionFunc(cli.StripDash(flags.SinceFlagName), completion.SuggestDurationUnits(ctx, completion.CommonDurationUnits))
	return cmd
}
