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

package source

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
)

const ImageTag = "source"
const SourceProxyService = "local-source-proxy"
const SourceProxyNamespace = "tap-local-source-system"

func LocalRegistryTransport(ctx context.Context, cl *kubernetes.Clientset,
	kubeconfig *rest.Config) (*Wrapper, error) {

	_, err := cl.CoreV1().Services(SourceProxyNamespace).Get(ctx, SourceProxyService, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	r := cl.CoreV1().RESTClient().Get().Namespace(SourceProxyNamespace).Resource("services").SubResource("proxy").Name(net.JoinSchemeNamePort("http", SourceProxyService, "5001"))

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
