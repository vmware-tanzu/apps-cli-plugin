/*
Copyright 2023 VMware, Inc.

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

package source

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	ImageTag             = "source"
	sourceProxyService   = "local-source-proxy"
	sourceProxyNamespace = "tap-local-source-system"
	sourceProxyDomain    = "svc.cluster.local"
	servicesResource     = "services"
	proxySubResource     = "proxy"
)

func GetLocalImageRepo() string {
	return fmt.Sprintf("%s.%s.%s", sourceProxyService, sourceProxyNamespace, sourceProxyDomain)
}

func LocalRegistryTransport(ctx context.Context, kubeconfig *rest.Config, restClient rest.Interface) (*Wrapper, error) {
	if wrapper := RetrieveContainerWrapper(ctx); wrapper != nil {
		return wrapper, nil
	}

	if restClient == nil {
		return nil, errors.New("no RESTClient was set for local proxy transport")
	}
	r := restClient.Get().Namespace(sourceProxyNamespace).Resource(servicesResource).SubResource(proxySubResource).Name(net.JoinSchemeNamePort("http", sourceProxyService, "5001"))

	gv := corev1.SchemeGroupVersion
	kubeconfig.GroupVersion = &gv
	kubeconfig.APIPath = "/api"
	kubeconfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	client, err := rest.RESTClientFor(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Wrap transport to rewrite paths
	return &Wrapper{
		Client:     client.Client,
		URL:        r.URL(),
		Repository: "",
	}, nil
}
