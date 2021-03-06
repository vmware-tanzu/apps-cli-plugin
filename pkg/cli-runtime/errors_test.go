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
	"errors"
	"fmt"
	"testing"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func TestSilenceError(t *testing.T) {
	err := fmt.Errorf("test error")
	silentErr := cli.SilenceError(err)

	if errors.Is(err, cli.SilentError) {
		t.Errorf("expected error to not be silent, got %#v", err)
	}
	if !errors.Is(silentErr, cli.SilentError) {
		t.Errorf("expected error to be silent, got %#v", err)
	}
	if expected, actual := err, errors.Unwrap(silentErr); expected != actual {
		t.Errorf("errors expected to match, expected %v, actually %v", expected, actual)
	}
	if expected, actual := err.Error(), silentErr.Error(); expected != actual {
		t.Errorf("errors expected to match, expected %q, actually %q", expected, actual)
	}
}
