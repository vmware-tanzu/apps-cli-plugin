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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/utils/integer"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

const TerminatingPhase = "Terminating"

type Podview struct {
	name     string
	ready    string
	status   string
	restarts string
	age      string
}

func PodTablePrinter(c *cli.Config, podList *corev1.PodList) error {
	printPodRow := func(pod *corev1.Pod, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		now := time.Now()
		row := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: pod},
		}
		phase := pod.Status.Phase
		if pod.DeletionTimestamp != nil {
			phase = TerminatingPhase
		}
		row.Cells = append(row.Cells,
			pod.Name,
			phase,
			maxContainerRestarts(pod.Status),
			printer.TimestampSince(pod.CreationTimestamp, now),
		)
		return []metav1beta1.TableRow{row}, nil
	}
	printPodList := func(pods *corev1.PodList, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		rows := make([]metav1beta1.TableRow, 0, len(pods.Items))
		for i := range pods.Items {
			r, err := printPodRow(&pods.Items[i], printOpts)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
		}
		return rows, nil
	}
	tablePrinter := table.NewTablePrinter(table.PrintOptions{PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Name", Type: "string"},
			{Name: "Status", Type: "string"},
			{Name: "Restarts", Type: "string"},
			{Name: "Age", Type: "string"},
		}
		h.TableHandler(columns, printPodList)
		h.TableHandler(columns, printPodRow)
	})
	return tablePrinter.PrintObj(podList, c.Stdout)
}

func maxContainerRestarts(status corev1.PodStatus) int {
	maxRestarts := 0
	for _, c := range status.ContainerStatuses {
		maxRestarts = integer.IntMax(maxRestarts, int(c.RestartCount))
	}
	return maxRestarts
}
func PodTablePrinterFromObject(c *cli.Config, tableResult runtime.Object) error {
	printPodList := func(printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		return nil, nil
	}
	tablePrinter := table.NewTablePrinter(table.PrintOptions{PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Name", Type: "string"},
			{Name: "Ready", Type: "string"},
			{Name: "Status", Type: "string"},
			{Name: "Restarts", Type: "string"},
			{Name: "Age", Type: "string"},
		}
		h.TableHandler(columns, printPodList)
	})

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
			return nil, fmt.Errorf("no pod details available")
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
	return nil, fmt.Errorf("no pod details available")
}

// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through

var recognizedTableVersions = map[schema.GroupVersionKind]bool{
	metav1beta1.SchemeGroupVersion.WithKind("Table"): true,
	metav1.SchemeGroupVersion.WithKind("Table"):      true,
}
