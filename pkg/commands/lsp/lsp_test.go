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
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/lsp"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
)

func TestGetStatus(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *cli.Config
	}
	tests := []struct {
		name    string
		args    args
		want    lsp.HealthStatus
		wantErr bool
	}{
		{
			name: "Cluster and LSP 200 OK",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, `{"statuscode": "200", "message": "any ignored message"}`)),
			},
			want: lsp.HealthStatus{
				UserHasPermission:     true,
				Reachable:             true,
				UpstreamAuthenticated: true,
				OverallHealth:         true,
				Message:               "All health checks passed",
			},
		},
		{
			name: "Cluster 403 Forbidden",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusUnauthorized, (`{"message": "403 Forbidden"}`))),
			},
			want: lsp.HealthStatus{
				Message: "The current user does not have permission to access the local source proxy.\nErrors:\n- 403 Forbidden",
			},
		},
		{
			name: "Cluster 401 Unauthorized",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusUnauthorized, (`{"message": "401 Unauthorized"}`))),
			},
			want: lsp.HealthStatus{
				Message: "The current user does not have permission to access the local source proxy.\nErrors:\n- 401 Unauthorized",
			},
		},
		{
			name: "Cluster 500 Internal Server Error",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusInternalServerError, (`{"message": "500 Internal Server Error"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy is not healthy.\nErrors:\n- 500 Internal Server Error",
			},
		},
		{
			name: "Cluster 504 Gateway Timeout",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusGatewayTimeout, (`{"message": "504 GatewayTimeout"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy is not healthy.\nErrors:\n- 504 GatewayTimeout",
			},
		},
		{
			name: "Cluster 503 Service Unavailable",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusServiceUnavailable, (`{"message": "503 Service Unavailable"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Message:           "Local source proxy is not healthy.\nErrors:\n- 503 Service Unavailable",
			},
		},
		{
			name: "Cluster 404 Not Found",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusNotFound, (`{"message": "404 Not Found"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Message:           "Local source proxy is not installed on the cluster.\nErrors:\n- 404 Not Found",
			},
		},

		{
			name: "Cluster 200 OK and LSP 302 Found",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, `{"statuscode": "302", "message": "Found"}`)),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Found",
			},
		},
		{
			name: "Cluster 200 OK and LSP 400 Bad Request",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, (`{"statuscode": "400", "message": "Bad Request"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Bad Request",
			},
		},
		{
			name: "Cluster 200 OK and LSP 401 Unauthorized",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, (`{"statuscode": "401", "message": "Unauthorized"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Unauthorized",
			},
		},
		{
			name: "Cluster 200 OK and LSP 403 Forbidden",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, (`{"statuscode": "403", "message": "Forbidden"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Forbidden",
			},
		},
		{
			name: "Cluster 200 OK and LSP 404 Not Found",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, (`{"statuscode": "404", "message": "Not Found"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Not Found",
			},
		},
		{
			name: "Cluster 200 OK and LSP 500 Internal Server Error",
			args: args{
				ctx: context.Background(),
				c:   getConfig(getResponse(http.StatusOK, (`{"statuscode": "500", "message": "Internal Server Error"}`))),
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- Internal Server Error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStatus(tt.args.ctx, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetStatus(): Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func Test_checkRequestResponseCode(t *testing.T) {
	msg := "my cool message"
	type args struct {
		resp *http.Response
		msg  string
	}
	tests := []struct {
		name string
		args args
		want *lsp.HealthStatus
	}{
		{
			name: http.StatusText(http.StatusOK),
			args: args{
				resp: getResponse(http.StatusOK, ""),
			},
			want: nil,
		},
		{
			name: http.StatusText(http.StatusForbidden),
			args: args{
				resp: getResponse(http.StatusForbidden, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				Message: `The current user does not have permission to access the local source proxy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusUnauthorized),
			args: args{
				resp: getResponse(http.StatusUnauthorized, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				Message: `The current user does not have permission to access the local source proxy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusInternalServerError),
			args: args{
				resp: getResponse(http.StatusInternalServerError, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message: `Local source proxy is not healthy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusNotImplemented),
			args: args{
				resp: getResponse(http.StatusNotImplemented, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message: `Local source proxy is not healthy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusGatewayTimeout),
			args: args{
				resp: getResponse(http.StatusGatewayTimeout, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message: `Local source proxy is not healthy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusServiceUnavailable),
			args: args{
				resp: getResponse(http.StatusServiceUnavailable, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				UserHasPermission: true,
				Message: `Local source proxy is not healthy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusNotFound),
			args: args{
				resp: getResponse(http.StatusNotFound, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				UserHasPermission: true,
				Message: `Local source proxy is not installed on the cluster.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusRequestURITooLong),
			args: args{
				resp: getResponse(http.StatusRequestURITooLong, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				Message: `The request is not valid for the query the health of the ocal source proxy.
Errors:
- my cool message`,
			},
		},
		{
			name: http.StatusText(http.StatusPermanentRedirect),
			args: args{
				resp: getResponse(http.StatusPermanentRedirect, ""),
				msg:  msg,
			},
			want: &lsp.HealthStatus{
				Message: `Local source proxy was moved and is not reachable in the defined url.
Errors:
- my cool message`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRequestResponseCode(tt.args.resp, tt.args.msg)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("checkRequestResponseCode(): Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func Test_getStatusFromLSPResponse(t *testing.T) {
	msg := "my cool message"
	respMsg := "Local source proxy was unable to authenticate against the target registry.\nErrors:\n- my cool message"
	type args struct {
		r lspResponse
	}
	tests := []struct {
		name    string
		args    args
		want    lsp.HealthStatus
		wantErr bool
	}{
		{
			name: "No status code",
			args: args{
				r: lspResponse{Message: msg},
			},
			wantErr: true,
		},
		{
			name: "200 OK",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusOK),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission:     true,
				Reachable:             true,
				UpstreamAuthenticated: true,
				OverallHealth:         true,
				Message:               "All health checks passed",
			},
		},
		{
			name: "204 No Content",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusNoContent),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission:     true,
				Reachable:             true,
				UpstreamAuthenticated: true,
				OverallHealth:         true,
				Message:               "All health checks passed",
			},
		},
		{
			name: "302 Found",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusFound),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "400 Bad Request",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusBadRequest),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "401 Unauthorized",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusUnauthorized),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "403 Forbidden",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusForbidden),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "404 Not Found",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusNotFound),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "500 Internal Server Error",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusInternalServerError),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "501 Not Implemented",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusNotImplemented),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "502 Bad Gateway",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusBadGateway),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
		{
			name: "503 Service Unavailable",
			args: args{
				r: lspResponse{
					StatusCode: strconv.Itoa(http.StatusServiceUnavailable),
					Message:    msg,
				},
			},
			want: lsp.HealthStatus{
				UserHasPermission: true,
				Reachable:         true,
				Message:           respMsg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStatusFromLSPResponse(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("getStatusFromLSPResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getStatusFromLSPResponse(): Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func getResponse(status int, body string) *http.Response {
	return &http.Response{
		Status:     http.StatusText(status),
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func getConfig(resp *http.Response) *cli.Config {
	scheme := k8sruntime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)
	c.Client = clitesting.NewFakeCliClientWithTransport(clitesting.NewFakeClient(scheme), clitesting.NewFakeTransportFromResponse(resp))
	return c
}
