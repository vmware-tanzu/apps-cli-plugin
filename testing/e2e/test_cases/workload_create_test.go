//go:build integration
// +build integration

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

package integration_test

import (
	"testing"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	it "github.com/vmware-tanzu/apps-cli-plugin/testing/e2e/suite"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateFromGitWithAnnotations(t *testing.T) {
	testSuite := it.CommandLineIntegrationTestSuite{
		{
			Name:         "Create workload with valid name from a git repo and annotations",
			WorkloadName: "test-create-git-annotations-workload",
			Command: func() it.CommandLine {
				c := *it.NewTanzuAppsCommandLine(
					"workload", "create", "test-create-git-annotations-workload",
					"--annotation=min-instances=2", "--annotation=max-instances=4",
					"--app=test-create-git-annotations-workload",
					"--git-tag=tap-1.2",
					"--git-repo=https://github.com/sample-accelerators/spring-petclinic",
					namespaceFlag,
					"--type=web",
					"--yes",
				)
				return c
			}(),
			ExpectedRemoteObject: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-git-annotations-workload",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of":           "test-create-git-annotations-workload",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						{
							Name: "annotations",
							Value: v1.JSON{
								Raw: []byte(`{"max-instances":"4","min-instances":"2"}`),
							},
						},
					},
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: "https://github.com/sample-accelerators/spring-petclinic",
							Ref: cartov1alpha1.GitRef{
								Tag: "tap-1.2",
							},
						},
					},
				},
			},
		},
		{
			Name:         "Update the created workload",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "apply", "test-create-git-annotations-workload", namespaceFlag, "--annotation=min-instances=3", "--annotation=max-instances=5", "-y"),
			ExpectedRemoteObject: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-git-annotations-workload",
					Namespace: it.TestingNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/part-of":           "test-create-git-annotations-workload",
						"apps.tanzu.vmware.com/workload-type": "web",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Params: []cartov1alpha1.Param{
						{
							Name: "annotations",
							Value: v1.JSON{
								Raw: []byte(`{"max-instances":"5","min-instances":"3"}`),
							},
						},
					},
					Source: &cartov1alpha1.Source{
						Git: &cartov1alpha1.GitSource{
							URL: "https://github.com/sample-accelerators/spring-petclinic",
							Ref: cartov1alpha1.GitRef{
								Tag: "tap-1.2",
							},
						},
					},
				},
			},
		},
		{
			Name:         "Delete the created workload",
			WorkloadName: "test-create-git-annotations-workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "delete", "test-create-git-annotations-workload", namespaceFlag, "-y"),
			ExpectedRemoteObject: emptyWorkload,
		},
		{
			Name: "Get the deleted workload",
			Command: *it.NewTanzuAppsCommandLine(
				"workload", "get", "test-create-git-annotations-workload", namespaceFlag),
			ExpectedCommandLineError: true,
			ExpectedRemoteObject:     emptyWorkload,
		},
	}
	testSuite.Run(t)
}
