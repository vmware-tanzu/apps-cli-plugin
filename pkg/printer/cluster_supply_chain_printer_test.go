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

package printer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestClusterSupplyChainPrinter(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = cartov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name           string
		supplychain    *cartov1alpha1.ClusterSupplyChain
		expectedOutput string
	}{{
		name: "labels",
		supplychain: &cartov1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cs-label",
			},
			Spec: cartov1alpha1.SupplyChainSpec{
				LegacySelector: cartov1alpha1.Selector{
					Selector: map[string]string{
						"apps.tanzu.vmware.com/workload-type":               "web",
						"apps.tanzu.vmware.com/workload-deployment-cluster": "test",
					},
				},
			},
		},
		expectedOutput: `
TYPE     KEY                                                 OPERATOR   VALUE
labels   apps.tanzu.vmware.com/workload-deployment-cluster              test
labels   apps.tanzu.vmware.com/workload-type                            web
`,
	}, {
		name: "expressions",
		supplychain: &cartov1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cs-label",
			},
			Spec: cartov1alpha1.SupplyChainSpec{
				LegacySelector: cartov1alpha1.Selector{
					SelectorMatchExpressions: []metav1.LabelSelectorRequirement{{
						Key:      "app",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"web", "svc"},
					}, {
						Key:      "foo",
						Operator: metav1.LabelSelectorOpNotIn,
					}},
				},
			},
		},
		expectedOutput: `
TYPE          KEY   OPERATOR   VALUE
expressions   app   In         web
expressions   app   In         svc
expressions   foo   NotIn
`,
	}, {
		name: "fields",
		supplychain: &cartov1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cs-label",
			},
			Spec: cartov1alpha1.SupplyChainSpec{
				LegacySelector: cartov1alpha1.Selector{
					SelectorMatchFields: []cartov1alpha1.FieldSelectorRequirement{{
						Key:      "spec.image",
						Operator: cartov1alpha1.FieldSelectorOperator("Exists"),
					}, {
						Key:      "resource.cpulimits",
						Operator: cartov1alpha1.FieldSelectorOperator("Exists"),
					}},
				},
			},
		},
		expectedOutput: `
TYPE     KEY                  OPERATOR   VALUE
fields   spec.image           Exists
fields   resource.cpulimits   Exists
`}, {
		// this is very unlikey scenario
		name: "fields with values",
		supplychain: &cartov1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cs-label",
			},
			Spec: cartov1alpha1.SupplyChainSpec{
				LegacySelector: cartov1alpha1.Selector{
					SelectorMatchFields: []cartov1alpha1.FieldSelectorRequirement{{
						Key:      "spec.image",
						Operator: cartov1alpha1.FieldSelectorOperator("Contains"),
						Values:   []string{"docker.io", "gcr.io"},
					}, {
						Key:      "resource.cpulimits",
						Operator: cartov1alpha1.FieldSelectorOperator("Exists"),
					}},
				},
			},
		},
		expectedOutput: `
TYPE     KEY                  OPERATOR   VALUE
fields   spec.image           Contains   docker.io
fields   spec.image           Contains   gcr.io
fields   resource.cpulimits   Exists
`}, {
		name: "all label selectors present",
		supplychain: &cartov1alpha1.ClusterSupplyChain{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cs-label",
			},
			Spec: cartov1alpha1.SupplyChainSpec{
				LegacySelector: cartov1alpha1.Selector{
					Selector: map[string]string{
						"apps.tanzu.vmware.com/workload-type": "web",
					},
					SelectorMatchFields: []cartov1alpha1.FieldSelectorRequirement{{
						Key:      "spec.source.image",
						Operator: cartov1alpha1.FieldSelectorOperator("Exists"),
					}},
					SelectorMatchExpressions: []metav1.LabelSelectorRequirement{{
						Key:      "app",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"web", "source-to-image"},
					}},
				},
			},
		},
		expectedOutput: `
TYPE          KEY                                   OPERATOR   VALUE
labels        apps.tanzu.vmware.com/workload-type              web
fields        spec.source.image                     Exists
expressions   app                                   In         web
expressions   app                                   In         source-to-image
`,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.ClusterSupplyChainPrinter(output, test.supplychain); err != nil {
				t.Errorf("ClusterSupplyChainPrinter() expected no error, got %v", err)
			}

			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
