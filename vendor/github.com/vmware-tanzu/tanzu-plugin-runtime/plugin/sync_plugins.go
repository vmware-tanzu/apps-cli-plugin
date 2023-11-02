// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
)

const (
	// customCommandName is the name of the command expected to be implemented
	// by the CLI should there be a need to discover and alternative invocation
	// method
	customCommandName string = "_custom_command"
)

// cmdOptions specifies the command options
type cmdOptions struct {
	outWriter io.Writer
	errWriter io.Writer
}

type CommandOptions func(o *cmdOptions)

// WithOutputWriter specifies the CommandOption for configuring Stdout
func WithOutputWriter(outWriter io.Writer) CommandOptions {
	return func(o *cmdOptions) {
		o.outWriter = outWriter
	}
}

// WithErrorWriter specifies the CommandOption for configuring Stderr
func WithErrorWriter(errWriter io.Writer) CommandOptions {
	return func(o *cmdOptions) {
		o.errWriter = errWriter
	}
}

// WithNoStdout specifies to ignore stdout
func WithNoStdout() CommandOptions {
	return func(o *cmdOptions) {
		o.outWriter = io.Discard
	}
}

// WithNoStderr specifies to ignore stderr
func WithNoStderr() CommandOptions {
	return func(o *cmdOptions) {
		o.errWriter = io.Discard
	}
}

func runCommand(commandPath string, args []string, opts *cmdOptions) (bytes.Buffer, bytes.Buffer, error) {
	command := exec.Command(commandPath, args...)

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	wout := io.MultiWriter(&stdout, os.Stdout)
	werr := io.MultiWriter(&stderr, os.Stderr)

	if opts.outWriter != nil {
		wout = io.MultiWriter(&stdout, opts.outWriter)
	}
	if opts.errWriter != nil {
		werr = io.MultiWriter(&stderr, opts.errWriter)
	}

	command.Stdout = wout
	command.Stderr = werr

	return stdout, stderr, command.Run()
}

// SyncPluginsForTarget will attempt to install plugins required by the active
// Context of the provided target. This is most useful for any plugin
// implementation which creates a new Context or updates an existing one as
// part of its operation, and prefers that the plugins appropriate for the
// Context are immediately available for use.
//
// Note: This API is considered EXPERIMENTAL. Both the function's signature and
// implementation are subjected to change/removal if an alternative means to
// provide equivalent functionality can be introduced.
//
// By default this API will write to os.Stdout and os.Stderr.
// To write the logs to different output and error streams as part of the plugin sync
// command invocation, configure CommandOptions as part of the parameters.
//
// Example:
//
//	var outBuf bytes.Buffer
//	var errBuf bytes.Buffer
//	SyncPluginsForTarget(types.TargetK8s, WithOutputWriter(&outBuf), WithErrorWriter(&errBuf))
//
// Deprecated: SyncPluginsForTarget is deprecated. Use SyncPluginsForContextType instead
func SyncPluginsForTarget(target types.Target, opts ...CommandOptions) (string, error) {
	return SyncPluginsForContextType(types.ConvertTargetToContextType(target), opts...)
}

// SyncPluginsForContextType will attempt to install plugins required by the active
// Context of the provided contextType. This is most useful for any plugin
// implementation which creates a new Context or updates an existing one as
// part of its operation, and prefers that the plugins appropriate for the
// Context are immediately available for use.
//
// Note: This API is considered EXPERIMENTAL. Both the function's signature and
// implementation are subjected to change/removal if an alternative means to
// provide equivalent functionality can be introduced.
//
// By default this API will write to os.Stdout and os.Stderr.
// To write the logs to different output and error streams as part of the plugin sync
// command invocation, configure CommandOptions as part of the parameters.
//
// Example:
//
//	var outBuf bytes.Buffer
//	var errBuf bytes.Buffer
//	SyncPluginsForContextType(types.ContextTypeK8s, WithOutputWriter(&outBuf), WithErrorWriter(&errBuf))
func SyncPluginsForContextType(contextType types.ContextType, opts ...CommandOptions) (string, error) {
	// For now, the implementation expects env var TANZU_BIN to be set and
	// pointing to the core CLI binary used to invoke the plugin sync with.

	options := &cmdOptions{}
	for _, opt := range opts {
		opt(options)
	}

	cliPath := os.Getenv("TANZU_BIN")
	if cliPath == "" {
		return "", errors.New("the environment variable TANZU_BIN is not set")
	}

	altCommandArgs := []string{customCommandName}
	args := []string{"plugin", "sync"}

	altCommandArgs = append(altCommandArgs, args...)
	altCommandArgs = append(altCommandArgs, "--type", string(contextType))

	// Check if there is an alternate means to perform the plugin syncing
	// operation, if not fall back to `plugin sync`
	stdoutOutput, _, err := runCommand(cliPath, altCommandArgs, &cmdOptions{outWriter: io.Discard, errWriter: io.Discard})
	if err == nil && stdoutOutput.String() != "" {
		args = strings.Fields(stdoutOutput.String())
	}

	// Runs the actual command
	stdoutOutput, stderrOutput, err := runCommand(cliPath, args, options)
	return fmt.Sprintf("%s%s", stdoutOutput.String(), stderrOutput.String()), err
}
