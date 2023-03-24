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

package printer

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/lsp"
)

func TestPrintLocalSourceProxyStatus(t *testing.T) {
	status := lsp.LSPStatus{
		UserHasPermission:     true,
		Reachable:             true,
		UpstreamAuthenticated: true,
		OverallHealth:         true,
		Message:               "my cool status",
	}
	type args struct {
		printType string
		s         lsp.LSPStatus
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				printType: "json",
				s:         status,
			},
			wantW: `{
  "user_has_permission": true,
  "reachable": true,
  "upstream_authenticated": true,
  "overall_health": true,
  "message": "my cool status"
}`,
		},
		{
			name: "yaml",
			args: args{
				printType: "yaml",
				s:         status,
			},
			wantW: `user_has_permission: true
reachable: true
upstream_authenticated: true
overall_health: true
message: my cool status
`,
		},
		{
			name: "yml",
			args: args{
				printType: "yml",
				s:         status,
			},
			wantW: `user_has_permission: true
reachable: true
upstream_authenticated: true
overall_health: true
message: my cool status
`,
		},
		{
			name: "text",
			args: args{
				printType: "",
				s:         status,
			},
			wantW: `user_has_permission: true
reachable: true
upstream_authenticated: true
overall_health: true
message: my cool status
`,
		},
		{
			name: "wrong type",
			args: args{
				printType: "foo",
				s:         status,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := PrintLocalSourceProxyStatus(w, tt.args.printType, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("PrintLocalSourceProxyStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantW, w.String()); diff != "" {
				t.Errorf("PrintLocalSourceProxyStatus(): Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
