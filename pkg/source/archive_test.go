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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			got, err := HandleZip(test.file)
			defer os.RemoveAll(got)
			if (err == nil) == test.shouldErr {
				t.Errorf("HandleZip() shouldErr %t %v", test.shouldErr, err)
			} else if test.shouldErr {
				return
			}
			err = filepath.Walk(got,
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

					if diff := cmp.Diff(wantFile, gotFile); diff != "" {
						t.Errorf("HandleZip() (-want, +got) = %v", diff)
					}
					return nil
				})
			if err != nil {
				t.Errorf("unexpected error comparing files: %v", err)
			}
		})
	}
}
