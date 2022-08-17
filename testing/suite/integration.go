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
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
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
	objOpts = []cmp.Option{
		cmpopts.IgnoreFields(cartov1alpha1.Workload{},
			"Status",
			"GenerateName",
			"SelfLink",
			"UID",
			"ResourceVersion",
			"CreationTimestamp",
			"DeletionTimestamp",
			"DeletionGracePeriodSeconds",
			"Finalizers",
			"ZZZ_DeprecatedClusterName",
			"ManagedFields",
		),
	}
)

const TestingNamespace = "apps-integration-testing"

type CommandLineIntegrationTestSuite []CommandLineIntegrationTestCase

type CommandLineIntegrationTestCase struct {
	Name                      string
	Skip                      bool
	Focus                     bool
	WorkloadName              string
	Command                   CommandLine
	ShouldError               bool
	ExpectedCommandLineOutput string
	ExpectedObject            client.Object
	IsList                    bool
	ExpectedObjectList        client.ObjectList
	Verify                    func(t *testing.T, output string, err error)
}

func (ts CommandLineIntegrationTestSuite) Run(t *testing.T) {
	ctx := context.Background()
	tearDown(t, t.Failed())
	setup(t)
	defer tearDown(t, t.Failed())

	restConfig, err := getRestConfig()
	if err != nil {
		t.Errorf("unexpected error fetching REST config: %v", err)
		t.FailNow()
	}
	dclient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		t.Errorf("unexpected error creating client: %v", err)
		t.FailNow()
	}

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
		// TODO: @shashwathi: do not stop running tests if one fails.
		if t.Failed() {
			break
		}
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
			gotData, err := appsclient.ListUsingGVK(ctx, cl.ExpectedObjectList.GetObjectKind().GroupVersionKind())
			if err != nil {
				t.Fatalf("Failed to get obj: %v", err)
			}

			var got client.ObjectList
			switch cl.ExpectedObjectList.(type) {
			case *cartov1alpha1.WorkloadList:
				got = &cartov1alpha1.WorkloadList{}
			case *cartov1alpha1.ClusterSupplyChainList:
				got = &cartov1alpha1.ClusterSupplyChainList{}
			default:
				t.Errorf("unexpected type: %v", reflect.TypeOf(cl.ExpectedObjectList))
				t.FailNow()
			}

			err = json.Unmarshal(gotData, got)
			if err != nil {
				t.Fatalf("Failed to get obj: %v", err)
			}
			if diff := cmp.Diff(cl.ExpectedObjectList, got, objOpts...); diff != "" {
				t.Errorf("%s(Object)\n(-expected, +actual)\n%s", cl.Name, diff)
				t.FailNow()
			}
		}

		if cl.ExpectedObject != nil {
			gotData, err := appsclient.GetUsingGVK(ctx, cl.ExpectedObject.GetObjectKind().GroupVersionKind(), cl.ExpectedObject.GetName())
			if err != nil {
				t.Fatalf("Failed to get obj: %v", err)
			}

			var got client.Object
			switch cl.ExpectedObject.(type) {
			case *cartov1alpha1.Workload:
				got = &cartov1alpha1.Workload{}
			case *cartov1alpha1.ClusterSupplyChain:
				got = &cartov1alpha1.ClusterSupplyChain{}
			default:
				t.Errorf("unexpected type: %v", reflect.TypeOf(cl.ExpectedObject))
				t.FailNow()
			}

			err = json.Unmarshal(gotData, got)
			if err != nil {
				t.Fatalf("Failed to get obj: %v", err)
			}
			if diff := cmp.Diff(cl.ExpectedObject, got, objOpts...); diff != "" {
				t.Errorf("%s(Object)\n(-expected, +actual)\n%s", cl.Name, diff)
				t.FailNow()
			}
		}

		if cl.Verify != nil {
			cl.Verify(t, cl.Command.GetOutput(), err)
		}
	})
}

func setup(t *testing.T) {
	// create namespace
	createNsCmd := NewKubectlCommandLine("create", "namespace", TestingNamespace)
	err := createNsCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createNsCmd.GetOutput(), createNsCmd.GetError())
		t.FailNow()
	}

	// create cluster supply chain
	createClusterCmd := NewKubectlCommandLine("apply", "-f", "testdata/prereq/cluster-supply-chain.yaml")
	err = createClusterCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createClusterCmd.GetOutput(), createClusterCmd.GetError())
		tearDown(t, true)
		t.FailNow()
	}
}

func tearDown(t *testing.T, fail bool) {
	// delete namespace
	NewKubectlCommandLine("delete", "namespace", TestingNamespace).Exec()

	// delete cluster supply chain
	NewKubectlCommandLine("delete", "-f", "testdata/prereq/cluster-supply-chain.yaml").Exec()
	if fail {
		t.FailNow()
	}
}

// getRestConfig returns REST config, which can be to use to create specific clientset
func getRestConfig() (*rest.Config, error) {
	var err error

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	// TODO (@shash) provide a param or flag to override kubeconfig path
	//loadingRules.ExplicitPath =
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
	if err != nil {
		return nil, err
	}

	return config.ClientConfig()
}

// TODO (@shash): Add DumpResourceInfo func to get describe on all available resources in test namespace before teardown
// func DumpResourceInfo(namespace string) {
// }
