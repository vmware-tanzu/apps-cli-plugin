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

package logger_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/logger"
)

func TestSinkLogger(t *testing.T) {
	tests := []struct {
		name           string
		level          int
		expectedLogMsg string
	}{{
		name:           "lower level",
		level:          1,
		expectedLogMsg: "",
	}, {
		name:           "same level",
		level:          2,
		expectedLogMsg: "sample log msg name:test  key:value  \n",
	}, {
		name:           "higher level",
		level:          4,
		expectedLogMsg: "sample log msg name:test  key:value  \n",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			logsink := logger.NewSinkLogger("test", int32Ptr(test.level), buf)
			logsink.V(2).Info("sample log msg", "key", "value")
			if d := cmp.Diff(buf.String(), test.expectedLogMsg); d != "" {
				t.Errorf("expected logs (-wanted,+actual):%s", d)
			}
		})
	}
}

func TestSinkLoggerError(t *testing.T) {
	buf := new(bytes.Buffer)
	logsink := logger.NewSinkLogger("cli", int32Ptr(2), buf)
	err := fmt.Errorf("dial timeout")
	logsink.Error(err, "Sample error", "action", "get")
	expectedLogMsg := "Sample error name:cli  action:get  error:dial timeout  \n"
	if d := cmp.Diff(buf.String(), expectedLogMsg); d != "" {
		t.Errorf("expected logs (-wanted,+actual):%s", d)
	}
}

func int32Ptr(i int) *int32 {
	p := int32(i)
	return &p
}
