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

package logs

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/labels"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

var _ Tailer = &FakeTailer{}

type FakeTailer struct {
	mock.Mock
}

func (f *FakeTailer) Tail(ctx context.Context, c *cli.Config, namespace string, selector labels.Selector, containers []string, since time.Duration, timestamps bool) error {
	args := f.Called(ctx, namespace, selector, containers, since, timestamps)
	c.Printf(color.CyanString("...tail output...\n"))
	if err := args.Error(0); err != nil {
		return err
	}
	// simulate tailing until the context is closed
	<-ctx.Done()
	return nil
}
