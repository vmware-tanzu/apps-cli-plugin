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

package v1alpha1

import (
	"io"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func WorkloadServiceClaimPrinter(w io.Writer, workload *Workload) error {
	printSvcRow := func(serviceClaim *WorkloadServiceClaim, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		row := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: &Workload{}},
		}
		row.Cells = append(row.Cells,
			serviceClaim.Name,
			serviceClaim.Ref.Name,
			serviceClaim.Ref.Kind,
			serviceClaim.Ref.APIVersion,
		)
		return []metav1beta1.TableRow{row}, nil
	}

	printSvcList := func(workload *Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		serviceClaims := workload.Spec.ServiceClaims
		rows := make([]metav1beta1.TableRow, 0, len(serviceClaims))
		for i := range serviceClaims {
			r, err := printSvcRow(&serviceClaims[i], printOpts)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
		}
		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Claim", Type: "string"},
			{Name: "Name", Type: "string"},
			{Name: "Kind", Type: "string"},
			{Name: "API Version", Type: "string"},
		}
		h.TableHandler(columns, printSvcList)
		h.TableHandler(columns, printSvcRow)
	})
	return tablePrinter.PrintObj(workload, w)
}
