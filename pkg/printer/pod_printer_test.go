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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

func TestPodTablePrinter(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	testConfig := cli.NewDefaultConfig("test", scheme)
	defaultNamespace := "default"
	podName := "my-pod"

	tests := []struct {
		name           string
		testPodList    []client.Object
		expectedOutput string
	}{{
		name: "running status",
		testPodList: []client.Object{diecorev1.PodBlank.
			MetadataDie(func(d *diemetav1.ObjectMetaDie) {
				d.Name(podName)
				d.Namespace(defaultNamespace)
			}).
			StatusDie(func(d *diecorev1.PodStatusDie) {
				d.Phase(corev1.PodRunning)
			})},
		expectedOutput: `
   NAME     READY   STATUS   RESTARTS   AGE
   my-pod   0/0              0          <unknown>
`,
	}, {
		name: "failed status",
		testPodList: []client.Object{diecorev1.PodBlank.
			MetadataDie(func(d *diemetav1.ObjectMetaDie) {
				d.Name(podName)
				d.Namespace(defaultNamespace)
			}).
			StatusDie(func(d *diecorev1.PodStatusDie) {
				d.Phase(corev1.PodFailed)
			})},
		expectedOutput: `
   NAME     READY   STATUS   RESTARTS   AGE
   my-pod   0/0              0          <unknown>
`,
	}, {
		name: "Unknown  pod",
		testPodList: []client.Object{diecorev1.PodBlank.
			MetadataDie(func(d *diemetav1.ObjectMetaDie) {
				d.Name(podName)
				d.Namespace(defaultNamespace)
			}).
			StatusDie(func(d *diecorev1.PodStatusDie) {
				d.Phase(corev1.PodUnknown)
			})},
		expectedOutput: `
   NAME     READY   STATUS   RESTARTS   AGE
   my-pod   0/0              0          <unknown>
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			testConfig.Stdout = output
			podObject := clitesting.TableMetaObject(test.testPodList)
			if err := printer.PodTablePrinter(testConfig, podObject); err != nil {
				t.Errorf("PodTablePrinter() expected no error, got %v", err)
			}
			outputString := output.String()
			if diff := cmp.Diff(strings.TrimPrefix(test.expectedOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
