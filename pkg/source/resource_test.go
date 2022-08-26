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

package source_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
	k8sscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/google/go-cmp/cmp"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

func TestWorkloadOptionsFetchResourceObjects(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	pod1Die := diecorev1.PodBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("pod1")
			d.Namespace(defaultNamespace)
			d.AddLabel(cartov1alpha1.WorkloadLabelName, workloadName)
		}).Kind("pod")

	scheme := runtime.NewScheme()
	fakeClient := clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))
	table := &metav1.Table{}
	tests := []struct {
		name        string
		args        []string
		shouldError bool
		expected    []metav1.TableRow
		builder     *resource.Builder
	}{{
		name:        "Fetch Resource Object successfully",
		args:        []string{"pods"},
		shouldError: false,
		builder: resource.NewFakeBuilder(
			func(version schema.GroupVersion) (resource.RESTClient, error) {
				codec := k8sscheme.Codecs.LegacyCodec(scheme.PrioritizedVersionsAllGroups()...)
				UnstructuredClient := &fake.RESTClient{
					NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
					Resp:                 &http.Response{StatusCode: http.StatusOK, Header: clitesting.DefaultHeader(), Body: clitesting.PodV1TableObjBody(codec, []client.Object{pod1Die})},
				}
				return UnstructuredClient, nil
			},
			fakeClient.ToRESTMapper,
			func() (restmapper.CategoryExpander, error) {
				return resource.FakeCategoryExpander, nil
			},
		),
		expected: append(table.Rows, metav1.TableRow{
			Cells: []interface{}{"pod1", "0/0", "", int64(0), "<unknown>", "<none>", "<none>", "<none>", "<none>"},
		})}, {
		name:        "Fetch Empty Resource Object successfully",
		args:        []string{"pods"},
		shouldError: false,
		builder: resource.NewFakeBuilder(
			func(version schema.GroupVersion) (resource.RESTClient, error) {
				codec := k8sscheme.Codecs.LegacyCodec(scheme.PrioritizedVersionsAllGroups()...)
				UnstructuredClient := &fake.RESTClient{
					NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
					Resp:                 &http.Response{StatusCode: http.StatusOK, Header: clitesting.DefaultHeader(), Body: clitesting.PodV1TableObjBody(codec, []client.Object{})},
				}
				return UnstructuredClient, nil
			},
			fakeClient.ToRESTMapper,
			func() (restmapper.CategoryExpander, error) {
				return resource.FakeCategoryExpander, nil
			},
		),
	}, {
		name:        "Fetch  Resource Object Error",
		args:        []string{"pods"},
		shouldError: true,
		builder: resource.NewFakeBuilder(
			func(version schema.GroupVersion) (resource.RESTClient, error) {
				UnstructuredClient := &fake.RESTClient{
					NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
					Resp:                 &http.Response{StatusCode: http.StatusPreconditionFailed, Header: clitesting.DefaultHeader(), Body: nil},
				}
				return UnstructuredClient, nil
			},
			fakeClient.ToRESTMapper,
			func() (restmapper.CategoryExpander, error) {
				return resource.FakeCategoryExpander, nil
			},
		),
	}, {
		name:        "Fetch  Resource  Error",
		args:        []string{"pods"},
		shouldError: true,
		builder: resource.NewFakeBuilder(
			func(version schema.GroupVersion) (resource.RESTClient, error) {
				codec := k8sscheme.Codecs.LegacyCodec(scheme.PrioritizedVersionsAllGroups()...)
				UnstructuredClient := &fake.RESTClient{
					NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
					Resp:                 &http.Response{StatusCode: http.StatusOK, Header: clitesting.DefaultHeader(), Body: podV1TableErrorObjBody(codec, []client.Object{pod1Die})},
				}
				return UnstructuredClient, nil
			},
			fakeClient.ToRESTMapper,
			func() (restmapper.CategoryExpander, error) {
				return resource.FakeCategoryExpander, nil
			},
		),
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			labelparam := fmt.Sprintf("%s%s%s", cartov1alpha1.WorkloadLabelName, "=", workloadName)
			obj, err := source.FetchResourceObjects(test.builder, defaultNamespace, labelparam, test.args)
			if obj != nil {
				gotTable := obj.(*metav1.Table)
				if d := cmp.Diff(gotTable.Rows, test.expected); d != "" {
					t.Errorf("Diff() %s", d)
				}
			}
			if err != nil && !test.shouldError {
				t.Errorf("FetchResourceObjects() errored %v", err)
			}
			if err == nil && test.shouldError {
				t.Errorf("FetchResourceObjects() expected error %v, got nil", test.shouldError)
			}
			if test.shouldError {
				return
			}
		})
	}
}

// build a meta table response from a pod list
func podV1TableErrorObjBody(codec runtime.Codec, pods []client.Object) io.ReadCloser {
	table := &metav1.Table{
		TypeMeta:          metav1.TypeMeta{APIVersion: "meta.k8s.io/v1", Kind: "pod"},
		ColumnDefinitions: []metav1.TableColumnDefinition{},
	}
	for _, obj := range pods {
		b := bytes.NewBuffer(nil)
		codec.Encode(obj, b)
		table.Rows = append(table.Rows, metav1.TableRow{})
	}

	data, err := json.Marshal(table)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(string(data), `"meta.k8s.io/v1"`) {
		panic("expected v1, got " + string(data))
	}
	return ioutil.NopCloser(bytes.NewReader(data))
}
