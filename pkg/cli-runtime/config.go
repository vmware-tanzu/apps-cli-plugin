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
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"

	// "k8s.io/cli-runtime/pkg/resource"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

const defaultTanzuIgnoreFile = ".tanzuignore"

type Config struct {
	Name string
	Client
	Scheme          *runtime.Scheme
	ViperConfigFile string
	KubeConfigFile  string
	CurrentContext  string
	TanzuIgnoreFile string
	Exec            func(ctx context.Context, command string, args ...string) *exec.Cmd
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
	Verbose         *int32
	Builder         *resource.Builder
}

func NewDefaultConfig(name string, scheme *runtime.Scheme) *Config {
	var v int32
	return &Config{
		Name:            name,
		Scheme:          scheme,
		Exec:            exec.CommandContext,
		Stdin:           os.Stdin,
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
		Verbose:         &v,
		TanzuIgnoreFile: defaultTanzuIgnoreFile,
	}
}

func (c *Config) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Eprintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Infof(format string, a ...interface{}) (n int, err error) {
	return printer.InfoColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Einfof(format string, a ...interface{}) (n int, err error) {
	return printer.InfoColor.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Successf(format string, a ...interface{}) (n int, err error) {
	return printer.SuccessColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Esuccessf(format string, a ...interface{}) (n int, err error) {
	return printer.SuccessColor.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Errorf(format string, a ...interface{}) (n int, err error) {
	return printer.ErrorColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Eerrorf(format string, a ...interface{}) (n int, err error) {
	return printer.ErrorColor.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Faintf(format string, a ...interface{}) (n int, err error) {
	return printer.FaintColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Efaintf(format string, a ...interface{}) (n int, err error) {
	return printer.FaintColor.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Boldf(format string, a ...interface{}) (n int, err error) {
	return printer.BoldColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Eboldf(format string, a ...interface{}) (n int, err error) {
	return printer.BoldColor.Fprintf(c.Stderr, format, a...)
}

func Initialize(name string, scheme *runtime.Scheme) *Config {
	c := NewDefaultConfig(name, scheme)

	cobra.OnInitialize(c.init)

	return c
}

func (c *Config) init() {
	if c.Client == nil {
		c.Client = NewClient(c.KubeConfigFile, c.CurrentContext, c.Scheme)
	}
	if c.Builder == nil {
		c.Builder = resource.NewBuilder(c.Client)
	}
}
