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
	"fmt"
	"os"
	"os/user"
	"path"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

type ClientConfig struct {
	config *cli.Config
	ctx    context.Context
}

func NewClientConfig() (*ClientConfig, error) {
	ctx := context.Background()
	scheme := runtime.NewScheme()

	_ = clientgoscheme.AddToScheme(scheme)
	_ = cartov1alpha1.AddToScheme(scheme)

	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("could not get current user: %v", err)
	}
	kubeConfig := path.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName)
	os.Setenv("KUBECONFIG", kubeConfig)
	c := cli.NewDefaultConfig("tanzu apps", scheme)
	c.KubeConfigFile = kubeConfig
	c.Client = cli.NewClient(c.KubeConfigFile, c.CurrentContext, c.Scheme)

	return &ClientConfig{
		config: c,
		ctx:    ctx,
	}, nil
}

func (c *ClientConfig) Get(key client.ObjectKey, obj client.Object) error {
	return c.config.Get(c.ctx, key, obj)
}

func (c *ClientConfig) List(list client.ObjectList, opts ...client.ListOption) error {
	return c.config.List(c.ctx, list, opts...)
}
