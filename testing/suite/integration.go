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

package suite_test

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	dfOpts = []cmp.Option{
		cmp.Transformer("T1", func(in string) string {
			s := strings.TrimSpace(in)
			s = strings.Replace(in, "  ", " ", -1)
			return regexp.MustCompile(`( (\d+[smhd]){1,4})[ \n$]`).ReplaceAllString(s, "0s")
		}),
		// ^([0-9]{1,10}[dhms]{1}){1,4}$
	}

	ignoreFields = []string{
		"CreationTimestamp",
		"DeletionTimestamp",
		"DeletionGracePeriodSeconds",
		"Finalizers",
		"Generation",
		"ManagedFields",
		"OwnerReferences",
		"ResourceVersion",
		"Status",
		"SelfLink",
		"UID",
		"ZZZ_DeprecatedClusterName",
	}

	containsignoreFields = func(str string) bool {
		for _, v := range ignoreFields {
			if strings.ToLower(v) == strings.ToLower(str) {
				return true
			}
		}
		return false
	}

	objOpts = []cmp.Option{
		cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
			if containsignoreFields(k) {
				return true
			}
			return false
		}),
	}
)

const TestingNamespace = "apps-integration-testing"

type CommandLineIntegrationTestSuite []CommandLineIntegrationTestCase

type CommandLineIntegrationTestCase struct {
	Name                      string
	Skip                      bool
	Focus                     bool
	RequireEnvs               []string
	WorkloadName              string
	Command                   CommandLine
	ShouldError               bool
	ExpectedCommandLineOutput string
	ExpectedObject            client.Object
	IsList                    bool
	ExpectedObjectList        client.ObjectList
	Verify                    func(t *testing.T, output string, err error)
	Prepare                   func(t *testing.T) error
	CleanUp                   func(t *testing.T) error
}

func (ts CommandLineIntegrationTestSuite) Run(t *testing.T) {
	ctx := context.Background()
	tearDown(t)
	defer func() {
		if t.Failed() {
			dumpResourceInfo(t, TestingNamespace)
		}
		tearDown(t)
	}()
	err := setup(t)
	assert.NilError(t, err)

	restConfig, err := getRestConfig()
	assert.NilError(t, err)

	dclient, err := dynamic.NewForConfig(restConfig)
	assert.NilError(t, err)

	testToRun := ts
	focused := CommandLineIntegrationTestSuite{}
	for _, tc := range ts {
		if tc.Focus && !tc.Skip {
			focused = append(focused, tc)
		}
	}
	isFocused := len(focused) != 0
	if isFocused {
		testToRun = focused
	}

	appsClient := NewDynamicClient(dclient, TestingNamespace)
	for _, tc := range testToRun {
		tc.Run(t, ctx, appsClient)
	}
	if isFocused {
		t.Errorf("test run focused on %d record(s), skipped %d record(s)", len(focused), len(ts)-len(focused))
	}

}

func (cl CommandLineIntegrationTestCase) Run(t *testing.T, ctx context.Context, appsclient DynamicClient) {
	t.Run(cl.Name, func(t *testing.T) {
		if cl.Skip {
			t.SkipNow()
		}
		for _, e := range cl.RequireEnvs {
			if _, ok := os.LookupEnv(e); !ok {
				t.Skipf("Required %q environment variable not present, skipping test", e)
			}
		}
		if cl.Prepare != nil {
			if err := cl.Prepare(t); err != nil {
				t.Errorf("unexpected error in prepare: %v", err)
				t.FailNow()
			}

		}
		err := cl.Command.Exec()
		// t.Logf("Command output:\n%s\n\n%v\n", cl.Command.String(), cl.Command.GetOutput())
		if err != nil && !cl.ShouldError {
			t.Errorf("unexpected error in exec: %v: %v, ", err, cl.Command.GetOutput())
			t.FailNow()
		}

		if err == nil && cl.ShouldError {
			t.Errorf("Expected error, got nil")
			t.FailNow()
		}

		if cl.ExpectedCommandLineOutput != "" {
			if diff := cmp.Diff(cl.ExpectedCommandLineOutput, cl.Command.GetOutput(), dfOpts...); diff != "" {
				t.Errorf("%s\n(-expected, +actual)\n%s", cl.Name, diff)
				t.FailNow()
			}
		}

		if cl.ExpectedObjectList != nil {
			e, err := json.Marshal(cl.ExpectedObjectList)
			if err != nil {
				t.Fatalf("unexpected error when marshall obj: %v: %v", cl.ExpectedObjectList, err)
			}
			expectedObjectList := &unstructured.UnstructuredList{}
			if _, _, err = unstructured.UnstructuredJSONScheme.Decode(e, nil, expectedObjectList); err != nil {
				t.Fatalf("unexpected error decode obj: %v: %v", cl.ExpectedObjectList, err)
			}

			got, err := appsclient.ListUsingGVK(ctx, cl.ExpectedObjectList.GetObjectKind().GroupVersionKind())
			if err != nil {
				t.Fatalf("unexpected error on get obj GVK %v, %v", cl.ExpectedObjectList.GetObjectKind().GroupVersionKind(), err)
			}

			if diff := cmp.Diff(expectedObjectList, got, objOpts...); diff != "" {
				t.Errorf("%s(ObjectList)\n(-expected, +actual)\n%s", cl.Name, diff)
				t.FailNow()
			}
		}

		if cl.ExpectedObject != nil {
			e, err := json.Marshal(cl.ExpectedObject)
			if err != nil {
				t.Fatalf("unexpected error when marshall obj: %v: %v", cl.ExpectedObject, err)
			}
			expectedOBj := &unstructured.Unstructured{}
			if _, _, err = unstructured.UnstructuredJSONScheme.Decode(e, nil, expectedOBj); err != nil {
				t.Fatalf("unexpected error decode obj: %v: %v", cl.ExpectedObject, err)
			}

			got, err := appsclient.GetUsingGVK(ctx, cl.ExpectedObject.GetObjectKind().GroupVersionKind(), cl.ExpectedObject.GetName())
			if err != nil {
				t.Fatalf("unexpected error on get obj GVK %v, name %s : %v", cl.ExpectedObject.GetObjectKind().GroupVersionKind(), cl.ExpectedObject.GetName(), err)
			}
			if diff := cmp.Diff(expectedOBj, got, objOpts...); diff != "" {
				t.Errorf("%s(Object)\n(-expected, +actual)\n%s", cl.Name, diff)
				t.FailNow()
			}
		}

		if cl.Verify != nil {
			cl.Verify(t, cl.Command.GetOutput(), err)
		}
		if cl.CleanUp != nil {
			if err := cl.CleanUp(t); err != nil {
				t.Errorf("unexpected error in clean up: %v", err)
				t.FailNow()
			}
		}
	})
}

func setup(t *testing.T) error {
	// create namespace
	createNsCmd := NewKubectlCommandLine("create", "namespace", TestingNamespace)
	err := createNsCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createNsCmd.GetOutput(), createNsCmd.GetError())
		return err
	}

	// create cluster supply chain
	createClusterCmd := NewKubectlCommandLine("apply", "-f", "testdata/prereq/cluster-supply-chain.yaml")
	err = createClusterCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createClusterCmd.GetOutput(), createClusterCmd.GetError())
		return err
	}
	// create pod
	createPodCmd := NewKubectlCommandLine("apply", "-f", "testdata/prereq/pod-withlabel.yaml")
	err = createPodCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createPodCmd.GetOutput(), createPodCmd.GetError())
		return err
	}
	return nil
}

func tearDown(t *testing.T) {
	// delete pod
	NewKubectlCommandLine("delete", "--all", "pods", "-n", TestingNamespace).Exec()

	// delete namespace
	NewKubectlCommandLine("delete", "namespace", TestingNamespace).Exec()

	// delete cluster supply chain
	NewKubectlCommandLine("delete", "-f", "testdata/prereq/cluster-supply-chain.yaml").Exec()
}

// getRestConfig returns REST config, which can be to use to create specific clientset
func getRestConfig() (*rest.Config, error) {
	var err error

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
	if err != nil {
		return nil, err
	}

	return config.ClientConfig()
}

// dumpResourceInfo func to get describe on relevant resource in namespace and cluster
func dumpResourceInfo(t *testing.T, namespace string) {
	t.Log("Dump resources...")
	// describe workloads namespace
	descrWorkload := NewKubectlCommandLine("describe", "workloads", "-n", TestingNamespace)
	descrWorkload.Exec()
	t.Logf("Describe workloads \n %s", descrWorkload.GetOutput())
	// describe clustersupplychains
	descrCSC := NewKubectlCommandLine("describe", "clustersupplychains")
	descrCSC.Exec()
	t.Logf("Describe clustersupplychains \n %s", descrCSC.GetOutput())
	// describe pod
	descrPod := NewKubectlCommandLine("describe", "pods", "-n", TestingNamespace)
	descrPod.Exec()
	t.Logf("Describe pods \n %s", descrPod.GetOutput())
}
