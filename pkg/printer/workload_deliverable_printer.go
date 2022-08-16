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
	"fmt"
	"io"
	"strings"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func DeliverableResourcesPrinter(w io.Writer, deliverable *cartov1alpha1.Deliverable) error {
	printResourceInfoRow := func(resource *cartov1alpha1.RealizedResource, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		var healthy string
		healthyCond := printer.FindCondition(resource.Conditions, cartov1alpha1.ConditionResourceHealthy)
		if healthyCond != nil {
			healthy = printer.ColorConditionStatus(string(healthyCond.Status))
		}

		ready, elapsedTransitionTime := findConditionReady(resource.Conditions, cartov1alpha1.ConditionResourceReady)
		row := metav1beta1.TableRow{
			Cells: []interface{}{
				resource.Name,
				ready,
				healthy,
				elapsedTransitionTime,
				getOutputRef(resource),
			},
		}
		return []metav1beta1.TableRow{row}, nil
	}

	printResourceInfoList := func(deliverable *cartov1alpha1.Deliverable, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		resourcesList := &deliverable.Status.Resources
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

	tablePrinter := table.NewTablePrinter(table.PrintOptions{PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Resource", Type: "string"},
			{Name: "Ready", Type: "string"},
			{Name: "Healthy", Type: "string"},
			{Name: "Time", Type: "string"},
			{Name: "Output", Type: "string"},
		}
		h.TableHandler(columns, printResourceInfoList)
		h.TableHandler(columns, printResourceInfoRow)
	})

	return tablePrinter.PrintObj(deliverable, w)
}

func DeliveryInfoPrinter(w io.Writer, deliverable *cartov1alpha1.Deliverable) error {
	printSupplyDeliveryInfo := func(deliverable *cartov1alpha1.Deliverable, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		deliveryRef := &deliverable.Status.DeliveryRef
		rows := []metav1beta1.TableRow{}
		if deliveryRef.Name != "" {
			nameRow := metav1beta1.TableRow{
				Cells: []interface{}{"name:", deliveryRef.Name},
			}
			rows = append(rows, nameRow)
		}
		return rows, nil

	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true, PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printSupplyDeliveryInfo)
	})

	return tablePrinter.PrintObj(deliverable, w)
}

func DeliverableIssuesPrinter(w io.Writer, deliverable *cartov1alpha1.Deliverable) error {
	readyCondition := printer.FindCondition(deliverable.Status.Conditions, cartov1alpha1.ConditionReady)
	healthyCondition := printer.FindCondition(deliverable.Status.Conditions, cartov1alpha1.ResourcesHealthy)
	if readyCondition == nil {
		return nil
	}
	printIssues := func(deliverable *cartov1alpha1.Deliverable, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		readyRow := metav1beta1.TableRow{
			Cells: []interface{}{
				fmt.Sprintf("%s [%s]:", cartov1alpha1.DeliverableKind, readyCondition.Reason),
				readyCondition.Message,
			},
		}
		rows := []metav1beta1.TableRow{readyRow}

		if healthyCondition != nil && healthyCondition.Message != "" {
			if strings.Compare(healthyCondition.Message, readyCondition.Message) != 0 {
				healthyRow := metav1beta1.TableRow{
					Cells: []interface{}{
						fmt.Sprintf("%s [%s]:", cartov1alpha1.DeliverableKind, healthyCondition.Reason),
						healthyCondition.Message,
					},
				}
				rows = append(rows, healthyRow)
			}
		}
		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true, PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printIssues)
	})

	return tablePrinter.PrintObj(deliverable, w)
}
