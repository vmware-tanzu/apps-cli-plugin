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

package printer

import (
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func WorkloadResourcesPrinter(w io.Writer, workload *cartov1alpha1.Workload) error {
	printResourceInfoRow := func(resource *cartov1alpha1.RealizedResource, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		ready, elapsedTransitionTime := findConditionReady(resource.Conditions, cartov1alpha1.ConditionResourceReady)

		row := metav1beta1.TableRow{
			Cells: []interface{}{
				resource.Name,
				ready,
				elapsedTransitionTime,
			},
		}
		return []metav1beta1.TableRow{row}, nil
	}

	printResourceInfoList := func(workload *cartov1alpha1.Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		resourcesList := &workload.Status.Resources
		rows := make([]metav1beta1.TableRow, 0, len(*resourcesList))
		for _, r := range *resourcesList {
			row, err := printResourceInfoRow(&r, printOpts)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row...)
		}
		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Resource", Type: "string"},
			{Name: "Ready", Type: "string"},
			{Name: "Time", Type: "string"},
		}
		h.TableHandler(columns, printResourceInfoList)
		h.TableHandler(columns, printResourceInfoRow)
	})
	return tablePrinter.PrintObj(workload, w)
}

func WorkloadSupplyChainInfoPrinter(w io.Writer, workload *cartov1alpha1.Workload) error {
	printSupplyChainInfo := func(workload *cartov1alpha1.Workload, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		workloadStatus := &workload.Status

		ready, elapsedTransitionTime := findConditionReady(workloadStatus.Conditions, cartov1alpha1.WorkloadConditionReady)

		var name string
		if workloadStatus.SupplyChainRef.Name != "" {
			name = workloadStatus.SupplyChainRef.Name
		} else {
			name = "<none>"
		}

		nameRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"name:",
				name,
			},
		}

		elapsedTimeRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"last update:",
				elapsedTransitionTime,
			},
		}

		readyRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"ready:",
				ready,
			},
		}

		rows := []metav1beta1.TableRow{nameRow, elapsedTimeRow, readyRow}
		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printSupplyChainInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}

func findConditionReady(conditions []metav1.Condition, strReadyCondition string) (string, string) {
	var ready string
	var elapsedTransitionTime string

	conditionReady := printer.FindCondition(conditions, strReadyCondition)

	if conditionReady != nil {
		ready = string(conditionReady.Status)
		elapsedTransitionTime = printer.TimestampSince(conditionReady.LastTransitionTime, time.Now())
	}

	return ready, elapsedTransitionTime
}
