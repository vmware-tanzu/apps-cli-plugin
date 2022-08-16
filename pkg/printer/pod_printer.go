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
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/utils/integer"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

const TerminatingPhase = "Terminating"

type podview struct {
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
func PodTablePrintery(c *cli.Config, info *resource.Info) error {
	tableResult, podlist, _ := decodeIntoTable(info.Object)
	printPodRow := func(poddtls *podview, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		row := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: nil},
		}
		row.Cells = append(row.Cells,
			poddtls.name,
			poddtls.ready,
			poddtls.status,
			poddtls.restarts,
			poddtls.age,
		)
		return []metav1beta1.TableRow{row}, nil
	}
	printPodList := func(info *resource.Info, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		rows := make([]metav1beta1.TableRow, 0, len(podlist))
		for i := range podlist {
			r, err := printPodRow(&podlist[i], printOpts)
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
			{Name: "Ready", Type: "string"},
			{Name: "Status", Type: "string"},
			{Name: "Restarts", Type: "string"},
			{Name: "Age", Type: "string"},
		}
		h.TableHandler(columns, printPodList)
		h.TableHandler(columns, printPodRow)
	})

	return tablePrinter.PrintObj(tableResult, c.Stdout)
}

func maxContainerRestarts(status corev1.PodStatus) int {
	maxRestarts := 0
	for _, c := range status.ContainerStatuses {
		maxRestarts = integer.IntMax(maxRestarts, int(c.RestartCount))
	}
	return maxRestarts
}

// below print function are not being used and should be removed before final version.
func PodTablePrinterx(c *cli.Config, podList *corev1.PodList, info *resource.Info) {
	var printer printers.ResourcePrinter
	var mapping *meta.RESTMapping
	mapping = info.Mapping
	printer, _ = toPrinter(mapping, nil, false, false)
	printer.PrintObj(info.Object, c.Stdout)
}

func decodeIntoTable(obj runtime.Object) (runtime.Object, []podview, error) {
	event, isEvent := obj.(*metav1.WatchEvent)
	if isEvent {
		obj = event.Object.Object
	}
	var poddtls []podview
	if !recognizedTableVersions[obj.GetObjectKind().GroupVersionKind()] {
		return nil, poddtls, fmt.Errorf("attempt to decode non-Table object")
	}

	unstr, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, poddtls, fmt.Errorf("attempt to decode non-Unstructured object")
	}
	table := &metav1.Table{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, table); err != nil {
		return nil, poddtls, err
	}
	poddtls = make([]podview, len(table.Rows))
	for i := range table.Rows {
		row := &table.Rows[i]
		if row.Object.Raw == nil || row.Object.Object != nil {
			continue
		}
		var converted runtime.Object
		var err error
		if len(row.Cells) != 0 {
			poddtls[i].name = fmt.Sprintf("%v", row.Cells[0])
			poddtls[i].ready = fmt.Sprintf("%v", row.Cells[1])
			poddtls[i].status = fmt.Sprintf("%v", row.Cells[2])
			poddtls[i].restarts = fmt.Sprintf("%v", row.Cells[3])
			poddtls[i].age = fmt.Sprintf("%v", row.Cells[4])
			converted, err = runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
			if err != nil {
				return nil, poddtls, err
			}
		}
		row.Object.Object = converted
	}

	return table, poddtls, nil
}

// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through

var recognizedTableVersions = map[schema.GroupVersionKind]bool{
	metav1beta1.SchemeGroupVersion.WithKind("Table"): true,
	metav1.SchemeGroupVersion.WithKind("Table"):      true,
}

type TablePrinter struct {
	Delegate printers.ResourcePrinter
}

func (t *TablePrinter) PrintObj(obj runtime.Object, writer io.Writer) error {
	table, _, err := decodeIntoTable(obj)
	if err == nil {
		return t.Delegate.PrintObj(table, writer)
	}
	// if we are unable to decode server response into a v1beta1.Table,
	// fallback to client-side printing with whatever info the server returned.
	klog.V(2).Infof("Unable to decode server response into a Table. Falling back to hardcoded types: %v", err)
	return t.Delegate.PrintObj(obj, writer)
}

func toPrinter(mapping *meta.RESTMapping, outputObjects *bool, withNamespace bool, withKind bool) (printers.ResourcePrinterFunc, error) {
	// make a new copy of current flags / opts before mutating

	printer, err := ToPrinter()
	if err != nil {
		return nil, err
	}
	printer, err = printers.NewTypeSetter(scheme.Scheme).WrapToPrinter(printer, nil)
	if err != nil {
		return nil, err
	}
	if true {
		printer = &TablePrinter{Delegate: printer}
	}
	return printer.PrintObj, nil
}

// ToPrinter attempts to find a composed set of PrintFlags suitable for
// returning a printer based on current flag values.
func ToPrinter() (printers.ResourcePrinter, error) {
	p := printers.NewTablePrinter(printers.PrintOptions{
		Kind:          schema.GroupKind{Kind: "Pod"},
		WithKind:      false,
		NoHeaders:     false,
		Wide:          false,
		WithNamespace: false,
		ColumnLabels:  []string{},
		ShowLabels:    false,
	})

	// TODO(juanvallejo): handle sorting here

	return p, nil
}
