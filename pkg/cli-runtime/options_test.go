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

package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

type StubValidate struct {
	called        bool
	validationErr validation.FieldErrors
}

func (o *StubValidate) Validate(ctx context.Context) validation.FieldErrors {
	o.called = true
	return o.validationErr
}

func TestValidateE(t *testing.T) {
	tests := []struct {
		name          string
		opts          *StubValidate
		expectedErr   error
		usageSilenced bool
	}{{
		name:          "valid, no error",
		opts:          &StubValidate{},
		usageSilenced: true,
	}, {
		name: "valid, empty error",
		opts: &StubValidate{
			validationErr: validation.FieldErrors{},
		},
		usageSilenced: true,
	}, {
		name: "validation error",
		opts: &StubValidate{
			validationErr: validation.ErrMissingField("field-name"),
		},
		expectedErr:   validation.ErrMissingField("field-name").ToAggregate(),
		usageSilenced: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := &cobra.Command{}
			err := cli.ValidateE(ctx, test.opts)(cmd, []string{})

			if expected, actual := true, test.opts.called; true != actual {
				t.Errorf("expected called to be %v, actually %v", expected, actual)
			}
			if expected, actual := test.expectedErr, err; fmt.Sprintf("%s", expected) != fmt.Sprintf("%s", actual) {
				t.Errorf("expected error to be %v, actually %v", expected, actual)
			}
			if expected, actual := test.usageSilenced, cmd.SilenceUsage; expected != actual {
				t.Errorf("expected cmd.SilenceUsage to be %v, actually %v", expected, actual)
			}
		})
	}
}

type StubExec struct {
	dryRun  bool
	called  bool
	config  *cli.Config
	cmd     *cobra.Command
	execErr error
}

var (
	_ cli.Executable = (*StubExec)(nil)
	_ cli.DryRunable = (*StubExec)(nil)
)

func (o *StubExec) Exec(ctx context.Context, c *cli.Config) error {
	o.called = true
	o.config = c
	o.cmd = cli.CommandFromContext(ctx)
	return o.execErr
}

func (o *StubExec) IsDryRun() bool {
	return o.dryRun
}

func TestExecE(t *testing.T) {
	tests := []struct {
		name        string
		opts        *StubExec
		expectedErr error
	}{{
		name: "success",
		opts: &StubExec{},
	}, {
		name: "failure",
		opts: &StubExec{
			execErr: fmt.Errorf("test exec error"),
		},
		expectedErr: fmt.Errorf("test exec error"),
	}, {
		name: "dry run",
		opts: &StubExec{
			dryRun: true,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := &cobra.Command{}
			config := &cli.Config{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
			}
			err := cli.ExecE(ctx, config, test.opts)(cmd, []string{})

			if expected, actual := true, test.opts.called; true != actual {
				t.Errorf("expected called to be %v, actually %v", expected, actual)
			}
			if expected, actual := test.expectedErr, err; fmt.Sprintf("%s", expected) != fmt.Sprintf("%s", actual) {
				t.Errorf("expected error to be %v, actually %v", expected, actual)
			}
			if expected, actual := config, test.opts.config; expected != actual {
				t.Errorf("expected config to be %v, actually %v", expected, actual)
			}
			if expected, actual := cmd, test.opts.cmd; expected != actual {
				t.Errorf("expected command to be %v, actually %v", expected, actual)
			}
			if test.opts.dryRun {
				if config.Stdout != config.Stderr {
					t.Errorf("expected stdout and stderr to be the same, actually %v %v", config.Stdout, config.Stderr)
				}
			} else {
				if config.Stdout == config.Stderr {
					t.Errorf("expected stdout and stderr to be different, actually %v %v", config.Stdout, config.Stderr)
				}
			}
		})
	}
}
