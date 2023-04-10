/*
Copyright 2019 VMware, Inc.

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

package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	fakerest "k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
	"k8s.io/kubectl/pkg/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func (c *fakeclient) DefaultNamespace() string {
	return "default"
}

func (c *fakeclient) KubeRestConfig() *rest.Config {
	if c.kubeConfig != nil {
		return c.kubeConfig
	}
	return &rest.Config{}
}

func (c *fakeclient) Discovery() discovery.DiscoveryInterface {
	return discovery.NewDiscoveryClientForConfigOrDie(&rest.Config{})
}

func (c *fakeclient) SetLogger(logger logr.Logger) {
	panic(fmt.Errorf("not implemented"))
}

func (c *fakeclient) GetClientSet() kubernetes.Interface {
	return newClientSet()
}

func NewFakeCliClient(c crclient.Client) cli.Client {
	return &fakeclient{
		defaultNamespace: "default",
		Client:           c,
	}
}

func NewFakeCliClientWithTransport(c crclient.Client, transport http.RoundTripper) cli.Client {
	var t http.RoundTripper
	if transport != nil {
		t = transport
	} else {
		t = fakeTransport{}
	}
	return &fakeclient{
		defaultNamespace: "default",
		Client:           c,
		kubeConfig: &rest.Config{
			Transport: t,
		},
	}
}

// RESTClient returns a REST client from TestFactory
func (c *fakeclient) RESTClient() (*rest.RESTClient, error) {
	panic(fmt.Errorf("not implemented"))
}

type fakeTransport struct {
	corev1Response *http.Response
}

func (t fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.corev1Response, nil
}

func NewFakeTransportFromResponse(resp *http.Response) http.RoundTripper {
	return fakeTransport{corev1Response: resp}
}

func (c *fakeclient) ToRESTMapper() (meta.RESTMapper, error) {
	return testRESTMapper(), nil
}

func (c *fakeclient) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return disk.NewCachedDiscoveryClientForConfig(&rest.Config{}, "", "", 0) // need alignment with sash
}

func (c *fakeclient) ToRESTConfig() (*rest.Config, error) {
	return c.KubeRestConfig(), nil
}

func NewFakeCachedDiscoveryClient() *FakeCachedDiscoveryClient {
	return &FakeCachedDiscoveryClient{
		Groups:             []*metav1.APIGroup{},
		Resources:          []*metav1.APIResourceList{},
		PreferredResources: []*metav1.APIResourceList{},
		Invalidations:      0,
	}
}
func (d *FakeCachedDiscoveryClient) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return d.Groups, d.Resources, nil
}

// NewBuilder returns a new resource builder for structured api objects.
func (c *fakeclient) NewBuilder() *resource.Builder {
	panic(fmt.Errorf("not implemented"))
}

type FakeCachedDiscoveryClient struct {
	discovery.DiscoveryInterface
	Groups             []*metav1.APIGroup
	Resources          []*metav1.APIResourceList
	PreferredResources []*metav1.APIResourceList
	Invalidations      int
}

type fakeclient struct {
	defaultNamespace string
	crclient.Client
	kubeConfig *rest.Config
}

func newClientSet() *fakeClientSet {
	return &fakeClientSet{Clientset: kubernetes.Clientset{}}
}

type fakeClientSet struct {
	kubernetes.Clientset
}

func (c *fakeClientSet) CoreV1() corev1.CoreV1Interface {
	return &fakeClientCoreV1{CoreV1Client: corev1.CoreV1Client{}}
}

type fakeClientCoreV1 struct {
	corev1.CoreV1Client
}

func (c *fakeClientCoreV1) RESTClient() rest.Interface {
	return &fakeRestClient{RESTClient: fakerest.RESTClient{}}
}

type fakeRestClient struct {
	fakerest.RESTClient
}

func testRESTMapper() meta.RESTMapper {
	groupResources := testDynamicResources()
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	// for backwards compatibility with existing tests, allow rest mappings from the scheme to show up
	// TODO: make this opt-in?
	mapper = meta.FirstHitRESTMapper{
		MultiRESTMapper: meta.MultiRESTMapper{
			mapper,
			testrestmapper.TestOnlyStaticRESTMapper(scheme.Scheme),
		},
	}

	fakeDs := NewFakeCachedDiscoveryClient()
	expander := restmapper.NewShortcutExpander(mapper, fakeDs)
	return expander
}

func testDynamicResources() []*restmapper.APIGroupResources {
	return []*restmapper.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "pods", Namespaced: true, Kind: "Pod"},
					{Name: "services", Namespaced: true, Kind: "Service"},
					{Name: "namespacedtype", Namespaced: true, Kind: "NamespacedType"},
					{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
				},
			},
		},
	}
}

// build a readercloser response from a pod list
func PodV1TableObjBody(codec runtime.Codec, pods []crclient.Object) io.ReadCloser {
	table := TableMetaObject(pods)
	data, err := json.Marshal(table)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(string(data), `"meta.k8s.io/v1"`) {
		panic("expected v1, got " + string(data))
	}
	return ioutil.NopCloser(bytes.NewReader(data))
}

func DefaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}

// build a meta table response from list of client objects
func TableMetaObject(objects []crclient.Object) *metav1.Table {
	var podColumns = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name"},
		{Name: "Ready", Type: "string", Format: ""},
		{Name: "Status", Type: "string", Format: ""},
		{Name: "Restarts", Type: "integer", Format: ""},
		{Name: "Age", Type: "string", Format: ""},
		{Name: "IP", Type: "string", Format: "", Priority: 1},
		{Name: "Node", Type: "string", Format: "", Priority: 1},
		{Name: "Nominated Node", Type: "string", Format: "", Priority: 1},
		{Name: "Readiness Gates", Type: "string", Format: "", Priority: 1},
	}
	table := &metav1.Table{
		TypeMeta:          metav1.TypeMeta{APIVersion: "meta.k8s.io/v1", Kind: "Table"},
		ColumnDefinitions: podColumns,
	}
	for i := range objects {
		b := bytes.NewBuffer(nil)
		table.Rows = append(table.Rows, metav1.TableRow{
			Object: runtime.RawExtension{Raw: b.Bytes()},
			Cells:  []interface{}{objects[i].GetName(), "0/0", "", int64(0), "<unknown>", "<none>", "<none>", "<none>", "<none>"},
		})
	}
	return table
}
