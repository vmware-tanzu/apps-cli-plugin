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

package printer

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func PodTablePrinter(c *cli.Config, tableResult runtime.Object) error {
	tablePrinter := table.NewTablePrinter(table.PrintOptions{PaddingStart: paddingStart})
	return tablePrinter.PrintObj(tableResult, c.Stdout)
}

func DecodeIntoTable(info []*resource.Info) (runtime.Object, error) {
	if len(info) == 1 {
		obj := info[0].Object
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
			return nil, fmt.Errorf("no pod details found")
		}
		for i := range table.Rows {
			row := &table.Rows[i]
			if row.Object.Raw == nil || row.Object.Object != nil {
				continue
			}
			var converted runtime.Object
			var err error

			converted, err = runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
			if err != nil {
				return nil, err
			}

			row.Object.Object = converted
		}

		return table, nil
	}
	return nil, fmt.Errorf("no pod details found")
}

// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through

var recognizedTableVersions = map[schema.GroupVersionKind]bool{
	metav1beta1.SchemeGroupVersion.WithKind("Table"): true,
	metav1.SchemeGroupVersion.WithKind("Table"):      true,
}
