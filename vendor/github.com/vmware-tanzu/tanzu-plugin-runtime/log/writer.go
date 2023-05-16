// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"io"
	"os"
	"path"
)

// Writer defines methods to write and configure tkg writer
type Writer interface {

	// Write writes message to stdout/stderr, logfile and sends to the channel if channel is set
	//
	// header is message header to append as a message prefix
	// msg is actual message to write
	// logEnabled is used to decide whether to write this message to stdout/stderr or not
	// logVerbosity is used to decide which message to write for different output types
	// logType used to decide should write to stdout or stderr
	Write(header []byte, msg []byte, logEnabled bool, logVerbosity int32, logType string) (n int, err error)

	// SetFile sets the logFile to writer
	// if the non-empty file name is used, writer will also
	// write the logs to this file
	SetFile(fileName string)

	// SetAuditLog sets the log file that should capture all logging activity. This
	// file will contain all logging regardless of set verbosity level.
	SetAuditLog(fileName string)

	// QuietMode sets the logging mode to quiet
	// If this mode is set, writer will not write anything to stderr
	QuietMode(quiet bool)

	// ShowTimestamp shows timestamp along with log
	ShowTimestamp(show bool)

	// SetVerbosity sets verbosity level and also updates default verbosity level
	SetVerbosity(verbosity int32)

	// SetStdout overrides stdout
	SetStdout(w io.Writer)

	// SetStderr overrides stderr
	SetStderr(w io.Writer)
}

var defaultVerbosity int32 = 4

// NewWriter returns new custom writer
func NewWriter() Writer {
	return &writer{}
}

type writer struct {
	logFile       string
	verbosity     int32
	quiet         bool
	auditFile     string
	showTimestamp bool
	stdout        io.Writer
	stderr        io.Writer
}

// SetFile sets the logFile to writer
// if the non-empty file name is used, writer will also
// write the logs to this file
func (w *writer) SetFile(fileName string) {
	w.logFile = fileName
}

// SetAuditLog sets the log file that should capture all logging activity. This
// file will contain all logging regardless of set verbosity level.
func (w *writer) SetAuditLog(fileName string) {
	w.auditFile = fileName
}

// QuietMode sets the logging mode to quiet
// If this mode is set, writer will not write anything to stderr
func (w *writer) QuietMode(quiet bool) {
	w.quiet = quiet
}

// SetVerbosity sets verbosity level and also updates default verbosity level
func (w *writer) SetVerbosity(verbosity int32) {
	w.verbosity = verbosity
	if verbosity > defaultVerbosity {
		defaultVerbosity = verbosity
	}
}

// ShowTimestamp shows timestamp along with log
func (w *writer) ShowTimestamp(show bool) {
	w.showTimestamp = show
}

// SetStdout overrides stdout
func (w *writer) SetStdout(writer io.Writer) {
	w.stdout = writer
}

// SetStderr overrides stderr
func (w *writer) SetStderr(writer io.Writer) {
	w.stderr = writer
}

// Write writes message to stdout/stderr, logfile
//
// header is message header to append as a message prefix
// msg is actual message to write
// logEnabled is used to decide whether to write this message to stdout/stderr or not
// logVerbosity is used to decide which message to write for different output types
// logType used to decide should write to stdout or stderr
func (w *writer) Write(header, msg []byte, logEnabled bool, logVerbosity int32, logType string) (n int, err error) {
	fullMsg := append(header, msg...) //nolint:gocritic

	// Always write to the audit log so it captures everything
	if w.auditFile != "" {
		w.fileWriter(w.auditFile, fullMsg)
	}

	// write to logfile only if verbosityLevel is <= default VerbosityLevel
	if logVerbosity <= defaultVerbosity {
		if w.logFile != "" {
			w.fileWriter(w.logFile, fullMsg)
		}
	}

	// If showTimestamp is set to true, log fullMsg with header
	// If log type is OUTPUT skip the header
	if w.showTimestamp && logType != logTypeOUTPUT {
		msg = fullMsg
	}

	// write to stdout/stderr if quiet mode is not set and logEnabled is true
	if !w.quiet && logEnabled {
		if logType == logTypeOUTPUT {
			w.stdoutWriter(msg)
		} else {
			w.stderrWriter(msg)
		}
	}

	return len(msg), nil
}

func (w *writer) stdoutWriter(msg []byte) {
	if w.stdout != nil {
		_, _ = w.stdout.Write(msg)
	} else {
		_, _ = os.Stdout.Write(msg)
	}
}

func (w *writer) stderrWriter(msg []byte) {
	if w.stderr != nil {
		_, _ = w.stderr.Write(msg)
	} else {
		_, _ = os.Stderr.Write(msg)
	}
}

func (w *writer) fileWriter(logFileName string, msg []byte) {
	basePath := path.Dir(logFileName)
	filePath := path.Base(logFileName)

	if os.MkdirAll(basePath, 0o600) != nil {
		msg := "Unable to create log directory: " + basePath
		w.stderrWriter([]byte(msg))
		os.Exit(1)
	}

	logFileName = path.Join(basePath, filePath)
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		w.stderrWriter([]byte(err.Error()))
		os.Exit(1)
	}
	defer file.Close()

	_, fileWriteErr := file.Write(msg)
	if fileWriteErr != nil {
		w.stderrWriter([]byte(fileWriteErr.Error()))
		os.Exit(1) //nolint:gocritic
	}
}
