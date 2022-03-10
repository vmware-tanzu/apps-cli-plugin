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

package watch

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

type lwKey struct{}

func GetWatcher(ctx context.Context, c *cli.Config) (client.WithWatch, error) {
	if lw, ok := ctx.Value(lwKey{}).(client.WithWatch); ok {
		return lw, nil
	}
	// TODO: update reconciler runtime Client to include Watch func
	// and delete the following
	return client.NewWithWatch(c.KubeRestConfig(), client.Options{Scheme: c.Scheme})
}
func WithWatcher(ctx context.Context, lw client.WithWatch) context.Context {
	return context.WithValue(ctx, lwKey{}, lw)
}
