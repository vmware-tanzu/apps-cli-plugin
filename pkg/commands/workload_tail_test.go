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

package commands_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	diemetav1 "dies.dev/apis/meta/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/logs"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	diecartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/dies/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

func TestWorkloadTailOptionsValidate(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:        "invalid empty",
			Validatable: &commands.WorkloadTailOptions{},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(flags.NamespaceFlagName),
				validation.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "valid",
			Validatable: &commands.WorkloadTailOptions{
				Namespace: "default",
				Name:      "my-workload",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Validatable: &commands.WorkloadTailOptions{
				Namespace: "default",
				Name:      "my-",
			},
			ExpectFieldErrors: validation.ErrInvalidValue("my-", cli.NameArgumentName),
		},
		{
			Name: "since",
			Validatable: &commands.WorkloadTailOptions{
				Namespace: "default",
				Name:      "my-workload",
				Since:     time.Minute,
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid since",
			Validatable: &commands.WorkloadTailOptions{
				Namespace: "default",
				Name:      "my-workload",
				Since:     -1,
			},
			ExpectFieldErrors: validation.ErrInvalidValue(-1*time.Nanosecond, flags.SinceFlagName),
		},
		{
			Name: "component",
			Validatable: &commands.WorkloadTailOptions{
				Namespace: "default",
				Name:      "my-workload",
				Component: "build",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid component",
			Validatable: &commands.WorkloadTailOptions{
				Namespace: "default",
				Name:      "my-workload",
				Component: "---",
			},
			ExpectFieldErrors: validation.ErrInvalidValue("---", flags.ComponentFlagName),
		},
	}
	table.Run(t)
}

func TestWorkloadTailCommand(t *testing.T) {
	workloadName := "test-workload"
	defaultNamespace := "default"

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	parent := diecartov1alpha1.WorkloadBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name(workloadName)
			d.Namespace(defaultNamespace)
		})

	table := clitesting.CommandTestSuite{
		{
			Name:        "empty",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:        "invalid namespace",
			Args:        []string{flags.NamespaceFlagName, "other-namespce", workloadName},
			ShouldError: true,
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
Workload "other-namespce/test-workload" not found
`,
		},
		{
			Name:        "failed to get workload",
			Args:        []string{flags.NamespaceFlagName, defaultNamespace, workloadName},
			ShouldError: true,
			WithReactors: []clitesting.ReactionFunc{
				clitesting.InduceFailure("get", "Workload"),
			},
		},

		{
			Name: "missing workload",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, workloadName},
			ExpectOutput: `
Workload "default/test-workload" not found
`,
			ShouldError: true,
		},
		{
			Name: "show logs for workload",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, flags.SinceFlagName, "1h", workloadName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Hour, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)
				// simulate a user exit after 10ms
				ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
				_ = cancel
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
...tail output...
`,
		},
		{
			Name: "show logs for workload since time being default",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, workloadName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Minute, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)
				// simulate a user exit after 10ms
				ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
				_ = cancel
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
...tail output...
`,
		},
		{
			Name: "show logs for workload with since time in seconds",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, flags.SinceFlagName, "1s", workloadName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Second, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)
				// simulate a user exit after 10ms
				ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
				_ = cancel
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
...tail output...
`,
		},
		{
			Name: "show logs for workload with component",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, flags.SinceFlagName, "1h", workloadName, flags.ComponentFlagName, "build"},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s,%s=%s", cartov1alpha1.WorkloadLabelName, workloadName, apis.ComponentLabelName, "build"))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Hour, false).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)
				// simulate a user exit after 10ms
				ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
				_ = cancel
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
...tail output...
`,
		},
		{
			Name: "error tailing logs",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, flags.SinceFlagName, "1h", workloadName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Hour, false).Return(fmt.Errorf("tail error")).Once()
				ctx = logs.StashTailer(ctx, tailer)
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent,
			},
			ShouldError: true,
			ExpectOutput: `
...tail output...
`,
		},
		{
			Name: "show logs for workload with timestamp flag",
			Args: []string{flags.NamespaceFlagName, defaultNamespace, flags.SinceFlagName, "1h", flags.TimestampFlagName, workloadName},
			Prepare: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) (context.Context, error) {
				tailer := &logs.FakeTailer{}
				selector, _ := labels.Parse(fmt.Sprintf("%s=%s", cartov1alpha1.WorkloadLabelName, workloadName))
				tailer.On("Tail", mock.Anything, "default", selector, []string{}, time.Hour, true).Return(nil).Once()
				ctx = logs.StashTailer(ctx, tailer)
				// simulate a user exit after 10ms
				ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
				_ = cancel
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, config *cli.Config, tc *clitesting.CommandTestCase) error {
				tailer := logs.RetrieveTailer(ctx).(*logs.FakeTailer)
				tailer.AssertExpectations(t)
				return nil
			},
			GivenObjects: []client.Object{
				parent,
			},
			ExpectOutput: `
...tail output...
`,
		},
	}
	table.Run(t, scheme, func(ctx context.Context, c *cli.Config) *cobra.Command {
		return commands.NewWorkloadTailCommand(ctx, c)
	})
}
