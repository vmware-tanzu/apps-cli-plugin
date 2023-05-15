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

package fake_source

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

type nilBodyRoundTripper struct {
	Headers http.Header
	Body    string
}

func (n nilBodyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Body:       io.NopCloser(strings.NewReader(n.Body)),
		Request:    req,
		Header:     n.Headers,
	}, nil
}

func GetFakeWrapper(headers http.Header) source.Wrapper {
	return source.Wrapper{
		Client: &http.Client{Transport: nilBodyRoundTripper{Headers: headers}},
		URL: &url.URL{
			Scheme: "http",
			Host:   "www.my-fake-url.com",
			Path:   "search",
		},
		Repository: "my-fake-repo",
	}
}
