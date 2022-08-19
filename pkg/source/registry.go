package source

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/registry"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/internal/util"
)

type RegistryOpts struct {
	CACertPaths      []string
	RegistryUsername string
	RegistryPassword string
	RegistryToken    string
}
type WriterObjects struct {
	OutWriter io.Writer
	ErrWriter io.Writer
}

func GetRegistry(ctx context.Context, registryOpts *RegistryOpts) (registry.Registry, error) {
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
	transport := RetrieveStashContainerRemoteTransport(ctx)
	if transport == nil {
		reg, err = registry.NewSimpleRegistry(options)
	} else {
		reg, err = registry.NewSimpleRegistryWithTransport(options, *transport)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create a registry with provided options: %v", err)
	}
	return reg, nil
}

// RegistrywithProgress created new registry with Progress Implementing Registry interface and provides
// progress updates to the logger
func GetRegistryWithProgress(reg registry.Registry, writer *WriterObjects) *registry.WithProgress {
	levelLogger := util.NewUILevelLogger(util.LogWarn, util.NewLogger(ui.NewWriterUI(writer.OutWriter, writer.OutWriter, nil)))
	imagesUploaderLogger := util.NewProgressBar(levelLogger, "", "", writer.OutWriter)
	return registry.NewRegistryWithProgress(reg, imagesUploaderLogger)
}
