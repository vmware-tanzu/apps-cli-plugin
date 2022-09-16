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

package logger

import (
	"context"
	"testing"

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/plainimage"
)

func TestRetrieveSourceImageLogger(t *testing.T) {
	actualLogger := NewNoopLogger()
	ctx := StashSourceImageLogger(context.Background(), actualLogger)
	expectedLogger := RetrieveSourceImageLogger(ctx)
	if expectedLogger != actualLogger {
		t.Errorf("RetrieveSourceImageLogger() failed. wanted %v, got %v", actualLogger, expectedLogger)
	}
}

func TestStashSourceImageLogger(t *testing.T) {
	actualLogger := NewNoopLogger()
	ctx := StashSourceImageLogger(context.Background(), actualLogger)
	if ctx.Value(sourceImageloggerkey{}).(plainimage.Logger) != actualLogger {
		t.Errorf("StashSourceImageLogger() failed. wanted %v, got %v", actualLogger, ctx.Value(sourceImageloggerkey{}).(plainimage.Logger))
	}
}
