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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func (c *fakeclient) DefaultNamespace() string {
	return "default"
}

// RESTClient returns a REST client from TestFactory
func (c *fakeclient) RESTClient() (*rest.RESTClient, error) {
	// Swap out the HTTP client out of the client with the fake's version.
	fakeClient := c.Clients.(*fake.RESTClient)
	restClient, err := rest.RESTClientFor(c.ClientConfigVal)
	if err != nil {
		panic(err)
	}
	restClient.Client = fakeClient.Client
	return restClient, nil
}

func (c *fakeclient) KubeRestConfig() *rest.Config {
	panic(fmt.Errorf("not implemented"))
	// below wll be removed once confirming after review

	// tmpFile, _ := ioutil.TempFile(os.TempDir(), "cmdtests_temp")
	// loadingRules := &clientcmd.ClientConfigLoadingRules{
	// 	Precedence:     []string{tmpFile.Name()},
	// 	MigrationRules: map[string]string{},
	// }

	// overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmdapi.Cluster{Server: "http://localhost:8080"}}
	// fallbackReader := bytes.NewBuffer([]byte{})
	// clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, fallbackReader)
	// restConfig, _ := clientConfig.ClientConfig()
	// return restConfig
}

func (c *fakeclient) Discovery() discovery.DiscoveryInterface {
	return discovery.NewDiscoveryClientForConfigOrDie(&rest.Config{})
}

func (c *fakeclient) SetLogger(logger logr.Logger) {
	panic(fmt.Errorf("not implemented"))
}

func (c *fakeclient) RESTMapper() meta.RESTMapper {
	return testRESTMapper()
}
func (c *fakeclient) ToRESTMapper() (meta.RESTMapper, error) {
	return c.RESTMapper(), nil
}
func (c *fakeclient) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return disk.NewCachedDiscoveryClientForConfig(&rest.Config{}, "", "", 0) // need alignment with sash
}
func (c *fakeclient) ToRESTConfig() (*rest.Config, error) {
	return c.KubeRestConfig(), nil
}
func (c *fakeclient) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}

func NewFakeClnt(c crclient.Client) *fakeclient {
	return &fakeclient{
		defaultNamespace: "default",
		Client:           c,
	}
}

func NewFakeCliClient(c crclient.Client) cli.Client {
	return &fakeclient{
		defaultNamespace: "default",
		Client:           c,
	}
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
	return resource.NewFakeBuilder(
		func(version schema.GroupVersion) (resource.RESTClient, error) {
			if c.UnstructuredClientForMappingFunc != nil {
				return c.UnstructuredClientForMappingFunc(version)
			}
			if c.UnstructuredClient != nil {
				return c.UnstructuredClient, nil
			}
			return c.Clients, nil
		},
		c.ToRESTMapper,
		func() (restmapper.CategoryExpander, error) {
			return resource.FakeCategoryExpander, nil
		},
	)
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
	UnstructuredClientForMappingFunc resource.FakeClientFunc
	UnstructuredClient               RESTClient
	Clients                          RESTClient
	ClientConfigVal                  *rest.Config
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
					{Name: "replicationcontrollers", Namespaced: true, Kind: "ReplicationController"},
					{Name: "componentstatuses", Namespaced: false, Kind: "ComponentStatus"},
					{Name: "nodes", Namespaced: false, Kind: "Node"},
					{Name: "secrets", Namespaced: true, Kind: "Secret"},
					{Name: "configmaps", Namespaced: true, Kind: "ConfigMap"},
					{Name: "namespacedtype", Namespaced: true, Kind: "NamespacedType"},
					{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
					{Name: "resourcequotas", Namespaced: true, Kind: "ResourceQuota"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "extensions",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1beta1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1beta1": {
					{Name: "deployments", Namespaced: true, Kind: "Deployment"},
					{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet"},
				},
			},
		},
	}
}
