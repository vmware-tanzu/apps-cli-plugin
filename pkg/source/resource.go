/*
Copyright 2022 VMware, Inc.

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

package source

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
)

var recognizedTableVersions = map[schema.GroupVersionKind]bool{
	metav1beta1.SchemeGroupVersion.WithKind("Table"): true,
	metav1.SchemeGroupVersion.WithKind("Table"):      true,
}

func FetchResourceObjects(builder *resource.Builder, namespace string, labelSelectorParam string, types []string) (runtime.Object, error) {
	r := builder.Unstructured().
		NamespaceParam(namespace).
		LabelSelectorParam(labelSelectorParam).
		ResourceTypeOrNameArgs(true, types...).
		Latest().
		Flatten().
		TransformRequests(func(req *rest.Request) {
			req.SetHeader("Accept", strings.Join([]string{
				fmt.Sprintf(runtime.ContentTypeJSON+";as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
				runtime.ContentTypeJSON,
			}, ","))
		}).
		Do()
	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}
	if len(infos) >= 1 {
		return decodeIntoObject(infos[0])
	}
	return nil, nil
}

func decodeIntoObject(info *resource.Info) (runtime.Object, error) {
	obj := info.Object
	event, isEvent := obj.(*metav1.WatchEvent)
	if isEvent {
		obj = event.Object.Object
	}
	if !recognizedTableVersions[obj.GetObjectKind().GroupVersionKind()] {
		return nil, fmt.Errorf("attempt to decode non-Table object")
	}
	unstr, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("attempt to decode non-Unstructured object")
	}
	table := &metav1.Table{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, table); err != nil {
		return nil, err
	}
	if len(table.Rows) == 0 {
		return nil, nil
	}
	for i := range table.Rows {
		row := &table.Rows[i]
		if row.Object.Raw == nil || row.Object.Object != nil {
			continue
		}
		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
		if err != nil {
			return nil, err
		}
		row.Object.Object = converted
	}
	return table, nil
}
