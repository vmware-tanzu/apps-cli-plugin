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

package cli

import (
	"context"
	"fmt"
	"io"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

type DryRunable interface {
	IsDryRun() bool
}

func DryRunResource(ctx context.Context, resource runtime.Object, gvk schema.GroupVersionKind) {
	stdout := StdoutFromContext(ctx)
	resource = defaultTypeMeta(resource, gvk)
	b, _ := yaml.Marshal(resource)
	fmt.Fprintf(stdout, "---\n%s", b)
}

func defaultTypeMeta(resource runtime.Object, gvk schema.GroupVersionKind) runtime.Object {
	apiVersion, kind := gvk.ToAPIVersionAndKind()
	tm := metav1.TypeMeta{
		APIVersion: apiVersion,
		Kind:       kind,
	}
	reflect.ValueOf(resource).Elem().FieldByName("TypeMeta").Set(reflect.ValueOf(tm))
	return resource
}

type stdoutKey struct{}

func WithStdout(ctx context.Context, stdout io.Writer) context.Context {
	return context.WithValue(ctx, stdoutKey{}, stdout)
}

func StdoutFromContext(ctx context.Context) io.Writer {
	stdout, _ := ctx.Value(stdoutKey{}).(io.Writer)
	return stdout
}
