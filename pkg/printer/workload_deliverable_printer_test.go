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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestDeliverableResourcesPrinter(t *testing.T) {
	defaultNamespace := "default"
	deliverableName := "my-deliverable"

	tests := []struct {
		name            string
		testDeliverable *cartov1alpha1.Deliverable
		expectedOutput  string
	}{{
		name: "various resources",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
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
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionTrue,
						},
					},
				}, {
					Name: "deployer",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionUnknown,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionUnknown,
						},
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
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
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionFalse,
						},
					},
				}},
			},
		},
		expectedOutput: `
   RESOURCE          READY     HEALTHY   TIME        OUTPUT
   source-provider   True      True      <unknown>   not found
   deployer          Unknown   Unknown   <unknown>   not found
   image-builder     False     False     <unknown>   not found
`,
	}, {
		name: "no resources",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
		},
		expectedOutput: `
   RESOURCE   READY   HEALTHY   TIME   OUTPUT
`,
	}, {
		name: "no ready condition inside resource",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionTrue,
						},
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionTrue,
						},
					},
				}, {
					Name: "deployer",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionUnknown,
						},
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
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
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionFalse,
						},
					},
				}},
			},
		},
		expectedOutput: `
   RESOURCE          READY   HEALTHY   TIME        OUTPUT
   source-provider           True                  not found
   deployer                  Unknown               not found
   image-builder     False   False     <unknown>   not found
`,
	}, {
		name: "no healthy condition inside resource",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionTrue,
						},
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionTrue,
						},
					},
				}, {
					Name: "deployer",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionUnknown,
						},
						{
							Type:   cartov1alpha1.ConditionResourceReady,
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
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionFalse,
						},
					},
				}},
			},
		},
		expectedOutput: `
   RESOURCE          READY     HEALTHY   TIME        OUTPUT
   source-provider   True                <unknown>   not found
   deployer          Unknown             <unknown>   not found
   image-builder     False     False     <unknown>   not found
`,
	}, {
		name: "with output details",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionTrue,
						},
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionTrue,
						},
					},
					StampedRef: &corev1.ObjectReference{Kind: "GitRepository", Name: "pet-clinic"},
				}, {
					Name: "deployer",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionUnknown,
						},
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionUnknown,
						},
					},
					StampedRef: &corev1.ObjectReference{Kind: "App", Name: "pet-clinic"},
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
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionFalse,
						},
					},
					StampedRef: &corev1.ObjectReference{},
				}, {
					Name: "config-provider",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionFalse,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionFalse,
						},
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionFalse,
						},
					},
					StampedRef: &corev1.ObjectReference{Kind: "", Name: "pet-clinic"},
				}, {
					Name: "app-config",
					Conditions: []metav1.Condition{
						{
							Type:   cartov1alpha1.ConditionResourceReady,
							Status: metav1.ConditionFalse,
						},
						{
							Type:   cartov1alpha1.ConditionResourceSubmitted,
							Status: metav1.ConditionFalse,
						},
						{
							Type:   cartov1alpha1.ConditionResourceHealthy,
							Status: metav1.ConditionFalse,
						},
					},
					StampedRef: &corev1.ObjectReference{Kind: "ConfigMap", Name: ""},
				}},
			},
		},
		expectedOutput: `
   RESOURCE          READY     HEALTHY   TIME        OUTPUT
   source-provider   True                <unknown>   GitRepository/pet-clinic
   deployer          Unknown             <unknown>   App/pet-clinic
   image-builder     False     False     <unknown>   not found
   config-provider   False     False     <unknown>   /pet-clinic
   app-config        False     False     <unknown>   ConfigMap/
`,
	}, {
		name: "resource without conditions",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				Resources: []cartov1alpha1.RealizedResource{{
					Name: "source-provider",
				}, {
					Name: "deployer",
				}},
			},
		},
		expectedOutput: `
   RESOURCE          READY   HEALTHY   TIME   OUTPUT
   source-provider                            not found
   deployer                                   not found
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.DeliverableResourcesPrinter(output, test.testDeliverable); err != nil {
				t.Errorf("DeliverableSourcePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestDeliveryInfoPrinter(t *testing.T) {
	defaultNamespace := "default"
	deliverableName := "my-deliverable"

	tests := []struct {
		name            string
		testDeliverable *cartov1alpha1.Deliverable
		expectedOutput  string
	}{{
		name: "various conditions",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				DeliveryRef: cartov1alpha1.ObjectReference{
					Name: "my-delivery",
				},
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:   cartov1alpha1.SupplyChainReady,
						Status: metav1.ConditionTrue,
					}, {
						Type:   cartov1alpha1.ConditionReady,
						Status: metav1.ConditionTrue,
					}},
				},
			},
		},
		expectedOutput: `
   name:   my-delivery
`,
	}, {
		name: "no conditions",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				DeliveryRef: cartov1alpha1.ObjectReference{
					Name: "my-delivery",
				},
			},
		},
		expectedOutput: `
   name:   my-delivery
`,
	}, {
		name: "no delivery info",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:   cartov1alpha1.ConditionReady,
						Status: metav1.ConditionFalse,
					}, {
						Type:   cartov1alpha1.ConditionReady,
						Status: metav1.ConditionFalse,
					}},
				},
			},
		},
		expectedOutput: ``,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.DeliveryInfoPrinter(output, test.testDeliverable); err != nil {
				t.Errorf("DeliveryInfoPrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestDeliverablesIssuesPrinter(t *testing.T) {
	defaultNamespace := "default"
	deliverableName := "my-deliverable"

	tests := []struct {
		name            string
		testDeliverable *cartov1alpha1.Deliverable
		expectedOutput  string
	}{{
		name: "condition ready with info",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:    cartov1alpha1.ConditionReady,
						Status:  metav1.ConditionFalse,
						Message: "a hopefully informative message",
						Reason:  "OopsieDoodle",
					}},
				},
			},
		},
		expectedOutput: `
   Deliverable [OopsieDoodle]:   a hopefully informative message
`,
	}, {
		name: "ready and healthy with same info",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:    cartov1alpha1.ConditionReady,
						Status:  metav1.ConditionFalse,
						Message: "a hopefully informative message",
						Reason:  "OopsieDoodle",
					}, {
						Type:    cartov1alpha1.ResourcesHealthy,
						Status:  metav1.ConditionFalse,
						Message: "a hopefully informative message",
						Reason:  "OopsieDoodle",
					}},
				},
			},
		},
		expectedOutput: `
   Deliverable [OopsieDoodle]:   a hopefully informative message
`,
	}, {
		name: "ready and healthy with different info",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:    cartov1alpha1.ConditionReady,
						Status:  metav1.ConditionFalse,
						Message: "a hopefully informative message",
						Reason:  "OopsieDoodle",
					}, {
						Type:    cartov1alpha1.ResourcesHealthy,
						Status:  metav1.ConditionFalse,
						Message: "a hopefully informative message for non-healthy deliverable",
						Reason:  "AnotherOopsieDoodle",
					}},
				},
			},
		},
		expectedOutput: `
   Deliverable [OopsieDoodle]:          a hopefully informative message
   Deliverable [AnotherOopsieDoodle]:   a hopefully informative message for non-healthy deliverable
`,
	}, {
		name: "condition ready with no info",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:   cartov1alpha1.ConditionReady,
						Status: metav1.ConditionFalse,
						Reason: "OopsieDoodle",
					}},
				},
			},
		},
		expectedOutput: `
   Deliverable [OopsieDoodle]:   
`,
	}, {
		name: "condition ready and health with no message",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:   cartov1alpha1.ConditionReady,
						Status: metav1.ConditionFalse,
						Reason: "OopsieDoodle",
					}, {
						Type:   cartov1alpha1.ResourcesHealthy,
						Status: metav1.ConditionUnknown,
						Reason: "AnotherOopsieDoodle",
					}},
				},
			},
		},
		expectedOutput: `
   Deliverable [OopsieDoodle]:   
`,
	}, {
		name: "no status",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
		},
	}, {
		name: "no ready condition",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{},
				},
			},
		},
	}, {
		name: "no ready condition but Health",
		testDeliverable: &cartov1alpha1.Deliverable{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deliverableName,
				Namespace: defaultNamespace,
			},
			Status: cartov1alpha1.DeliverableStatus{
				OwnerStatus: cartov1alpha1.OwnerStatus{
					Conditions: []metav1.Condition{{
						Type:    cartov1alpha1.ConditionReady,
						Status:  metav1.ConditionUnknown,
						Message: "a hopefully informative message",
						Reason:  "OopsieDoodle",
					}, {
						Type:    cartov1alpha1.ResourcesHealthy,
						Status:  metav1.ConditionFalse,
						Message: "a hopefully informative message for non-healthy deliverable",
						Reason:  "AnotherOopsieDoodle",
					}},
				},
			},
		},
		expectedOutput: `
   Deliverable [OopsieDoodle]:          a hopefully informative message
   Deliverable [AnotherOopsieDoodle]:   a hopefully informative message for non-healthy deliverable
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := printer.DeliverableIssuesPrinter(output, test.testDeliverable); err != nil {
				t.Errorf("DeliverablePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
