/*
Copyright 2019 VMware, Inc.

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

package testing_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	clitestingresource "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing/resource"
)

func TestValidateCreates(t *testing.T) {
	tests := []struct {
		name      string
		action    clientgotesting.Action
		handled   bool
		returned  runtime.Object
		shouldErr bool
	}{{
		name: "valid create",
		action: clientgotesting.NewCreateAction(
			clitestingresource.SchemeBuilder.GroupVersion.WithResource("TestResource"),
			"default",
			&clitestingresource.TestResource{},
		),
		handled:   false,
		shouldErr: false,
	}, {
		name: "invalid create",
		action: clientgotesting.NewCreateAction(
			clitestingresource.SchemeBuilder.GroupVersion.WithResource("TestResource"),
			"default",
			&clitestingresource.TestResource{
				Spec: clitestingresource.TestResourceSpec{
					Fields: map[string]string{
						"invalid": "true",
					},
				},
			},
		),
		handled:   true,
		shouldErr: true,
	}, {
		name: "not validatable",
		action: clientgotesting.NewCreateAction(
			clitestingresource.SchemeBuilder.GroupVersion.WithResource("TestResource"),
			"default",
			&clitestingresource.TestResource{},
		),
		handled:   false,
		shouldErr: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()
			handled, returned, err := clitesting.ValidateCreates(ctx, test.action)
			if expected, actual := test.handled, handled; expected != actual {
				t.Errorf("Expected handled %v, actually %v", expected, actual)
			}
			if diff := cmp.Diff(test.returned, returned); diff != "" {
				t.Errorf("Unexpected returned object (-expected, +actual): %s", diff)
			}
			if (err != nil) != test.shouldErr {
				t.Errorf("Expected error %v, actually %q", test.shouldErr, err)
			}
		})
	}
}

func TestValidateUpdates(t *testing.T) {
	tests := []struct {
		name      string
		action    clientgotesting.Action
		handled   bool
		returned  runtime.Object
		shouldErr bool
	}{{
		name: "valid create",
		action: clientgotesting.NewUpdateAction(
			clitestingresource.SchemeBuilder.GroupVersion.WithResource("TestResource"),
			"default",
			&clitestingresource.TestResource{},
		),
		handled:   false,
		shouldErr: false,
	}, {
		name: "invalid create",
		action: clientgotesting.NewUpdateAction(
			clitestingresource.SchemeBuilder.GroupVersion.WithResource("TestResource"),
			"default",
			&clitestingresource.TestResource{
				Spec: clitestingresource.TestResourceSpec{
					Fields: map[string]string{
						"invalid": "true",
					},
				},
			},
		),
		handled:   true,
		shouldErr: true,
	}, {
		name: "not validatable",
		action: clientgotesting.NewUpdateAction(
			clitestingresource.SchemeBuilder.GroupVersion.WithResource("TestResourceNoStatus"),
			"default",
			&clitestingresource.TestResource{},
		),
		handled:   false,
		shouldErr: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.TODO()
			handled, returned, err := clitesting.ValidateUpdates(ctx, test.action)
			if expected, actual := test.handled, handled; expected != actual {
				t.Errorf("Expected handled %v, actually %v", expected, actual)
			}
			if diff := cmp.Diff(test.returned, returned); diff != "" {
				t.Errorf("Unexpected returned object (-expected, +actual): %s", diff)
			}
			if (err != nil) != test.shouldErr {
				t.Errorf("Expected error %v, actually %q", test.shouldErr, err)
			}
		})
	}
}
