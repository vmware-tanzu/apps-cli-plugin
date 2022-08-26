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

package logger

import (
	"context"

	pb "github.com/cheggaaa/pb/v3"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

// ProgressLogger interface should match with imgpkg util Progress logger
type ProgressLogger interface {
	Start(ctx context.Context, progress <-chan regv1.Update)
	End()
}

// NewProgressBar constructor to build a ProgressLogger responsible for printing out a progress bar using updates when
// writing to a registry via ggcr
func NewProgressBar() ProgressLogger {
	return &progressBar{}
}

// progressBar display progress bar on output
type progressBar struct {
	cancelFunc context.CancelFunc
	bar        *pb.ProgressBar
}

// Start displays the Progress Bar
func (l *progressBar) Start(ctx context.Context, progressChan <-chan regv1.Update) {
	ctx, cancelFunc := context.WithCancel(ctx)
	l.cancelFunc = cancelFunc
	l.bar = pb.New64(0)
	l.bar.Set(pb.Bytes, true)
	l.bar.Set(pb.SIBytesPrefix, true)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-progressChan:
				if update.Error != nil {
					continue
				}
				if update.Total == 0 {
					return
				}
				if !l.bar.IsStarted() {
					l.bar.SetTotal(update.Total)
					l.bar.Start()
				}
				l.bar.SetCurrent(update.Complete)
				l.bar.Write()
			}
		}
	}()
}

// End stops the progress bar
func (l *progressBar) End() {
	if l.cancelFunc != nil {
		l.cancelFunc()
	}
	l.bar.Finish()
}

type progressLoggerKey struct{}

func RetrieveProgressBarLogger(ctx context.Context) ProgressLogger {
	logger, ok := ctx.Value(progressLoggerKey{}).(ProgressLogger)
	if !ok {
		return nil
	}
	return logger
}

func StashProgressBarLogger(ctx context.Context, logger ProgressLogger) context.Context {
	return context.WithValue(ctx, progressLoggerKey{}, logger)
}
