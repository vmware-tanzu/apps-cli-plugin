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

package logs

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

type Tailer interface {
	Tail(ctx context.Context, c *cli.Config, namespace string, selector labels.Selector, containers []string, since time.Duration, timestamps bool) error
}

func Tail(ctx context.Context, c *cli.Config, namespace string, selector labels.Selector, containers []string, since time.Duration, timestamps bool) error {
	tailer := RetrieveTailer(ctx)
	if tailer == nil {
		return fmt.Errorf("unable to retrieve tailer from the context: set the tailer on context with StashTailer(ctx context.Context, tailer Tailer) context.Context")
	}
	return tailer.Tail(ctx, c, namespace, selector, containers, since, timestamps)
}

var tailerStashKey = struct{}{}

func StashTailer(ctx context.Context, tailer Tailer) context.Context {
	return context.WithValue(ctx, tailerStashKey, tailer)
}

func RetrieveTailer(ctx context.Context) Tailer {
	value := ctx.Value(tailerStashKey)
	if tailer, ok := value.(Tailer); ok {
		return tailer
	}
	return nil
}
