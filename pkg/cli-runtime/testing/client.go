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
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func (c *fakeclient) DefaultNamespace() string {
	return "default"
}

func (c *fakeclient) KubeRestConfig() *rest.Config {
	panic(fmt.Errorf("not implemented"))
}

func (c *fakeclient) Discovery() discovery.DiscoveryInterface {
	panic(fmt.Errorf("not implemented"))
}

func (c *fakeclient) SetLogger(logger logr.Logger) {
	panic(fmt.Errorf("not implemented"))
}

func NewFakeCliClient(c crclient.Client) cli.Client {
	return &fakeclient{
		defaultNamespace: "default",
		Client:           c,
	}
}

type fakeclient struct {
	defaultNamespace string
	crclient.Client
}
