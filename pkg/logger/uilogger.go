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

package logger

import (
	"context"

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/plainimage"
)

// Logger interface that defines the logger functions
type loggerkey struct{}

var _ plainimage.Logger = &NoopLogger{}

// NewNoopLogger creates a new noop logger
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// NoopLogger this logger will not print
type NoopLogger struct{}

// Logf does nothing
func (n NoopLogger) Logf(string, ...interface{}) {}

func RetrieveStashLogger(ctx context.Context) plainimage.Logger {
	val, ok := ctx.Value(loggerkey{}).(plainimage.Logger)
	if ok {
		return val
	}
	return NewNoopLogger()
}
