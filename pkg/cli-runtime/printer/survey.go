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

package printer

import (
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func WithSurveyStdio(stdin io.Reader, stdout, stderr io.Writer) survey.AskOpt {
	in, ok := stdin.(terminal.FileReader)
	if !ok {
		in = fileReaderWrapper{Reader: stdin}
	}
	out, ok := stdout.(terminal.FileWriter)
	if !ok {
		out = fileWriterWrapper{Writer: stdout}
	}
	return survey.WithStdio(in, out, stderr)
}

type fileReaderWrapper struct {
	io.Reader
}

func (fileReaderWrapper) Fd() uintptr {
	return 0
}

type fileWriterWrapper struct {
	io.Writer
}

func (fileWriterWrapper) Fd() uintptr {
	return 1
}
