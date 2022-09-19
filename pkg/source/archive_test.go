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
package source

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg"
)

func TestIsDir(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		dirName  string
	}{{
		name:     "invalid dir",
		expected: false,
		dirName:  "testdata/helloworld.jar",
	}, {
		name:     "valid dir",
		expected: true,
		dirName:  "testdata/hello_jar",
	}, {
		name:     "non existent file",
		expected: false,
		dirName:  "testdata/nondir",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsDir(test.dirName)
			if actual != test.expected {
				t.Errorf("IsDir() errored; expected %v actual %v", test.expected, actual)
			}
		})
	}
}
func TestIsZip(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		fileName string
	}{{
		name:     "valid jar",
		expected: true,
		fileName: "testdata/helloworld.jar",
	}, {
		name:     "valid zip",
		expected: true,
		fileName: "testdata/hello.go.zip",
	}, {
		name:     "invalid jar",
		expected: false,
		fileName: "testdata/invalid.jar",
	}, {
		name:     "invalid zip",
		expected: false,
		fileName: "testdata/invalid.zip",
	}, {
		name:     "non existing file",
		fileName: "testdata/non_file",
		expected: false,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsZip(test.fileName)
			if actual != test.expected {
				t.Errorf("IsZip() errored; expected %v actual %v", test.expected, actual)
			}
		})
	}
}

func TestHandleZip(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		want      string
		shouldErr bool
	}{{
		name:      "valid jar",
		file:      "testdata/hello.go.jar",
		shouldErr: false,
		want:      "testdata/hello_jar",
	}, {
		name:      "Java jar",
		file:      "testdata/helloworld.jar",
		shouldErr: false,
		want:      "testdata/helloworld_java",
	}, {
		name:      "non existing file",
		file:      "testdata/non_file",
		shouldErr: true,
	}, {
		name:      "valid zip",
		file:      "testdata/hello.go.zip",
		shouldErr: false,
		want:      "testdata/hello_zip",
	}, {
		name:      "invalid zip",
		file:      "testdata/invalid.zip",
		shouldErr: true,
	}, {
		name:      "invalid jar",
		file:      "testdata/invalid.jar",
		shouldErr: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "")
			defer os.RemoveAll(tmpDir)
			if err != nil {
				t.Error("failed to create temp dir", err)
				return
			}

			err = ExtractZip(tmpDir, test.file)
			if (err == nil) == test.shouldErr {
				t.Errorf("ExtractZip() shouldErr %t %v", test.shouldErr, err)
			} else if test.shouldErr {
				return
			}
			err = filepath.Walk(tmpDir,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					gotFile, err1 := ioutil.ReadFile(path)
					if err1 != nil {
						return err
					}

					wantFile, err2 := ioutil.ReadFile(filepath.Join(test.want, info.Name()))
					if err2 != nil {
						return err
					}

					if diff := cmp.Diff(normalizeNewlines(wantFile), normalizeNewlines(gotFile)); diff != "" {
						t.Errorf("ExtractZip() (-want, +got) = %v", diff)
					}
					return nil
				})
			if err != nil {
				t.Errorf("unexpected error comparing files: %v", err)
			}
		})
	}
}

func normalizeNewlines(d []byte) string {
	// replace CR LF \r\n (windows) with LF \n (unix)
	normalizedOutput := strings.ReplaceAll(string(d), fmt.Sprintf("%s%s", pkg.CR, pkg.LF), pkg.LF)
	// replace CR \r (mac) with LF \n (unix)
	normalizedOutput = strings.ReplaceAll(normalizedOutput, pkg.CR, pkg.LF)
	return normalizedOutput
}
