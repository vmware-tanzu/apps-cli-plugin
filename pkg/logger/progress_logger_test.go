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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestProgressBar(t *testing.T) {
	var regexProgress = regexp.MustCompile(`[0-9]*\.*[0-9]*\s[a-zA-Z]+\s/\s[0-9]*[\.]*[0-9]*\s[a-zA-Z]+\s\[-+>*_*\]\s[0-9]*\.[0-9]+%.*`)
	tests := []struct {
		name              string
		regObject         []regv1.Update
		expectProgressBar bool
	}{{
		name:              "progressbar success",
		regObject:         []regv1.Update{{Total: 100, Complete: 100, Error: nil}},
		expectProgressBar: true,
	}, {
		name:              "progressbar with error",
		regObject:         []regv1.Update{{Total: 100, Complete: 20, Error: fmt.Errorf("test Error")}, {Total: 100, Complete: 100, Error: nil}},
		expectProgressBar: true,
	}, {
		name:              "progressbar end mid way",
		regObject:         []regv1.Update{{Total: 100, Complete: 20, Error: nil}, {Total: 100, Complete: 50, Error: nil}},
		expectProgressBar: true,
	}, {
		name:              "progressbar with total as 0",
		regObject:         []regv1.Update{{Total: 0, Complete: 0, Error: nil}},
		expectProgressBar: false,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// store os,stderr
			oldStdErr := os.Stderr
			r, w, _ := os.Pipe()
			// override with pipe writer
			os.Stderr = w
			outC := make(chan string)

			// copy the output in a separate goroutine so printing can't block indefinitely
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()

			progressLogger := NewProgressBar()
			valuesChannel := make(chan regv1.Update)
			progressLogger.Start(context.Background(), valuesChannel)
			for _, val := range test.regObject {
				valuesChannel <- val
				time.Sleep(100 * time.Millisecond)
			}

			w.Close()
			//restore os.stderr
			os.Stderr = oldStdErr
			out := <-outC

			defer progressLogger.End()
			defer close(valuesChannel)
			if test.expectProgressBar && regexProgress.FindString(out) == "" {
				t.Errorf("expected progress bar in the output \n: got %v", out)
			}
			if !test.expectProgressBar && regexProgress.FindString(out) != "" {
				t.Errorf("no progress bar expected in the output \n: got %v", out)
			}
		})
	}
}

func TestRetrieveProgressBarLogger(t *testing.T) {
	actualProgressLogger := NewProgressBar()
	ctx := StashProgressBarLogger(context.Background(), actualProgressLogger)
	expectedProgressLogger := RetrieveProgressBarLogger(ctx)
	if expectedProgressLogger != actualProgressLogger {
		t.Errorf("RetrieveProgressBarLogger() failed. wanted %v, got %v", actualProgressLogger, expectedProgressLogger)
	}
}

func TestStashProgressBarLogger(t *testing.T) {
	actualProgressLogger := NewProgressBar()
	ctx := StashProgressBarLogger(context.Background(), actualProgressLogger)
	if ctx.Value(progressLoggerKey{}).(ProgressLogger) != actualProgressLogger {
		t.Errorf("StashProgressBarLogger() failed. wanted %v, got %v", actualProgressLogger, ctx.Value(progressLoggerKey{}).(ProgressLogger))
	}
}
