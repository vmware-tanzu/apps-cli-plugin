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
package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	localSourceProxyRegistryPath = "Lsp-Registry-Path"
)

// Wrapper implements RoundTripper by appending request path and parameters to
// its URL.
type Wrapper struct {
	Client     *http.Client
	URL        *url.URL
	Repository string
}

type containerWrapperStashKey struct{}

func StashContainerWrapper(ctx context.Context, wrapper Wrapper) context.Context {
	return context.WithValue(ctx, containerWrapperStashKey{}, wrapper)
}

func RetrieveContainerWrapper(ctx context.Context) *Wrapper {
	wrapper, ok := ctx.Value(containerWrapperStashKey{}).(Wrapper)
	if !ok {
		return nil
	}

	return &wrapper
}

// RoundTrip implements the http.RoundTripper interface.
func (w *Wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	if w.Client.Transport == nil {
		return nil, fmt.Errorf("client transport not provided")
	}

	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(w.URL.String())
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(req.URL.Path, "/api/v1/namespaces/") {
		u.Path = path.Join(w.URL.Path, req.URL.Path)
	} else {
		u.Path = req.URL.Path
	}

	if strings.HasSuffix(u.Path, "v2") || strings.HasSuffix(u.Path, "uploads") {
		u.Path = u.Path + "/"
	}

	u.RawQuery = params.Encode()
	req.URL = u

	resp, err := w.Client.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		for k, vs := range resp.Header {
			for _, v := range vs {
				if k == localSourceProxyRegistryPath {
					w.Repository = v
					break
				}
			}
		}
	}

	return resp, err
}
