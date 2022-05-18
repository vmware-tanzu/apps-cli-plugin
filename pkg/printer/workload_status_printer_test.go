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

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestWorkloadResourcesPrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *cartov1alpha1.Workload
		expectedOutput string
	}{{
		name: "various resources",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionTrue,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionTrue,
						},
					},
				}, {
					Name: "deliverable",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionUnknown,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionUnknown,
						},
					},
				}, {
					Name: "image-builder",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionFalse,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionFalse,
						},
					},
				}},
			},
		},
		expectedOutput: `
RESOURCE          READY     TIME
source-provider   True      <unknown>
deliverable       Unknown   <unknown>
image-builder     False     <unknown>
`,
	}, {
		name: "no resources",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
		},
		expectedOutput: `
RESOURCE   READY   TIME
`,
	}, {
		name: "no ready condition inside resource",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionTrue,
						},
					},
				}, {
					Name: "deliverable",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionUnknown,
						},
					},
				}, {
					Name: "image-builder",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionFalse,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionFalse,
						},
					},
				}},
			},
		},
		expectedOutput: `
RESOURCE          READY   TIME
source-provider           
deliverable               
image-builder     False   <unknown>
`,
	}, {
		name: "resource without conditions",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
				}, {
					Name: "deliverable",
				}},
			},
		},
		expectedOutput: `
RESOURCE          READY   TIME
source-provider           
deliverable               
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.WorkloadResourcesPrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestWorkloadSupplyChainInfoPrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *cartov1alpha1.Workload
		expectedOutput string
	}{{
		name: "various conditions",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				SupplyChainRef: cartov1alpha1.ObjectReference{
					Name: "my-supply-chain",
				},
				Conditions: []metav1.Condition{{
					Type:   cartov1alpha1.SupplyChainReady,
					Status: metav1.ConditionTrue,
				}, {
					Type:   cartov1alpha1.WorkloadResourceSubmitted,
					Status: metav1.ConditionTrue,
				}, {
					Type:   cartov1alpha1.WorkloadReady,
					Status: metav1.ConditionTrue,
				}},
			},
		},
		expectedOutput: `
name:          my-supply-chain
last update:   <unknown>
ready:         True
`,
	}, {
		name: "no conditions",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				SupplyChainRef: cartov1alpha1.ObjectReference{
					Name: "my-supply-chain",
				},
			},
		},
		expectedOutput: `
name:          my-supply-chain
last update:   
ready:         
`,
	}, {
		name: "no ready condition in status",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				SupplyChainRef: cartov1alpha1.ObjectReference{
					Name: "my-supply-chain",
				},
				Conditions: []metav1.Condition{{
					Type:   cartov1alpha1.WorkloadSupplyChainReady,
					Status: metav1.ConditionFalse,
				}, {
					Type:   cartov1alpha1.WorkloadResourceSubmitted,
					Status: metav1.ConditionFalse,
				}},
			},
		},
		expectedOutput: `
name:          my-supply-chain
last update:   
ready:         
`,
	}, {
		name: "no supply chain info",
		testWorkload: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.WorkloadStatus{
				Conditions: []metav1.Condition{{
					Type:   cartov1alpha1.WorkloadSupplyChainReady,
					Status: metav1.ConditionFalse,
				}, {
					Type:   cartov1alpha1.WorkloadResourceSubmitted,
					Status: metav1.ConditionFalse,
				}, {
					Type:   cartov1alpha1.WorkloadReady,
					Status: metav1.ConditionFalse,
				}},
			},
		},
		expectedOutput: `
name:          <none>
last update:   <unknown>
ready:         False
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.WorkloadSupplyChainInfoPrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
