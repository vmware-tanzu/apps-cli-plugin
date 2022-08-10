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
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			"Generation",
			"CreationTimestamp",
			"DeletionTimestamp",
			"DeletionGracePeriodSeconds",
			"OwnerReferences",
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
	ExpectedRemoteList        client.ObjectList
	Verify                    func(t *testing.T, output string, err error)
}

func (ts CommandLineIntegrationTestSuite) Run(t *testing.T) {

	cleanUp(t, false)
	prepareEnvironment(t)
	defer cleanUp(t, t.Failed())

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

	conf, err := NewClientConfig()
	if err != nil {
		t.Errorf("Error creating validation config: %v", err)
		t.FailNow()
	}

	for _, tc := range testToRun {
		tc.Run(t, conf)
		if t.Failed() {
			break
		}
	}
	if isFocused {
		t.Errorf("test run focused on %d record(s), skipped %d record(s)", len(focused), len(ts)-len(focused))
	}

}

func (cl CommandLineIntegrationTestCase) Run(t *testing.T, conf *ClientConfig) {
	t.Run(cl.Name, func(t *testing.T) {

		if cl.Skip {
			t.SkipNow()
		}
		err := cl.Command.Exec()
		// t.Logf("Command output:\n%s\n\n%v\n", cl.Command.String(), cl.Command.GetOutput())
		if err != nil && !cl.ShouldError {
			t.Errorf("unexpected error: %v: %v, ", err, cl.Command.GetOutput())
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

		if cl.ExpectedRemoteList != nil {
			var got client.ObjectList
			switch cl.ExpectedRemoteList.(type) {
			case *cartov1alpha1.WorkloadList:
				got = &cartov1alpha1.WorkloadList{}
			case *cartov1alpha1.ClusterSupplyChainList:
				got = &cartov1alpha1.ClusterSupplyChainList{}
			default:
				t.Errorf("unexpected type: %v", reflect.TypeOf(cl.ExpectedRemoteList))
				t.FailNow()
			}
			if err := conf.List(got, client.InNamespace(TestingNamespace), client.MatchingLabels(map[string]string{})); err != nil {
				t.Errorf("unexpected error %s: %v", cl.Name, err)
				t.FailNow()
			}
			if diff := cmp.Diff(cl.ExpectedRemoteList, got); diff != "" {
				t.Errorf("%s(List)\n(-expected, +actual)\n%s", cl.Name, diff)
				t.FailNow()
			}
		}

		if cl.ExpectedObject != nil {
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
			if err := conf.Get(client.ObjectKey{Namespace: TestingNamespace, Name: cl.ExpectedObject.GetName()}, got); err != nil {
				t.Errorf("unexpected error %s: %v", cl.Name, err)
				t.FailNow()
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

func prepareEnvironment(t *testing.T) {
	// create namespace
	createNsCmd := NewCommandLine("kubectl", "create", "namespace", TestingNamespace)
	err := createNsCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createNsCmd.GetOutput(), createNsCmd.GetError())
		t.FailNow()
	}

	// create cluster supply chain
	createClusterCmd := NewCommandLine("kubectl", "apply", "-f", "testdata/prereq/cluster-supply-chain.yaml")
	err = createClusterCmd.Exec()
	if err != nil {
		t.Errorf("unexpected error: %v\n%s\n%s", err, createClusterCmd.GetOutput(), createClusterCmd.GetError())
		cleanUp(t, true)
		t.FailNow()
	}
}
func cleanUp(t *testing.T, fail bool) {
	deleteNsCmd := NewCommandLine("kubectl", "delete", "namespace", TestingNamespace)
	deleteNsCmd.Exec()

	deleteCSCCmd := NewCommandLine("kubectl", "delete", "-f", "testdata/prereq/cluster-supply-chain.yaml")
	deleteCSCCmd.Exec()
	if fail {
		t.FailNow()
	}
}
