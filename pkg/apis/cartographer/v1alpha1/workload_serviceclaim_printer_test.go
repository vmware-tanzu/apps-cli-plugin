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
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWorkloadServiceClaimPrinterTablePrinter(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"

	tests := []struct {
		name           string
		testWorkload   *Workload
		expectedOutput string
	}{{
		name: "multiple services",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: WorkloadSpec{
				ServiceClaims: []WorkloadServiceClaim{{
					Name: "database",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "PostgreSQL",
						Name:       "my-prod-db",
					},
				}, {
					Name: "cache",
					Ref: &WorkloadServiceClaimReference{
						APIVersion: "services.tanzu.vmware.com/v1alpha1",
						Kind:       "Gemfire",
						Name:       "my-prod-cache",
					},
				}},
			},
		},
		expectedOutput: `
   CLAIM      NAME            KIND         API VERSION
   database   my-prod-db      PostgreSQL   services.tanzu.vmware.com/v1alpha1
   cache      my-prod-cache   Gemfire      services.tanzu.vmware.com/v1alpha1
`,
	}, {
		name: "no service",
		testWorkload: &Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: defaultNamespace,
			},
			Spec: WorkloadSpec{},
		},
		expectedOutput: `
   CLAIM   NAME   KIND   API VERSION
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := WorkloadServiceClaimPrinter(output, test.testWorkload); err != nil {
				t.Errorf("WorkloadServiceClaimPrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
