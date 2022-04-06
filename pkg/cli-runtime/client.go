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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

var (
	// The rate limiter allows bursts of up to 'burst' to exceed the QPS, while still maintaining a
	// smoothed qps rate of 'qps'.
	// The bucket is initially filled with 'burst' tokens, and refills at a rate of 'qps'.
	// The maximum number of tokens in the bucket is capped at 'burst'.
	burst int     = 100
	qps   float32 = 100
)

type Client interface {
	DefaultNamespace() string
	KubeRestConfig() *rest.Config
	Discovery() discovery.DiscoveryInterface
	SetLogger(logger logr.Logger)
	crclient.Client
}

func (c *client) DefaultNamespace() string {
	return c.lazyLoadDefaultNamespaceOrDie()
}

func (c *client) KubeRestConfig() *rest.Config {
	return c.lazyLoadRestConfigOrDie()
}

func (c *client) Discovery() discovery.DiscoveryInterface {
	return c.lazyLoadKubernetesClientsetOrDie().Discovery()
}

func (c *client) Client() crclient.Client {
	return c.lazyLoadClientOrDie()
}

func (c *client) Get(ctx context.Context, key crclient.ObjectKey, obj crclient.Object) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "key", key, "action", "Get")
	err := c.Client().Get(ctx, key, obj)
	c.log.V(2).Info("Results", "object", obj)
	c.logError(err)
	return err
}

func (c *client) List(ctx context.Context, list crclient.ObjectList, opts ...crclient.ListOption) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "action", "List")
	err := c.Client().List(ctx, list, opts...)
	c.log.V(2).Info("Results", "objects", list)
	c.logError(err)
	return err
}

func (c *client) Create(ctx context.Context, obj crclient.Object, opts ...crclient.CreateOption) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "action", "Create", "object", obj)
	err := c.Client().Create(ctx, obj, opts...)
	c.log.V(2).Info("Results", "object", obj)
	c.logError(err)
	return err
}

func (c *client) Delete(ctx context.Context, obj crclient.Object, opts ...crclient.DeleteOption) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "action", "Delete", "object", obj)
	err := c.Client().Delete(ctx, obj, opts...)
	c.logError(err)
	return err
}

func (c *client) Update(ctx context.Context, obj crclient.Object, opts ...crclient.UpdateOption) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "action", "Update", "object", obj)
	err := c.Client().Update(ctx, obj, opts...)
	c.log.V(2).Info("Results", "object", obj)
	c.logError(err)
	return err
}

func (c *client) Patch(ctx context.Context, obj crclient.Object, patch crclient.Patch, opts ...crclient.PatchOption) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "action", "Patch", "data", patch)
	err := c.Client().Patch(ctx, obj, patch, opts...)
	c.log.V(2).Info("Results", "object", obj)
	c.logError(err)
	return err
}

func (c *client) DeleteAllOf(ctx context.Context, obj crclient.Object, opts ...crclient.DeleteAllOfOption) error {
	c.log.V(2).Info("API Request", "host", c.KubeRestConfig().Host, "action", "DeleteAllOf")
	err := c.Client().DeleteAllOf(ctx, obj, opts...)
	c.log.V(2).Info("Results", "object", obj)
	c.logError(err)
	return err
}

func (c *client) logError(err error) {
	if err != nil && c.log.V(2).Enabled() {
		c.log.V(2).Error(err, "API Error")
	}
}

func (c *client) Status() crclient.StatusWriter {
	panic(fmt.Errorf("not implemented"))
}

func (c *client) Scheme() *runtime.Scheme {
	return c.Client().Scheme()
}

func (c *client) RESTMapper() meta.RESTMapper {
	return c.Client().RESTMapper()
}

func (c *client) SetLogger(logger logr.Logger) {
	c.log = logger
}

func NewClient(kubeConfigFile string, currentContext string, scheme *runtime.Scheme) Client {
	return &client{
		kubeConfigFile: kubeConfigFile,
		currentContext: currentContext,
		scheme:         scheme,
		log:            logr.Discard(),
	}
}

type client struct {
	defaultNamespace string
	kubeConfigFile   string
	currentContext   string
	scheme           *runtime.Scheme
	kubeConfig       clientcmd.ClientConfig
	restConfig       *rest.Config
	kubeClientset    *kubernetes.Clientset
	client           crclient.Client
	log              logr.Logger
}

func (c *client) lazyLoadKubeConfig() clientcmd.ClientConfig {
	if c.kubeConfig == nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = c.kubeConfigFile
		configOverrides := &clientcmd.ConfigOverrides{
			CurrentContext: c.currentContext,
		}
		c.kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	}
	return c.kubeConfig
}

func (c *client) lazyLoadRestConfigOrDie() *rest.Config {
	if c.restConfig == nil {
		kubeConfig := c.lazyLoadKubeConfig()
		restConfig, err := kubeConfig.ClientConfig()
		if err != nil {
			if clientcmd.IsEmptyConfig(err) {
				fmt.Printf("%s Unable to connect: no configuration has been found. If a kubeconfig is not set, it can be provided with the --kubeconfig flag or KUBECONFIG environment variable.\n", printer.Serrorf("Error:"))
			} else {
				fmt.Printf("%s %v \n", printer.Serrorf("Error:"), err)
			}
			c.logError(err)
			os.Exit(2)
		}
		restConfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(qps, burst)
		c.restConfig = restConfig
	}
	return c.restConfig
}

func (c *client) lazyLoadKubernetesClientsetOrDie() *kubernetes.Clientset {
	if c.kubeClientset == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		c.kubeClientset = kubernetes.NewForConfigOrDie(restConfig)
	}
	return c.kubeClientset
}

func (c *client) lazyLoadClientOrDie() crclient.Client {
	if c.client == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		client, err := crclient.New(restConfig, crclient.Options{Scheme: c.scheme})
		if err != nil {
			fmt.Printf("%s Unable to connect: connection refused. Confirm kubeconfig details and try again.\n", printer.Serrorf("Error:"))
			c.logError(err)
			os.Exit(2)
		}
		c.client = client
	}
	return c.client
}

func (c *client) lazyLoadDefaultNamespaceOrDie() string {
	if c.defaultNamespace == "" {
		kubeConfig := c.lazyLoadKubeConfig()
		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			if clientcmd.IsEmptyConfig(err) {
				fmt.Printf("%s Unable to connect: no configuration has been found. If a kubeconfig is not set, it can be provided with the --kubeconfig flag or KUBECONFIG environment variable.\n", printer.Serrorf("Error:"))
			} else {
				fmt.Printf("%s %v \n", printer.Serrorf("Error:"), err)
			}

			c.logError(err)
			os.Exit(2)
		}
		c.defaultNamespace = namespace
	}
	return c.defaultNamespace
}
