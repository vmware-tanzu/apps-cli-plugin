//go:build integration
// +build integration

/*
Copyright 2022 VMware, Inc.

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

package suite_test

import (
	"path/filepath"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

var (
	ConsoleOutBasePath = filepath.Join("testdata", "console-out")
	ExpectedBasePath   = filepath.Join("testdata", "expected")
	SrcBasePath        = filepath.Join("testdata", "src")
	namespaceFlag      = "--namespace=" + TestingNamespace
	ApplyEmojis        = []cli.Icon{cli.Magnifying, cli.ThumbsUp}
	UpdateEmojis       = []cli.Icon{cli.Exclamation, cli.Magnifying, cli.ThumbsUp, cli.Question}
	GetEmojis          = []cli.Icon{cli.Antenna, cli.Delivery, cli.SpeechBalloon}
	DeleteEmojis       = []cli.Icon{cli.ThumbsUp}
)
