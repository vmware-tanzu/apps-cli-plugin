// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"

// CmdGroup is a group of CLI commands.
type CmdGroup string

// PluginCompletionType is the mechanism used for determining command line completion options.
type PluginCompletionType int

// Hook is the mechanism used to define function for plugin hooks
type Hook func() error

const (
	// NativePluginCompletion indicates command line completion is determined using the built in
	// cobra.Command __complete mechanism.
	NativePluginCompletion PluginCompletionType = iota
	// StaticPluginCompletion indicates command line completion will be done by using a statically
	// defined list of options.
	StaticPluginCompletion
	// DynamicPluginCompletion indicates command line completion will be retrieved from the plugin
	// at runtime.
	DynamicPluginCompletion

	// RunCmdGroup are commands associated with Tanzu Run.
	RunCmdGroup CmdGroup = "Run"

	// ManageCmdGroup are commands associated with Tanzu Manage.
	ManageCmdGroup CmdGroup = "Manage"

	// BuildCmdGroup are commands associated with Tanzu Build.
	BuildCmdGroup CmdGroup = "Build"

	// ObserveCmdGroup are commands associated with Tanzu Observe.
	ObserveCmdGroup CmdGroup = "Observe"

	// SystemCmdGroup are system commands.
	SystemCmdGroup CmdGroup = "System"

	// TargetCmdGroup are various target commands.
	TargetCmdGroup CmdGroup = "Target"

	// VersionCmdGroup are version commands.
	VersionCmdGroup CmdGroup = "Version"

	// AdminCmdGroup are admin commands.
	AdminCmdGroup CmdGroup = "Admin"

	// TestCmdGroup is the test command group.
	TestCmdGroup CmdGroup = "Test"

	// ExtraCmdGroup is the extra command group.
	ExtraCmdGroup CmdGroup = "Extra"
)

// PluginDescriptor describes a plugin binary.
type PluginDescriptor struct {
	// Name is the name of the plugin.
	Name string `json:"name" yaml:"name"`

	// Description is the plugin's description.
	Description string `json:"description" yaml:"description"`

	// Target is the target to which plugin is applicable.
	Target types.Target `json:"target" yaml:"target"`

	// Version of the plugin. Must be a valid semantic version https://semver.org/
	Version string `json:"version" yaml:"version"`

	// BuildSHA is the git commit hash the plugin was built with.
	BuildSHA string `json:"buildSHA" yaml:"buildSHA"`

	// Digest is the SHA256 hash of the plugin binary.
	Digest string `json:"digest" yaml:"digest"`

	// Command group for the plugin.
	Group CmdGroup `json:"group" yaml:"group"`

	// DocURL for the plugin.
	DocURL string `json:"docURL" yaml:"docURL"`

	// Hidden tells whether the plugin should be hidden from the help command.
	Hidden bool `json:"hidden,omitempty" yaml:"hidden,omitempty"`

	// CompletionType determines how command line completion will be determined.
	CompletionType PluginCompletionType `json:"completionType" yaml:"completionType"`

	// CompletionArgs contains the valid command line completion values if `CompletionType`
	// is set to `StaticPluginCompletion`.
	CompletionArgs []string `json:"completionArgs,omitempty" yaml:"completionArgs,omitempty"`

	// CompletionCommand is the command to call from the plugin to retrieve a list of
	// valid completion nouns when `CompletionType` is set to `DynamicPluginCompletion`.
	CompletionCommand string `json:"completionCmd,omitempty" yaml:"completionCmd,omitempty"`

	// Aliases are other text strings used to call this command
	Aliases []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`

	// PostInstallHook is function to be run post install of a plugin.
	PostInstallHook Hook `json:"-" yaml:"-"`

	// DefaultFeatureFlags is default featureflags to be configured if missing when invoking plugin
	DefaultFeatureFlags map[string]bool `json:"defaultFeatureFlags,omitempty" yaml:"defaultFeatureFlags,omitempty"`
}
