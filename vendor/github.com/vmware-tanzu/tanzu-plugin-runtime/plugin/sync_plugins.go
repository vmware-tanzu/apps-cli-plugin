// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"errors"
	"fmt"
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

func runCommand(commandPath string, args []string) (bytes.Buffer, bytes.Buffer, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	command := exec.Command(commandPath, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout, stderr, err
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
// The output of the plugin syncing will be return as a string.
func SyncPluginsForTarget(target types.Target) (string, error) {
	// For now, the implementation expects env var TANZU_BIN to be set and
	// pointing to the core CLI binary used to invoke the plugin sync with.

	cliPath := os.Getenv("TANZU_BIN")
	if cliPath == "" {
		return "", errors.New("the environment variable TANZU_BIN is not set")
	}

	altCommandArgs := []string{customCommandName}
	args := []string{"plugin", "sync"}

	altCommandArgs = append(altCommandArgs, args...)
	altCommandArgs = append(altCommandArgs, "--target", string(target))

	// Check if there is an alternate means to perform the plugin syncing
	// operation, if not fall back to `plugin sync`
	output, _, err := runCommand(cliPath, altCommandArgs)
	if err == nil && output.String() != "" {
		args = strings.Fields(output.String())
	}

	// Runs the actual command
	stdoutOutput, stderrOutput, err := runCommand(cliPath, args)
	return fmt.Sprintf("%s%s", stdoutOutput.String(), stderrOutput.String()), err
}
