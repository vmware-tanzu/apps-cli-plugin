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
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func KnativeServicePrinter(c *cli.Config, kserviceList *knativeservingv1.ServiceList) error {
	printKnativeServiceRow := func(ksvc *knativeservingv1.Service, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		row := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: ksvc},
		}
		row.Cells = append(row.Cells,
			ksvc.Name,
		)
		row.Cells = append(row.Cells,
			printer.ConditionStatus(printer.FindCondition(ksvc.Status.Conditions, knativeservingv1.ServiceConditionReady)),
		)
		row.Cells = append(row.Cells, printer.EmptyString(ksvc.Status.URL))
		return []metav1beta1.TableRow{row}, nil
	}
	printKnativeServiceList := func(kserviceList *knativeservingv1.ServiceList, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		rows := make([]metav1beta1.TableRow, 0, len(kserviceList.Items))
		for i := range kserviceList.Items {
			r, err := printKnativeServiceRow(&kserviceList.Items[i], printOpts)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
		}
		return rows, nil
	}
	tablePrinter := table.NewTablePrinter(table.PrintOptions{}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Name", Type: "string"},
			{Name: "Ready", Type: "string"},
			{Name: "URL", Type: "string"},
		}
		h.TableHandler(columns, printKnativeServiceList)
		h.TableHandler(columns, printKnativeServiceRow)
	})
	return tablePrinter.PrintObj(kserviceList, c.Stdout)
}
