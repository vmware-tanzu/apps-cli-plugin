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

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

func WorkloadSourceImagePrinter(w io.Writer, workload *cartov1alpha1.Workload) error {
	printImageInfo := func(workload *cartov1alpha1.Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		sourceRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"type:",
				"image",
			},
		}

		imageRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"image:",
				workload.Spec.Image,
			},
		}

		rows := []metav1beta1.TableRow{sourceRow, imageRow}
		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true, PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printImageInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}

func WorkloadLocalSourceImagePrinter(w io.Writer, workload *cartov1alpha1.Workload) error {
	printLocalSourceInfo := func(workload *cartov1alpha1.Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		sourceRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"type:",
				"source image",
			},
		}

		rows := []metav1beta1.TableRow{sourceRow}

		if workload.Spec.Source.Subpath != "" {
			subPathRow := metav1beta1.TableRow{
				Cells: []interface{}{
					"sub-path:",
					workload.Spec.Source.Subpath,
				},
			}
			rows = append(rows, subPathRow)
		}

		imageRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"image:",
				workload.Spec.Source.Image,
			},
		}
		rows = append(rows, imageRow)

		return rows, nil
	}
	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true, PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printLocalSourceInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}

func WorkloadSourceGitPrinter(w io.Writer, workload *cartov1alpha1.Workload) error {
	printGitInfo := func(workload *cartov1alpha1.Workload, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		sourceRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"type:",
				"git",
			},
		}

		urlRow := metav1beta1.TableRow{
			Cells: []interface{}{
				"url:",
				workload.Spec.Source.Git.URL,
			},
		}

		rows := []metav1beta1.TableRow{sourceRow, urlRow}

		if workload.Spec.Source.Subpath != "" {
			subPathRow := metav1beta1.TableRow{
				Cells: []interface{}{
					"sub-path:",
					workload.Spec.Source.Subpath,
				},
			}
			rows = append(rows, subPathRow)
		}

		if workload.Spec.Source.Git.Ref.Branch != "" {
			branchRow := metav1beta1.TableRow{
				Cells: []interface{}{
					"branch:",
					workload.Spec.Source.Git.Ref.Branch,
				},
			}
			rows = append(rows, branchRow)
		}

		if workload.Spec.Source.Git.Ref.Tag != "" {
			tagRow := metav1beta1.TableRow{
				Cells: []interface{}{
					"tag:",
					workload.Spec.Source.Git.Ref.Tag,
				},
			}
			rows = append(rows, tagRow)
		}

		if workload.Spec.Source.Git.Ref.Commit != "" {
			commitRow := metav1beta1.TableRow{
				Cells: []interface{}{
					"commit:",
					workload.Spec.Source.Git.Ref.Commit,
				},
			}
			rows = append(rows, commitRow)
		}

		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{NoHeaders: true, PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		h.TableHandler(nil, printGitInfo)
	})

	return tablePrinter.PrintObj(workload, w)
}
