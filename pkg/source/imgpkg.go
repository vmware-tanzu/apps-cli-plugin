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
	"path"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/k14s/imgpkg/pkg/imgpkg/plainimage"
	"github.com/k14s/imgpkg/pkg/imgpkg/registry"
)

func ImgpkgPush(ctx context.Context, dir string, image string) (string, error) {
	options := RetrieveGgcrRemoteOptions(ctx)

	// TODO support more registry options
	reg, err := registry.NewRegistry(registry.Opts{VerifyCerts: true}, options...)
	if err != nil {
		return "", fmt.Errorf("unable to create a registry with provided options: %v", err)
	}

	uploadRef, err := regname.NewTag(image, regname.WeakValidation)
	if err != nil {
		return "", fmt.Errorf("parsing '%s': %s", image, err)
	}

	digest, err := plainimage.NewContents([]string{dir}, []string{path.Join(dir, ".imgpkg")}).Push(uploadRef, nil, reg, ui.NewNoopUI())
	if err != nil {
		return "", err
	}

	// get an image ref with a tag and digest
	digestRef, _ := regname.NewDigest(digest, regname.WeakValidation)
	return fmt.Sprintf("%s@%s", uploadRef.Name(), digestRef.DigestStr()), nil
}

type ggcrRemoteOptionsStashKey struct{}

func StashGgcrRemoteOptions(ctx context.Context, options ...remote.Option) context.Context {
	return context.WithValue(ctx, ggcrRemoteOptionsStashKey{}, options)
}

func RetrieveGgcrRemoteOptions(ctx context.Context) []remote.Option {
	options, ok := ctx.Value(ggcrRemoteOptionsStashKey{}).([]remote.Option)
	if !ok {
		return []remote.Option{}
	}
	return options
}
