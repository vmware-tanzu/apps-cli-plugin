/*
Copyright 2019 VMware, Inc.

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

package testing_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

type TestResourceCreateOptions struct {
	Namespace string
	Name      string
}

var (
	_ validation.Validatable = (*TestResourceCreateOptions)(nil)
	_ cli.Executable         = (*TestResourceCreateOptions)(nil)
)

func (opts *TestResourceCreateOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NamespaceFlagName))
	}

	if opts.Name == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NameArgumentName))
	} else {
		errs = errs.Also(validation.K8sName(opts.Name, cli.NameArgumentName))
	}

	return errs
}

func (opts *TestResourceCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	resource := &rtesting.TestResource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
	}
	err := c.Create(ctx, resource)
	if err != nil {
		return err
	}
	c.Successf("Created resource %q\n", resource.Name)
	return nil
}

func NewTestResourceCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &TestResourceCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a resource",
		Example: strings.Join([]string{
			fmt.Sprintf("%s resource create my-resource", c.Name),
		}, "\n"),
		PreRunE: cli.ValidateE(ctx, opts),
		RunE:    cli.ExecE(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(ctx, cmd, c, &opts.Namespace)

	return cmd
}

func TestValidatableTestSuite(t *testing.T) {
	table := clitesting.ValidatableTestSuite{
		{
			Name:        "invalid resource",
			Validatable: &TestResourceCreateOptions{},
			ExpectFieldErrors: validation.FieldErrors{}.Also(
				validation.ErrMissingField(cli.NamespaceFlagName),
				validation.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "valid",
			Validatable: &TestResourceCreateOptions{
				Namespace: "default",
				Name:      "my-resource",
			},
			ShouldValidate: true,
		},
		{
			Name:           "skipped test, but would fail if run",
			Skip:           true,
			Validatable:    &TestResourceCreateOptions{},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestWorkloadCreateCommand(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = rtesting.AddToScheme(scheme)

	table := clitesting.CommandTestSuite{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "create resource",
			Args: []string{"my-resource"},
			ExpectCreates: []clitesting.Factory{
				clitesting.Wrapper(&rtesting.TestResource{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-resource",
					},
					Spec: rtesting.TestResourceSpec{},
				}),
			},
			ExpectOutput: `
Created resource "my-resource"
`,
		},
		{
			Name: "create failed",
			Args: []string{"my-resource"},
			GivenObjects: []rtesting.Factory{
				clitesting.Wrapper(&rtesting.TestResource{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-resource",
					},
					Spec: rtesting.TestResourceSpec{},
				}),
			},
			ExpectCreates: []clitesting.Factory{
				clitesting.Wrapper(&rtesting.TestResource{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-resource",
					},
					Spec: rtesting.TestResourceSpec{},
				}),
			},
			ShouldError: true,
		},
	}

	table.Run(t, scheme, NewTestResourceCreateCommand)
}
