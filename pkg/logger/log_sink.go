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
	"fmt"
	"io"

	"github.com/go-logr/logr"
)

// writerLogSink is a logr.Logger that logs to io.Writer interface.
type writerLogSink struct {
	name   string
	writer io.Writer
	level  *int32
}

var _ logr.Logger = &writerLogSink{}

func (l *writerLogSink) Info(msg string, kvs ...interface{}) {
	fmt.Fprintf(l.writer, "%s ", msg)
	if l.name != "" {
		fmt.Fprintf(l.writer, "%s:%+v  ", "name", l.name)
	}
	for i := 0; i < len(kvs); i += 2 {
		fmt.Fprintf(l.writer, "%s:%+v  ", kvs[i], kvs[i+1])
	}
	fmt.Fprintf(l.writer, "\n")
}

func (_ *writerLogSink) Enabled() bool {
	return true
}

func (l *writerLogSink) Error(err error, msg string, kvs ...interface{}) {
	kvs = append(kvs, "error", err)
	l.Info(msg, kvs...)
}

func (l *writerLogSink) V(level int) logr.Logger {
	if *l.level < int32(level) {
		return logr.Discard()
	}
	return l
}

func (l *writerLogSink) WithName(name string) logr.Logger {
	return l
}

func (l *writerLogSink) WithValues(kvs ...interface{}) logr.Logger {
	return l
}

func NewSinkLogger(name string, level *int32, writer io.Writer) logr.Logger {
	return &writerLogSink{
		name:   name,
		writer: writer,
		level:  level,
	}
}
