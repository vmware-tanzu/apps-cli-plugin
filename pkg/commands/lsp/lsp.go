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

package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/lsp"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

const errFormat = "%s\nMessages:\n- %s"

type lspResponse struct {
	Message    string `json:"message"`
	StatusCode string `json:"statuscode"`
}

func (l lspResponse) getDecodedMessage() string {
	m, err := url.QueryUnescape(l.Message)
	if err != nil {
		return l.Message
	}
	return m
}

func GetStatus(ctx context.Context, c *cli.Config) (lsp.HealthStatus, error) {
	r := &lspResponse{}
	var localTransport *source.Wrapper
	var resp *http.Response
	var err error

	if localTransport, err = source.LocalRegistryTransport(ctx, c.KubeRestConfig(), c.GetClientSet().CoreV1().RESTClient(), "health"); err != nil {
		return lsp.HealthStatus{}, err
	}
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, localTransport.URL.Path, nil); err != nil {
		return lsp.HealthStatus{}, err
	} else if resp, err = localTransport.RoundTrip(req); err != nil {
		return lsp.HealthStatus{}, err
	}

	if b, err := io.ReadAll(resp.Body); err != nil {
		return lsp.HealthStatus{}, err
	} else if err := json.Unmarshal(b, r); err != nil {
		r = &lspResponse{Message: string(b)}
	}

	if s := checkRequestResponseCode(resp, r.getDecodedMessage()); s != nil {
		return *s, nil
	}

	return getStatusFromLSPResponse(*r)
}

func checkRequestResponseCode(resp *http.Response, msg string) *lsp.HealthStatus {
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return &lsp.HealthStatus{
			Message: fmt.Sprintf(errFormat, "The current user does not have permission to access the local source proxy.", msg),
		}
	}

	if resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode != http.StatusServiceUnavailable {
		return &lsp.HealthStatus{
			UserHasPermission: true,
			Reachable:         true,
			Message:           fmt.Sprintf(errFormat, "Local source proxy is not healthy.", msg),
		}
	}

	if resp.StatusCode == http.StatusServiceUnavailable {
		return &lsp.HealthStatus{
			UserHasPermission: true,
			Message:           fmt.Sprintf(errFormat, "Local source proxy is not healthy.", msg),
		}
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return &lsp.HealthStatus{
			Message: fmt.Sprintf(errFormat, "The request is not valid for the query the health of the ocal source proxy.", msg),
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return &lsp.HealthStatus{
			UserHasPermission: true,
			Message:           fmt.Sprintf(errFormat, "Local source proxy is not installed on the cluster.", msg),
		}
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return &lsp.HealthStatus{
			Message: fmt.Sprintf(errFormat, "Local source proxy was moved and is not reachable in the defined url.", msg),
		}
	}
	return nil
}

func getStatusFromLSPResponse(r lspResponse) (lsp.HealthStatus, error) {
	if r.StatusCode == "" {
		return lsp.HealthStatus{}, fmt.Errorf("unable to read local source proxy response: %+v", r)
	}

	if s, err := strconv.Atoi(r.StatusCode); err == nil {
		switch {
		case s >= http.StatusOK && s < http.StatusMultipleChoices:
			return lsp.HealthStatus{
				UserHasPermission:     true,
				Reachable:             true,
				UpstreamAuthenticated: true,
				OverallHealth:         true,
				Message:               "All health checks passed",
			}, nil
		default:
			return lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           fmt.Sprintf(errFormat, "Local source proxy was unable to authenticate against the target registry.", r.getDecodedMessage()),
			}, nil
		}
	} else {
		return lsp.HealthStatus{}, err
	}
}
