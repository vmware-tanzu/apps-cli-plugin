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
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
)

var kubeConfPath = filepath.Join("testdata", ".kube", "config")

func TestInitialize(t *testing.T) {
	scheme := runtime.NewScheme()
	c := Initialize("cli name", scheme)

	if c.Name != "cli name" {
		t.Errorf("expected Name to be %q", "cli name")
	}
	if c.Stdin != os.Stdin {
		t.Errorf("expected Stdin to be os.Stdin")
	}
	if c.Stdout != os.Stdout {
		t.Errorf("expected Stdout to be os.Stdout")
	}
	if c.Stderr != os.Stderr {
		t.Errorf("expected Stderr to be os.Stderr")
	}
	if c.Scheme != scheme {
		t.Errorf("expected Scheme to be scheme")
	}
	if reflect.ValueOf(c.Exec).Pointer() != reflect.ValueOf(exec.CommandContext).Pointer() {
		t.Errorf("expected Exec to be exec.CommandContext")
	}
}

func TestInitKubeConfig_Flag(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	scheme := runtime.NewScheme()
	c := NewDefaultConfig("cli name", scheme)
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	c.KubeConfigFile = kubeConfPath

	if expected, actual := "cli name", c.Name; expected != actual {
		t.Errorf("Expected name %q, actually %q", expected, actual)
	}
	if expected, actual := kubeConfPath, c.KubeConfigFile; expected != actual {
		t.Errorf("Expected kubeconfig path %q, actually %q", expected, actual)
	}
	if diff := cmp.Diff("", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
}

func TestInit(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	scheme := runtime.NewScheme()
	c := NewDefaultConfig("cli name", scheme)
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	c.KubeConfigFile = kubeConfPath
	c.init()

	if expected, actual := "my-namespace", c.DefaultNamespace(); expected != actual {
		t.Errorf("Expected default namespace %q, actually %q", expected, actual)
	}
	if diff := cmp.Diff("", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
	if c.Client == nil {
		t.Errorf("Expected c.Client tp be set, actually %v", c.Client)
	}
}
