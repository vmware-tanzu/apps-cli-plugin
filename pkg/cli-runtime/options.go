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

package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

// ValidateE bridges a cobra RunE function to the Validatable interface.  All flags and
// arguments must already be bound, with explicit or default values, to the struct being
// validated. This function is typically used to define the PreRunE phase of a command.
//
// ```
//
//	cmd := &cobra.Command{
//		   ...
//		   PreRunE: cli.ValidateE(obj),
//	}
//
// ```
func ValidateE(ctx context.Context, obj validation.Validatable) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := WithCommand(ctx, cmd)
		if err := obj.Validate(ctx); len(err) != 0 {
			return err.ToAggregate()
		}
		cmd.SilenceUsage = true
		return nil
	}
}

type Executable interface {
	Exec(ctx context.Context, c *Config) error
}

// ExecE bridges a cobra RunE function to the Executable interface.  All flags and
// arguments must already be bound, with explicit or default values, to the struct being
// executed. This function is typically used to define the RunE phase of a command.
//
// ```
//
//	cmd := &cobra.Command{
//		   ...
//		   RunE: cli.ExecE(c, obj),
//	}
//
// ```
func ExecE(ctx context.Context, c *Config, obj Executable) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := WithCommand(ctx, cmd)
		if o, ok := obj.(DryRunable); ok && o.IsDryRun() {
			// reserve Stdout for resources, redirect normal stdout to stderr
			ctx = WithStdout(ctx, c.Stdout)
			c.Stdout = c.Stderr
		}
		return obj.Exec(ctx, c)
	}
}
