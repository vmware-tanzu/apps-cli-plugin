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
	"sync"
	"time"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	BackOffTime = 5 * time.Second
)

type ConditionFunc = func(client.Object) (bool, error)

func UntilCondition(ctx context.Context, watchClient client.WithWatch, target types.NamespacedName, listType client.ObjectList, condition ConditionFunc) error {
	eventWatcher, err := watchClient.Watch(ctx, listType, &client.ListOptions{Namespace: target.Namespace})
	if err != nil {
		return err
	}
	defer eventWatcher.Stop()
	for {
		select {
		case event := <-eventWatcher.ResultChan():
			obj := event.Object.(client.Object)
			if obj.GetName() != target.Name || obj.GetNamespace() != target.Namespace {
				continue
			}
			cond, err := condition(obj)
			if err != nil {
				return err
			}
			if cond {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func UntilDelete(ctx context.Context, c client.Client, obj client.Object) error {
	t := time.NewTicker(BackOffTime)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := c.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, obj); err != nil {
				if apierrs.IsNotFound(err) {
					return nil
				}
				return err
			}
		}
	}
}

type Worker func(context.Context) error

// Race multiple worker functions each in a goroutine. The first worker to return
// commits the result of the Race function. All workers must return when the context
// is closed before the Race function will return.
func Race(ctx context.Context, timeout time.Duration, workers []Worker) error {
	var wg sync.WaitGroup
	output := make(chan error, len(workers)+1)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		output <- ctx.Err()
	}()

	for _, worker := range workers {
		wg.Add(1)
		go func(worker Worker) {
			defer wg.Done()
			defer cancel()
			output <- worker(ctx)
		}(worker)
	}

	wg.Wait()
	return <-output
}
