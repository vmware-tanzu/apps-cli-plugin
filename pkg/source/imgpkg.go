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
	"fmt"

	"os"
	"path"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/plainimage"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/registry"
)

func ImgpkgPush(ctx context.Context, dir string, excludedFiles []string, image string) (string, error) {
	options := RetrieveGgcrRemoteOptions(ctx)
	envFunction := func() []string {
		envVars := os.Environ()
		_, present := os.LookupEnv("IMGPKG_ENABLE_IAAS_AUTH")
		if !present {
			envVars = append(envVars, "IMGPKG_ENABLE_IAAS_AUTH=false")
		}
		return envVars
	}

	options.EnvironFunc = envFunction

	// TODO: support more registry options using apps plugin configuration
	reg, err := registry.NewSimpleRegistry(options)
	if err != nil {
		return "", fmt.Errorf("unable to create a registry with provided options: %v", err)
	}

	uploadRef, err := regname.NewTag(image, regname.WeakValidation)
	if err != nil {
		return "", fmt.Errorf("parsing '%s': %s", image, err)
	}

	excludedFiles = append(excludedFiles, path.Join(dir, ".imgpkg"))
	digest, err := plainimage.NewContents([]string{dir}, excludedFiles).Push(uploadRef, nil, reg, ui.NewNoopUI())
	if err != nil {
		return "", err
	}

	// get an image ref with a tag and digest
	digestRef, _ := regname.NewDigest(digest, regname.WeakValidation)
	return fmt.Sprintf("%s@%s", uploadRef.Name(), digestRef.DigestStr()), nil
}

type ggcrRemoteOptionsStashKey struct{}

func StashGgcrRemoteOptions(ctx context.Context, options registry.Opts) context.Context {
	return context.WithValue(ctx, ggcrRemoteOptionsStashKey{}, options)
}

func RetrieveGgcrRemoteOptions(ctx context.Context) registry.Opts {
	options, ok := ctx.Value(ggcrRemoteOptionsStashKey{}).(registry.Opts)
	if !ok {
		return registry.Opts{
			VerifyCerts:           true,
			RetryCount:            8,
			ResponseHeaderTimeout: 300 * time.Second,
		}
	}
	return options
}
