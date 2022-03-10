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

package testing

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

func ValidateCreates(ctx context.Context, action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	got := action.(clientgotesting.CreateAction).GetObject()
	obj, ok := got.(validation.FieldValidator)
	if !ok {
		return false, nil, nil
	}
	if err := obj.Validate().ToAggregate(); err != nil {
		return true, nil, err
	}
	return false, nil, nil
}

func ValidateUpdates(ctx context.Context, action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	got := action.(clientgotesting.UpdateAction).GetObject()
	obj, ok := got.(validation.FieldValidator)
	if !ok {
		return false, nil, nil
	}
	if err := obj.Validate().ToAggregate(); err != nil {
		return true, nil, err
	}
	return false, nil, nil
}
