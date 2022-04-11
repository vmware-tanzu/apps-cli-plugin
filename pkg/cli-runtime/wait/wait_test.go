/*
Copyright 2021 VMware, Inc.

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

package wait

import (
	"context"
	"fmt"
	"testing"
	"time"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
)

func TestUntilReady(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	workload := &cartov1alpha1.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      workloadName,
		},
		Status: cartov1alpha1.WorkloadStatus{
			Conditions: []metav1.Condition{
				{
					Type:   cartov1alpha1.WorkloadConditionReady,
					Status: metav1.ConditionUnknown,
				},
			},
		},
	}
	anotherWorkload := &cartov1alpha1.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      "anotherworkload",
		},
		Status: cartov1alpha1.WorkloadStatus{
			Conditions: []metav1.Condition{
				{
					Type:   cartov1alpha1.WorkloadConditionReady,
					Status: metav1.ConditionTrue,
				},
			},
		},
	}
	anotherWorkloadNs := &cartov1alpha1.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "anotherns",
			Name:      workloadName,
		},
		Status: cartov1alpha1.WorkloadStatus{
			Conditions: []metav1.Condition{
				{
					Type:   cartov1alpha1.WorkloadConditionReady,
					Status: metav1.ConditionTrue,
				},
			},
		},
	}
	tests := []struct {
		name               string
		resource           *cartov1alpha1.Workload
		additionalresource *cartov1alpha1.Workload
		condFunc           ConditionFunc
		err                error
	}{
		{
			name:     "transitions true",
			resource: workload.DeepCopy(),
			condFunc: func(c client.Object) (bool, error) {
				return true, nil
			},
		},
		{
			name:     "transitions false",
			resource: workload.DeepCopy(),
			condFunc: func(c client.Object) (bool, error) {
				return false, fmt.Errorf("failed to become ready: %s", "test not ready")
			},
			err: fmt.Errorf("failed to become ready: %s", "test not ready"),
		}, {
			name:     "mismatched names",
			resource: workload.DeepCopy(),
			condFunc: func(c client.Object) (bool, error) {
				return true, nil
			},
			additionalresource: anotherWorkload.DeepCopy(),
		},
		{
			name:     "mismatched namespaces",
			resource: workload.DeepCopy(),
			condFunc: func(c client.Object) (bool, error) {
				return true, nil
			},
			additionalresource: anotherWorkloadNs.DeepCopy(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = cartov1alpha1.AddToScheme(scheme)

			ctx := context.Background()
			objs := []client.Object{}
			if test.additionalresource != nil {
				objs = append(objs, test.additionalresource)
			}
			objs = append(objs, test.resource)

			fakeWithWatcher := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
			done := make(chan error, 1)
			defer close(done)
			go func() {
				done <- UntilCondition(ctx, fakeWithWatcher, types.NamespacedName{Name: test.resource.Name, Namespace: test.resource.Namespace}, &cartov1alpha1.WorkloadList{}, test.condFunc)
			}()

			for _, r := range objs {
				time.Sleep(10 * time.Millisecond)
				if err := fakeWithWatcher.Update(ctx, r); err != nil {
					t.Errorf("Update error %v", err)
				}
			}

			err := <-done
			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("expected error %v, actually %v", expected, actual)
			}
		})
	}
}

func TestUntilDelete(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	previousBackOffTime := BackOffTime
	defer func() {
		BackOffTime = previousBackOffTime
	}()
	BackOffTime = 10 * time.Millisecond
	workload := &cartov1alpha1.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      workloadName,
		},
		Status: cartov1alpha1.WorkloadStatus{
			Conditions: []metav1.Condition{
				{
					Type:   cartov1alpha1.WorkloadConditionReady,
					Status: metav1.ConditionUnknown,
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)
	gr := schema.GroupResource{
		Group:    cartov1alpha1.GroupName,
		Resource: "Workload",
	}

	tests := []struct {
		name        string
		resource    *cartov1alpha1.Workload
		timeout     time.Duration
		err         error
		reactorFunc clitesting.ReactionFunc
	}{{
		name:    "smaller timeout",
		timeout: time.Millisecond,
		err:     context.DeadlineExceeded,
	}, {
		name:    "API not found error",
		timeout: time.Second,
		reactorFunc: func(action clitesting.Action) (bool, runtime.Object, error) {
			if _, ok := action.(clitesting.GetAction); ok {
				return true, workload, apierrs.NewNotFound(gr, workloadName)
			}
			// never handle the action
			return false, nil, nil
		},
	}, {
		name:    "API error",
		timeout: 15 * time.Second,
		err:     fmt.Errorf("client error"),
		reactorFunc: func(action clitesting.Action) (bool, runtime.Object, error) {
			if _, ok := action.(clitesting.GetAction); ok {
				return true, nil, fmt.Errorf("client error")
			}
			// never handle the action
			return false, nil, nil
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, test.timeout)
			defer cancelFunc()
			client := clitesting.NewFakeClient(scheme, workload.DeepCopy())

			// reactor fails with retryable error for 1st call and then fails with ResourseNotFoundError
			client.AddReactor("get", "*", test.reactorFunc)
			err := UntilDelete(ctx, client, workload)
			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("expected error %v, actually %v", expected, actual)
			}
		})
	}
}
