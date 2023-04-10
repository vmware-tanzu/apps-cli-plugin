/*
Copyright 2023 VMware, Inc.

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

package source_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

type FailureTransport struct {
	req *http.Request
}

type SuccessTransport struct {
	req *http.Request
}

func (t *FailureTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.req = req
	return nil, errors.New("failed impl")
}

func (t *SuccessTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.req = req
	response := &http.Response{
		Header: http.Header{
			"Lsp-Registry-Path": {"my-local-registry"},
		},
		StatusCode: 200,
	}
	return response, nil
}

func TestRoundTrip(t *testing.T) {
	req, err := http.NewRequest("POST", "http://www.my-fake-url.com/", nil)
	if err != nil {
		fmt.Println(err)
	}
	tests := []struct {
		name        string
		req         *http.Request
		wrapper     source.Wrapper
		shouldError bool
	}{{
		name: "fake successful transport",
		wrapper: source.Wrapper{
			Client: &http.Client{Transport: &SuccessTransport{}},
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.my-fake-url.com",
				Path:   "search",
			},
		},
	}, {
		name: "fake failure transport",
		wrapper: source.Wrapper{
			Client: &http.Client{Transport: &FailureTransport{}},
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.my-fake-url.com",
				Path:   "search",
			},
		},
		shouldError: true,
	}, {
		name: "nil transport",
		wrapper: source.Wrapper{
			Client: &http.Client{},
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.my-fake-url.com",
				Path:   "search",
			},
		},
		shouldError: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.wrapper.RoundTrip(req)

			if err != nil && !test.shouldError {
				t.Errorf("RoundTrip not expected to fail %v", err)
			}
		})
	}
}
