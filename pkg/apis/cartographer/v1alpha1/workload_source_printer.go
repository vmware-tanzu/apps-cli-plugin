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
	"strings"
	"time"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

type resourcesQuery struct {
	resourceName  string
	outputName    string
	imageHeader   string
	updatedHeader string
	rows          []metav1beta1.TableRow
}

func findResource(rq resourcesQuery, resources []RealizedResource) []metav1beta1.TableRow {
	imageRow := metav1beta1.TableRow{}
	updatedRow := metav1beta1.TableRow{}

	for _, r := range resources {
		for _, o := range r.Outputs {
			if r.Name == rq.resourceName && o.Name == rq.outputName {
				imageRow.Cells = append(imageRow.Cells,
					rq.imageHeader,
					strings.Replace(o.Preview, "\n", "", -1),
				)
				updatedRow.Cells = append(updatedRow.Cells,
					rq.updatedHeader,
					printer.TimestampSince(o.LastTransitionTime, time.Now()),
				)
				rq.rows = append(rq.rows, imageRow, updatedRow)
				break
			}
		}
	}

	return rq.rows
}

func WorkloadSourceImagePrinter(w io.Writer, workload *Workload) error {
	printImageInfo := func(workload *Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		sourceRow := metav1beta1.TableRow{}
		sourceRow.Cells = append(sourceRow.Cells,
			"Source:",
			"Image",
		)

		rows := []metav1beta1.TableRow{sourceRow}

		rq := resourcesQuery{
			resourceName:  "image-provider",
			outputName:    "image",
			imageHeader:   "Image:",
			updatedHeader: "Updated:",
			rows:          rows,
		}

		rows = findResource(rq, workload.Status.Resources)

		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printImageInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}

func WorkloadSourceLocalImagePrinter(w io.Writer, workload *Workload) error {
	printLocalSourceInfo := func(workload *Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		sourceRow := metav1beta1.TableRow{}
		sourceRow.Cells = append(sourceRow.Cells,
			"Source:",
			"Local source code",
		)
		rows := []metav1beta1.TableRow{sourceRow}

		if workload.Spec.Source.Subpath != "" {
			subPathRow := metav1beta1.TableRow{}
			subPathRow.Cells = append(subPathRow.Cells,
				"Sub-path:",
				workload.Spec.Source.Subpath,
			)
			rows = append(rows, subPathRow)
		}

		rq := resourcesQuery{
			resourceName:  "image-builder",
			outputName:    "image",
			imageHeader:   "Source-image:",
			updatedHeader: "Updated:",
			rows:          rows,
		}

		rows = findResource(rq, workload.Status.Resources)

		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printLocalSourceInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}

func WorkloadSourceGitPrinter(w io.Writer, workload *Workload) error {
	printGitInfo := func(workload *Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		sourceRow := metav1beta1.TableRow{}
		sourceRow.Cells = append(sourceRow.Cells,
			"Source:",
			workload.Spec.Source.Git.URL,
		)
		rows := []metav1beta1.TableRow{sourceRow}

		if workload.Spec.Source.Subpath != "" {
			subPathRow := metav1beta1.TableRow{}
			subPathRow.Cells = append(subPathRow.Cells,
				"Sub-path:",
				workload.Spec.Source.Subpath,
			)
			rows = append(rows, subPathRow)
		}

		if workload.Spec.Source.Git.Ref.Branch != "" {
			branchRow := metav1beta1.TableRow{}
			branchRow.Cells = append(branchRow.Cells,
				"Branch:",
				workload.Spec.Source.Git.Ref.Branch,
			)
			rows = append(rows, branchRow)
		}

		if workload.Spec.Source.Git.Ref.Tag != "" {
			tagRow := metav1beta1.TableRow{}
			tagRow.Cells = append(tagRow.Cells,
				"Tag:",
				workload.Spec.Source.Git.Ref.Tag,
			)
			rows = append(rows, tagRow)
		}

		if workload.Spec.Source.Git.Ref.Commit != "" {
			commitRow := metav1beta1.TableRow{}
			commitRow.Cells = append(commitRow.Cells,
				"Commit:",
				workload.Spec.Source.Git.Ref.Commit,
			)
			rows = append(rows, commitRow)
		}

		rq := resourcesQuery{
			resourceName:  "image-builder",
			outputName:    "image",
			imageHeader:   "Built image:",
			updatedHeader: "Updated:",
			rows:          rows,
		}

		rows = findResource(rq, workload.Status.Resources)

		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printGitInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}
