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

package fake

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ client.WithWatch = &FakeWithWatch{}

type FakeWithWatch struct {
	err bool
	client.Client
	events []watch.Event
}

func NewFakeWithWatch(throwErr bool, client client.Client, events []watch.Event) *FakeWithWatch {
	return &FakeWithWatch{
		err:    throwErr,
		Client: client,
		events: events,
	}
}
func (c *FakeWithWatch) Watch(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
	if c.err {
		return nil, fmt.Errorf("failed to create watcher")
	}
	watcher := watch.NewRaceFreeFake()
	go func() {
		for _, event := range c.events {
			if !watcher.IsStopped() {
				switch event.Type {
				case watch.Added:
					watcher.Add(event.Object)
				case watch.Modified:
					watcher.Modify(event.Object)
				case watch.Deleted:
					watcher.Delete(event.Object)
				default:
					return
				}
			}
		}
	}()
	return watcher, nil
}
