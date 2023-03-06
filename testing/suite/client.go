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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// AppsPluginDynamicClient to client-go Dynamic client. All methods are relative to the
// namespace specified during construction
type DynamicClient interface {
	// Namespace in which this client is operating for
	Namespace() string

	// GetUsingGVK returns unmarshalled byte array provided GVK and name of the API object
	GetUsingGVK(ctx context.Context, gvk schema.GroupVersionKind, name string) (*unstructured.Unstructured, error)

	// ListUsingGVK returns unmarshalled byte array provided GVK and name of the API object
	ListUsingGVK(ctx context.Context, gvk schema.GroupVersionKind) (*unstructured.UnstructuredList, error)

	// RawClient returns the raw dynamic client interface
	RawClient() dynamic.Interface
}

// dynamicClient is a combination of client-go Dynamic client interface and namespace
type dynamicClient struct {
	client    dynamic.Interface
	namespace string
}

// NewDynamicClient is to invoke any resource to get/list obj in namespace
func NewDynamicClient(client dynamic.Interface, namespace string) DynamicClient {
	return &dynamicClient{
		client:    client,
		namespace: namespace,
	}
}

// Return the client's namespace
func (c *dynamicClient) Namespace() string {
	return c.namespace
}

func (c dynamicClient) RawClient() dynamic.Interface {
	return c.client
}

func (c *dynamicClient) GetUsingGVK(ctx context.Context, gvk schema.GroupVersionKind, name string) (*unstructured.Unstructured, error) {
	gvr := gvk.GroupVersion().WithResource(strings.ToLower(gvk.Kind) + "s")
	return c.client.Resource(gvr).Namespace(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *dynamicClient) ListUsingGVK(ctx context.Context, gvk schema.GroupVersionKind) (*unstructured.UnstructuredList, error) {
	gvr := gvk.GroupVersion().WithResource(strings.ToLower(gvk.Kind) + "s")
	return c.client.Resource(gvr).Namespace(c.namespace).List(ctx, metav1.ListOptions{})
}
