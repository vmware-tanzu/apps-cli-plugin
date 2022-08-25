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

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func WorkloadOverviewPrinter(w io.Writer, workload *cartov1alpha1.Workload) error {
	printLocalSourceInfo := func(workload *cartov1alpha1.Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		labels := workload.Labels
		if labels == nil {
			labels = map[string]string{}
		}
		nameRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"name:",
				workload.GetName(),
			},
		}
		sourceRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"type:",
				printer.EmptyString(labels[apis.WorkloadTypeLabelName]),
			},
		}

		rows := []metav1beta1.TableRow{nameRow, sourceRow}

		return rows, nil
	}
	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true, PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printLocalSourceInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}
