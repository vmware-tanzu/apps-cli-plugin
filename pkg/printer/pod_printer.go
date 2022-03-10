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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/integer"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

const TerminatingPhase = "Terminating"

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
	tablePrinter := table.NewTablePrinter(table.PrintOptions{}).With(func(h table.PrintHandler) {
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
