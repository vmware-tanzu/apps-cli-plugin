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
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestPodTablePrinter(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	testConfig := cli.NewDefaultConfig("test", scheme)
	defaultNamespace := "default"
	podName := "my-pod"
	deletiontime := metav1.NewTime(time.Now())

	tests := []struct {
		name           string
		testPodList    *corev1.PodList
		expectedOutput string
	}{{
		name: "running status",
		testPodList: &corev1.PodList{
			Items: []corev1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: defaultNamespace,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{
						Name:  "web",
						Ready: true,
					}, {
						Name:  "sidecar",
						Ready: true,
					}},
				},
			}},
		},
		expectedOutput: `
NAME     STATUS    RESTARTS   AGE
my-pod   Running   0          <unknown>
`,
	}, {
		name: "failed status",
		testPodList: &corev1.PodList{
			Items: []corev1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: defaultNamespace,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{{
						Name:         "web",
						Ready:        false,
						RestartCount: 1,
					}, {
						Name:  "sidecar",
						Ready: true,
					}},
				},
			}},
		},
		expectedOutput: `
NAME     STATUS   RESTARTS   AGE
my-pod   Failed   1          <unknown>
`,
	}, {
		name: "terminating pod",
		testPodList: &corev1.PodList{
			Items: []corev1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:              podName,
					Namespace:         defaultNamespace,
					DeletionTimestamp: &deletiontime,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}},
		},
		expectedOutput: `
NAME     STATUS        RESTARTS   AGE
my-pod   Terminating   0          <unknown>
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			testConfig.Stdout = output
			if err := printer.PodTablePrinter(testConfig, test.testPodList); err != nil {
				t.Errorf("PodTablePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
