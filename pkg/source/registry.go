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

package source

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/registry"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/logger"
)

type RegistryOpts struct {
	CACertPaths      []string
	RegistryUsername string
	RegistryPassword string
	RegistryToken    string
}

// NewRegistryWithProgress creates new registry instance that provides
// progress updates to the logger
func NewRegistryWithProgress(ctx context.Context, registryOpts *RegistryOpts) (*registry.WithProgress, error) {
	options := registry.Opts{
		CACertPaths:           registryOpts.CACertPaths,
		Username:              registryOpts.RegistryUsername,
		Password:              registryOpts.RegistryPassword,
		Token:                 registryOpts.RegistryToken,
		VerifyCerts:           true,
		RetryCount:            5,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	var reg registry.Registry
	var err error
	transport := RetrieveContainerRemoteTransport(ctx)
	if transport == nil {
		reg, err = registry.NewSimpleRegistry(options)
	} else {
		reg, err = registry.NewSimpleRegistryWithTransport(options, *transport)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create a registry with provided options: %v", err)
	}
	progressBar := logger.RetrieveProgressBarLogger(ctx)
	if progressBar == nil {
		progressBar = logger.NewProgressBar()
	}
	return registry.NewRegistryWithProgress(reg, progressBar), nil
}
