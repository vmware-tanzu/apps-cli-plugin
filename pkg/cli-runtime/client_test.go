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

package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	rtesting "github.com/vmware-labs/reconciler-runtime/testing"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	clitestingresource "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing/resource"
)

var kubeConfigPath = filepath.Join("testdata", ".kube", "config")

func TestNewClient(t *testing.T) {
	scheme := runtime.NewScheme()
	clitestingresource.AddToScheme(scheme)

	c := NewClient(kubeConfigPath, "", scheme)
	r := &clitestingresource.TestResource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "my-namespace",
			Name:      "my-resource",
		},
	}
	c.(*client).client = rtesting.NewFakeClient(scheme, r.DeepCopy())
	ctx := context.TODO()

	if c.KubeRestConfig() == nil {
		t.Errorf("unexpected restconfig")
	}
	if c.Scheme() != scheme {
		t.Errorf("unexpected scheme")
	}

	if c.Discovery() == nil {
		t.Errorf("expected discovery client")
	}
	if _, err := c.ToRESTConfig(); err != nil {
		t.Errorf("error durring get Restconfig: %v", err)
	}
	if _, err := c.ToDiscoveryClient(); err != nil {
		t.Errorf("error durring get DiscoveryClient: %v", err)
	}
	if _, err := c.ToRESTMapper(); err != nil {
		t.Errorf("error durring get RESTMapper: %v", err)
	}

	tr := &clitestingresource.TestResource{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: "my-namespace", Name: "my-resource"}, tr); err != nil {
		t.Errorf("error durring Get(): %v", err)
	}

	trl := &clitestingresource.TestResourceList{}
	if err := c.List(ctx, trl); err != nil {
		t.Errorf("error durring List(): %v", err)
	} else if len(trl.Items) != 1 {
		t.Errorf("unexpected list item")
	}
	if err := c.Update(ctx, tr); err != nil {
		t.Errorf("error durring Update(): %v", err)
	}
	if err := c.Create(ctx, tr); apierrs.IsAlreadyExists(err) {
		t.Errorf("expected AlreadyExists error durring Create(): %v", err)
	}
	if err := c.Delete(ctx, tr); err != nil {
		t.Errorf("error durring Delete(): %v", err)
	}
	if err := c.DeleteAllOf(ctx, tr); err != nil {
		t.Errorf("error durring DeleteAllOf(): %v", err)
	}
	if subresource := c.SubResource("pod"); subresource == nil {
		t.Errorf("error durring fetching Subresource()")
	}
}

func TestNewClientWithEnvVarKubeconfig(t *testing.T) {
	scheme := runtime.NewScheme()
	clitestingresource.AddToScheme(scheme)

	kubeconfig, kubeconfigisset := os.LookupEnv("KUBECONFIG")
	defer func() {
		if kubeconfigisset {
			os.Setenv("KUBECONFIG", kubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()
	os.Setenv("KUBECONFIG", kubeConfigPath)

	c := NewClient("", "", scheme)

	c.(*client).client = rtesting.NewFakeClient(scheme)

	if c.KubeRestConfig() == nil {
		t.Errorf("unexpected restconfig")
	}
	if c.DefaultNamespace() != "my-namespace" {
		t.Errorf("unexpected default namespace")
	}
	if _, err := c.ToRESTConfig(); err != nil {
		t.Errorf("error durring get Restconfig: %v", err)
	}
	if _, err := c.ToDiscoveryClient(); err != nil {
		t.Errorf("error durring get DiscoveryClient: %v", err)
	}
	if _, err := c.ToRESTMapper(); err != nil {
		t.Errorf("error durring get RESTMapper: %v", err)
	}
}

func TestNewClientWithEnvVarKubeconfigPathWithColon(t *testing.T) {
	scheme := runtime.NewScheme()
	clitestingresource.AddToScheme(scheme)

	kubeconfig, kubeconfigisset := os.LookupEnv("KUBECONFIG")
	defer func() {
		if kubeconfigisset {
			os.Setenv("KUBECONFIG", kubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()
	os.Setenv("KUBECONFIG", fmt.Sprintf("%s%s", kubeConfigPath, string(filepath.ListSeparator)))

	c := NewClient("", "", scheme)

	c.(*client).client = rtesting.NewFakeClient(scheme)

	if c.KubeRestConfig() == nil {
		t.Errorf("unexpected restconfig")
	}
	if c.DefaultNamespace() != "my-namespace" {
		t.Errorf("unexpected default namespace")
	}
	if _, err := c.ToRESTConfig(); err != nil {
		t.Errorf("error durring get Restconfig: %v", err)
	}
	if _, err := c.ToDiscoveryClient(); err != nil {
		t.Errorf("error durring get DiscoveryClient: %v", err)
	}
	if _, err := c.ToRESTMapper(); err != nil {
		t.Errorf("error durring get RESTMapper: %v", err)
	}
}
