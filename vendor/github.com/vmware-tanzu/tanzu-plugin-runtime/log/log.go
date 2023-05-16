// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package log provides logging functionalities
package log

import (
	"fmt"
	"io"
	"os"
)

// LoggerImpl represents the ability to log messages, both errors and not.
type LoggerImpl interface {
	// Enabled tests whether this Logger is enabled.  For example, commandline
	// flags might be used to set the logging verbosity and disable some info
	// logs.
	Enabled() bool

	// Info logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Info(msg string, keysAndValues ...interface{})

	// Error logs an error, with the given message and key/value pairs as context.
	// It functions similarly to calling Info with the "error" named value, but may
	// have unique behavior, and should be preferred for logging errors (see the
	// package documentations for more information).
	//
	// The msg field should be used to add context to any underlying error,
	// while the err field should be used to attach the actual error that
	// triggered this log line, if present.
	Error(err error, msg string, keysAndValues ...interface{})

	// V returns an Logger value for a specific verbosity level, relative to
	// this Logger.  In other words, V values are additive.  V higher verbosity
	// level means a log message is less important.  It's illegal to pass a log
	// level less than zero.
	V(level int) LoggerImpl

	// WithValues adds some key-value pairs of context to a logger.
	// See Info for documentation on how key/value pairs work.
	WithValues(keysAndValues ...interface{}) LoggerImpl

	// WithName adds a new element to the logger's name.
	// Successive calls with WithName continue to append
	// suffixes to the logger's name.  It's strongly recommended
	// that name segments contain only letters, digits, and hyphens
	// (see the package documentation for more information).
	WithName(name string) LoggerImpl

	// Infof logs a non-error messages with the given message format with format specifier and arguments.
	Infof(format string, args ...interface{})
	// Errorf logs a error message with the given key/value pairs as context.
	Errorf(format string, args ...interface{})
	// Warning logs a warning messages with the given key/value pairs as context.
	Warning(msg string, kvs ...interface{})
	// Warningf logs a warning messages with the given message format with format specifier and arguments.
	Warningf(format string, args ...interface{})
	// Success logs a success messages with the given key/value pairs as context.
	Success(msg string, kvs ...interface{})
	// Successf logs a success messages with the given message format with format specifier and arguments.
	Successf(format string, args ...interface{})
	// Fatal logs a fatal message with the given key/value pairs as context and returns with os.Exit(1)
	Fatal(err error, msg string, kvs ...interface{})
	// Print logs a message of generic type
	Print(msg string, err error, logType string, kvs ...interface{})
	// Output writes a message to stdout
	Outputf(msg string, kvs ...interface{})
	// SetThreshold implements a New Option that allows to set the threshold level for a logger.
	// The logger will write only log messages with a level/V(x) equal or higher to the threshold.
	SetThreshold(threshold *int32)
	// Clone creates cloned copy of the logger
	Clone() LoggerImpl
	// CloneWithLevel creates cloned copy of the logger with updated log level
	CloneWithLevel(level int) LoggerImpl
	// WithCallDepth implements a New Option that allows to set the callDepth level for a logger.
	WithCallDepth(callDepth int) LoggerImpl
}

var l = NewLogger()

// Info logs a non-error message with the given key/value pairs as context.
func Info(msg string, kvs ...interface{}) {
	l.Print(msg, nil, logTypeINFO, kvs...)
}

// Infof logs a non-error message with the given key/value pairs as context.
func Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, logTypeINFO)
}

// Error logs an error message with the given key/value pairs as context.
func Error(err error, msg string, kvs ...interface{}) {
	l.Print(msg, err, logTypeERROR, kvs...)
}

// Errorf logs a error message with the given key/value pairs as context.
func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, logTypeERROR)
}

// Warning logs a warning messages with the given key/value pairs as context.
func Warning(msg string, kvs ...interface{}) {
	l.Print(msg, nil, logTypeWARN, kvs...)
}

// Warningf logs a warning messages with the given message format with format specifier and arguments.
func Warningf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, logTypeWARN)
}

// Success logs a success messages with the given key/value pairs as context.
func Success(msg string, kvs ...interface{}) {
	l.Print(msg, nil, logTypeSUCCESS, kvs...)
}

// Successf logs a success messages with the given message format with format specifier and arguments.
func Successf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, logTypeSUCCESS)
}

// Fatal logs a fatal message with the given key/value pairs as context and returns with os.Exit(1)
func Fatal(err error, msg string, kvs ...interface{}) {
	l.Print(msg, err, logTypeERROR, kvs...)
	os.Exit(1)
}

// Outputf writes a message to stdout
func Outputf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, logTypeOUTPUT)
}

// V returns an InfoLogger value for a specific verbosity level.
func V(level int) LoggerImpl {
	return l.CloneWithLevel(level)
}

// WithName adds a new element to the logger's name.
func WithName(name string) LoggerImpl {
	return l.Clone().WithName(name)
}

// WithValues adds some key-value pairs of context to a logger.
func WithValues(kvList ...interface{}) LoggerImpl {
	return l.Clone().WithValues(kvList...)
}

var logWriter = NewWriter()

// SetFile sets the logFile to writer
// if the non-empty file name is used, writer will also
// write the logs to this file
func SetFile(fileName string) {
	logWriter.SetFile(fileName)
}

// SetAuditLog sets the log file that should capture all logging activity. This
// file will contain all logging regardless of set verbosity level.
func SetAuditLog(fileName string) {
	logWriter.SetAuditLog(fileName)
}

// QuietMode sets the logging mode to quiet
// If this mode is set, writer will not write anything to stderr
func QuietMode(quiet bool) {
	logWriter.QuietMode(quiet)
}

// ShowTimestamp shows timestamp along with log
func ShowTimestamp(show bool) {
	logWriter.ShowTimestamp(show)
}

// SetVerbosity sets verbosity level and also updates default verbosity level
func SetVerbosity(verbosity int32) {
	l.SetThreshold(&verbosity)
	logWriter.SetVerbosity(verbosity)
}

// SetStdout overrides stdout
func SetStdout(writer io.Writer) {
	logWriter.SetStdout(writer)
}

// SetStderr overrides stderr
func SetStderr(writer io.Writer) {
	logWriter.SetStderr(writer)
}
