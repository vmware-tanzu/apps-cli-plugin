// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	"github.com/k14s/imgpkg/pkg/imgpkg/bundle"
	"github.com/k14s/imgpkg/pkg/imgpkg/lockconfig"
	"github.com/k14s/imgpkg/pkg/imgpkg/plainimage"
	"github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/spf13/cobra"
)

type PushOptions struct {
	ui ui.UI

	ImageFlags      ImageFlags
	BundleFlags     BundleFlags
	LockOutputFlags LockOutputFlags
	FileFlags       FileFlags
	RegistryFlags   RegistryFlags
}

func NewPushOptions(ui ui.UI) *PushOptions {
	return &PushOptions{ui: ui}
}

func NewPushCmd(o *PushOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push files as image",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
		Example: `
  # Push bundle repo/app1-config with contents of config/ directory
  imgpkg push -b repo/app1-config -f config/

  # Push image repo/app1-config with contents from multiple locations
  imgpkg push -i repo/app1-config -f config/ -f additional-config.yml`,
	}
	o.ImageFlags.Set(cmd)
	o.BundleFlags.Set(cmd)
	o.LockOutputFlags.Set(cmd)
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	return cmd
}

func (po *PushOptions) Run() error {
	reg, err := registry.NewRegistry(po.RegistryFlags.AsRegistryOpts())
	if err != nil {
		return fmt.Errorf("Unable to create a registry with provided options: %v", err)
	}

	var imageURL string

	isBundle := po.BundleFlags.Bundle != ""
	isImage := po.ImageFlags.Image != ""

	switch {
	case isBundle && isImage:
		return fmt.Errorf("Expected only one of image or bundle")

	case !isBundle && !isImage:
		return fmt.Errorf("Expected either image or bundle")

	case isBundle:
		imageURL, err = po.pushBundle(reg)
		if err != nil {
			return err
		}

	case isImage:
		imageURL, err = po.pushImage(reg)
		if err != nil {
			return err
		}

	default:
		panic("Unreachable code")
	}

	po.ui.BeginLinef("Pushed '%s'", imageURL)

	return nil
}

func (po *PushOptions) pushBundle(registry registry.Registry) (string, error) {
	uploadRef, err := regname.NewTag(po.BundleFlags.Bundle, regname.WeakValidation)
	if err != nil {
		return "", fmt.Errorf("Parsing '%s': %s", po.BundleFlags.Bundle, err)
	}

	imageURL, err := bundle.NewContents(po.FileFlags.Files, po.FileFlags.ExcludedFilePaths).Push(uploadRef, registry, po.ui)
	if err != nil {
		return "", err
	}

	if po.LockOutputFlags.LockFilePath != "" {
		bundleLock := lockconfig.BundleLock{
			LockVersion: lockconfig.LockVersion{
				APIVersion: lockconfig.BundleLockAPIVersion,
				Kind:       lockconfig.BundleLockKind,
			},
			Bundle: lockconfig.BundleRef{
				Image: imageURL,
				Tag:   uploadRef.TagStr(),
			},
		}

		err := bundleLock.WriteToPath(po.LockOutputFlags.LockFilePath)
		if err != nil {
			return "", err
		}
	}

	return imageURL, nil
}

func (po *PushOptions) pushImage(registry registry.Registry) (string, error) {
	if po.LockOutputFlags.LockFilePath != "" {
		return "", fmt.Errorf("Lock output is not compatible with image, use bundle for lock output")
	}

	uploadRef, err := regname.NewTag(po.ImageFlags.Image, regname.WeakValidation)
	if err != nil {
		return "", fmt.Errorf("Parsing '%s': %s", po.ImageFlags.Image, err)
	}

	isBundle, err := bundle.NewContents(po.FileFlags.Files, po.FileFlags.ExcludedFilePaths).PresentsAsBundle()
	if err != nil {
		return "", err
	}
	if isBundle {
		return "", fmt.Errorf("Images cannot be pushed with '.imgpkg' directories, consider using --bundle (-b) option")
	}

	return plainimage.NewContents(po.FileFlags.Files, po.FileFlags.ExcludedFilePaths).Push(uploadRef, nil, registry, po.ui)
}
