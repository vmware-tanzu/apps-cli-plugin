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
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer/table"
)

const (
	SelectorType        = "labels"
	MatchExpressionType = "expressions"
	MatchFieldType      = "fields"
)

func ClusterSupplyChainPrinter(w io.Writer, clustersupplychain *cartov1alpha1.ClusterSupplyChain) error {
	printRow := func(typeStr, key, operator string, values ...string) metav1beta1.TableRow {
		row := metav1beta1.TableRow{
			Object: runtime.RawExtension{Object: clustersupplychain},
		}
		row.Cells = append(row.Cells,
			typeStr,
			key,
			operator,
		)
		if len(values) == 1 {
			row.Cells = append(row.Cells,
				values[0])
		}
		return row
	}
	printSelectorRow := func(selectors map[string]string, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		if len(selectors) == 0 {
			return nil, nil
		}
		keys := make([]string, 0, len(selectors))
		for k := range selectors {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		selectorRows := []metav1beta1.TableRow{}
		for _, k := range keys {
			selectorRows = append(selectorRows, printRow(SelectorType, k, "", selectors[k]))
		}
		return selectorRows, nil
	}
	printMatchExpressionsRows := func(expressions []metav1.LabelSelectorRequirement, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		if len(expressions) == 0 {
			return nil, nil
		}
		expressionRows := []metav1beta1.TableRow{}
		for i := range expressions {
			if len(expressions[i].Values) == 0 {
				expressionRows = append(expressionRows, printRow(MatchExpressionType, expressions[i].Key, string(expressions[i].Operator)))
			} else {
				for j := range expressions[i].Values {
					expressionRows = append(expressionRows, printRow(MatchExpressionType, expressions[i].Key, string(expressions[i].Operator), expressions[i].Values[j]))
				}
			}
		}
		return expressionRows, nil
	}
	printMatchFieldsRow := func(fields []cartov1alpha1.FieldSelectorRequirement, _ table.PrintOptions) ([]metav1beta1.TableRow, error) {
		if len(fields) == 0 {
			return nil, nil
		}
		expressionRows := []metav1beta1.TableRow{}
		for i := range fields {
			if len(fields[i].Values) == 0 {
				expressionRows = append(expressionRows, printRow(MatchFieldType, fields[i].Key, string(fields[i].Operator)))
			} else {
				for j := range fields[i].Values {
					expressionRows = append(expressionRows, printRow(MatchFieldType, fields[i].Key, string(fields[i].Operator), fields[i].Values[j]))
				}
			}
		}
		return expressionRows, nil
	}

	printClusterSupplyChainSelectors := func(clustersupplychain *cartov1alpha1.ClusterSupplyChain, printOpts table.PrintOptions) ([]metav1beta1.TableRow, error) {
		rows := []metav1beta1.TableRow{}
		selectorRows, err := printSelectorRow(clustersupplychain.Spec.Selector, printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, selectorRows...)
		fieldRows, err := printMatchFieldsRow(clustersupplychain.Spec.SelectorMatchFields, printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, fieldRows...)
		expressionRows, err := printMatchExpressionsRows(clustersupplychain.Spec.SelectorMatchExpressions, printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, expressionRows...)
		return rows, nil
	}

	tablePrinter := table.NewTablePrinter(table.PrintOptions{PaddingStart: paddingStart}).With(func(h table.PrintHandler) {
		columns := []metav1beta1.TableColumnDefinition{
			{Name: "Type", Type: "string"},
			{Name: "Key", Type: "string"},
			{Name: "Operator", Type: "string"},
			{Name: "Value", Type: "string"},
		}
		h.TableHandler(columns, printSelectorRow)
		h.TableHandler(columns, printMatchFieldsRow)
		h.TableHandler(columns, printMatchExpressionsRows)
		h.TableHandler(columns, printClusterSupplyChainSelectors)
	})
	return tablePrinter.PrintObj(clustersupplychain, w)
}
