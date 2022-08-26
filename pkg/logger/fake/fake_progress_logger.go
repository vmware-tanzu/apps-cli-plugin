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

package fake

import (
	"context"

	regv1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/logger"
)

// NewNoopProgressBar constructs a Noop Progress bar that will not display anything
func NewNoopProgressBar() logger.ProgressLogger {
	return &progressBarNoTTY{}
}

// progressBarNoTTY does not display the progress bar
type progressBarNoTTY struct {
	cancelFunc context.CancelFunc
}

// Start consuming the progress channel but does not display anything
func (l *progressBarNoTTY) Start(ctx context.Context, progressChan <-chan regv1.Update) {
	ctx, cancelFunc := context.WithCancel(ctx)
	l.cancelFunc = cancelFunc
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-progressChan:
			}
		}
	}()
}

// End stops the progress bar
func (l *progressBarNoTTY) End() {
	if l.cancelFunc != nil {
		l.cancelFunc()
	}
}
