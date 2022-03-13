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

var _ logr.LogSink = &writerLogSink{}

func (*writerLogSink) Init(info logr.RuntimeInfo) {

}

func (*writerLogSink) Enabled(level int) bool {
	return true
}

func (l *writerLogSink) Info(level int, msg string, kvs ...interface{}) {
	if *l.level < int32(level) {
		return
	}

	fmt.Fprintf(l.writer, "%s ", msg)
	if l.name != "" {
		fmt.Fprintf(l.writer, "%s:%+v  ", "name", l.name)
	}
	for i := 0; i < len(kvs); i += 2 {
		fmt.Fprintf(l.writer, "%s:%+v  ", kvs[i], kvs[i+1])
	}

	fmt.Fprintf(l.writer, "\n")
}

func (l *writerLogSink) Error(err error, msg string, kvs ...interface{}) {
	kvs = append(kvs, "error", err)
	l.Info(0, msg, kvs...)
}

func (l *writerLogSink) WithName(name string) logr.LogSink {
	return &writerLogSink{
		name:   l.name + "." + name,
		writer: l.writer,
		level:  l.level,
	}
}

func (l *writerLogSink) WithValues(kvs ...interface{}) logr.LogSink {
	return l
}

func NewSinkLogger(name string, level *int32, writer io.Writer) logr.Logger {
	sink := &writerLogSink{
		name:   name,
		writer: writer,
		level:  level,
	}

	return logr.New(sink)
}
