// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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
