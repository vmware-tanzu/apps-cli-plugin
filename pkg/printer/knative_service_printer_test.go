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

package printer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestKnativeServiceTablePrinter(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = knativeservingv1.AddToScheme(scheme)
	testConfig := cli.NewDefaultConfig("test", scheme)
	output := &bytes.Buffer{}
	testConfig.Stdout = output
	defaultNamespace := "default"
	url := "https://example.com"

	testknativeServiceList := &knativeservingv1.ServiceList{
		Items: []knativeservingv1.Service{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ksvc",
				Namespace: defaultNamespace,
			},
			Status: knativeservingv1.ServiceStatus{
				Conditions: []metav1.Condition{{
					Status: metav1.ConditionFalse,
					Type:   knativeservingv1.ServiceConditionReady,
				}},
				URL: url,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ksvc-no-url",
				Namespace: defaultNamespace,
			},
			Status: knativeservingv1.ServiceStatus{
				Conditions: []metav1.Condition{{
					Status: metav1.ConditionUnknown,
					Type:   knativeservingv1.ServiceConditionReady,
				}},
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-ksvc-no-status",
				Namespace: defaultNamespace,
			},
		}},
	}

	if err := printer.KnativeServicePrinter(testConfig, testknativeServiceList); err != nil {
		t.Errorf("KnativeServicePrinter() expected no error, got %v", err)
	}

	outputString := output.String()
	expectedOutput := `
   NAME                READY       URL
   my-ksvc             not-Ready   https://example.com
   my-ksvc-no-url      Unknown     <empty>
   my-ksvc-no-status   <unknown>   <empty>
`
	if diff := cmp.Diff(strings.TrimPrefix(expectedOutput, "\n"), outputString); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
}
