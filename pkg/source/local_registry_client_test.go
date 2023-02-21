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
	"context"
	"net/http"
	"testing"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

func TestLocalRegistryTransport(t *testing.T) {
	config := &rest.Config{}
	fakeClient := &fake.RESTClient{
		Resp: &http.Response{StatusCode: http.StatusOK, Header: nil, Body: nil},
	}

	tests := []struct {
		name           string
		fakeRestClient *fake.RESTClient
		shouldError    bool
	}{{
		name: "success",
		fakeRestClient: &fake.RESTClient{
			Resp: &http.Response{StatusCode: http.StatusOK, Header: nil, Body: nil},
		},
	}, {
		name:           "fail",
		fakeRestClient: nil,
		shouldError:    true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := source.LocalRegistryTransport(context.Background(), config, fakeClient)
			if err != nil && !test.shouldError {
				t.Errorf("LocalRegistryTransport() not expected to fail %v", err)
			}
		})
	}
}
