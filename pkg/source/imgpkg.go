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
	"net/http"
	"path"

	regname "github.com/google/go-containerregistry/pkg/name"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/plainimage"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/internal/util"
)

func ImgpkgPush(dir string, excludedFiles []string, writer plainimage.ImagesWriter, image string) (string, error) {

	uploadRef, err := regname.NewTag(image, regname.WeakValidation)
	if err != nil {
		return "", fmt.Errorf("parsing '%s': %s", image, err)
	}
	excludedFiles = append(excludedFiles, path.Join(dir, ".imgpkg"))
	digest, err := plainimage.NewContents([]string{dir}, excludedFiles).Push(uploadRef, nil, writer, util.NewNoopLevelLogger())
	if err != nil {
		return "", err
	}

	// get an image ref with a tag and digest
	digestRef, _ := regname.NewDigest(digest, regname.WeakValidation)
	return fmt.Sprintf("%s@%s", uploadRef.Name(), digestRef.DigestStr()), nil
}

type registryOptionsStashKey struct{}
type containerRemoteTransportStashKey struct{}

func StashContainerRemoteTransport(ctx context.Context, rTripper http.RoundTripper) context.Context {
	return context.WithValue(ctx, containerRemoteTransportStashKey{}, rTripper)
}

func RetrieveStashContainerRemoteTransport(ctx context.Context) *http.RoundTripper {
	transport, ok := ctx.Value(containerRemoteTransportStashKey{}).(http.RoundTripper)
	if !ok {
		return nil
	}

	return &transport
}
