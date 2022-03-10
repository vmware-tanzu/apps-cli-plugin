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

package flags

import (
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

const (
	AllFlagName           = "--all"
	AllNamespacesFlagName = cli.AllNamespacesFlagName
	AnnotationFlagName    = "--annotation"
	AppFlagName           = "--app"
	BuildEnvFlagName      = "--build-env"
	ComponentFlagName     = "--component"
	ConfigFlagName        = "--config"
	ContextFlagName       = cli.ContextFlagName
	DebugFlagName         = "--debug"
	DryRunFlagName        = "--dry-run"
	EnvFlagName           = "--env"
	ExportFlagName        = "--export"
	FilePathFlagName      = "--file"
	GitBranchFlagName     = "--git-branch"
	GitCommitFlagName     = "--git-commit"
	GitFlagWildcard       = "--git-*"
	GitRepoFlagName       = "--git-repo"
	GitTagFlagName        = "--git-tag"
	ImageFlagName         = "--image"
	KubeConfigFlagName    = cli.KubeConfigFlagName
	LabelFlagName         = "--label"
	LimitCPUFlagName      = "--limit-cpu"
	LimitMemoryFlagName   = "--limit-memory"
	LiveUpdateFlagName    = "--live-update"
	LocalPathFlagName     = "--local-path"
	NamespaceFlagName     = cli.NamespaceFlagName
	NoColorFlagName       = cli.NoColorFlagName
	OutputFlagName        = "--output"
	ParamFlagName         = "--param"
	RequestCPUFlagName    = "--request-cpu"
	RequestMemoryFlagName = "--request-memory"
	ServiceRefFlagName    = "--service-ref"
	SinceFlagName         = "--since"
	SourceImageFlagName   = "--source-image"
	TailFlagName          = "--tail"
	TimestampFlagName     = "--timestamp"
	TailTimestampFlagName = "--tail-timestamp"
	TypeFlagName          = "--type"
	VerboseLevelFlagName  = "--verbose"
	WaitFlagName          = "--wait"
	WaitTimeoutFlagName   = "--wait-timeout"
	YesFlagName           = "--yes"
)
