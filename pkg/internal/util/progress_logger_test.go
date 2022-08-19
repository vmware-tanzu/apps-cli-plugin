package util_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/google/go-cmp/cmp"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/internal/util"
)

func TestProgressLogger(t *testing.T) {
	t.Run("Test progress logger", func(t *testing.T) {
		outputwriter := &bytes.Buffer{}
		scheme := runtime.NewScheme()
		c := cli.NewDefaultConfig("test", scheme)
		c.Stdout = outputwriter
		levelLogger := util.NewUILevelLogger(util.LogWarn, util.NewLogger(ui.NewWriterUI(c.Stdout, c.Stdout, nil)))
		testprogresslogger := util.NewNoopProgressBar(levelLogger)
		processData(testprogresslogger)
		expectedOutput := `
Progress Bar End
		`

		if diff := cmp.Diff(strings.TrimSpace(expectedOutput), strings.TrimSpace(outputwriter.String())); diff != "" {
			t.Errorf("PublishLocalSource() (-want, +got) = %s", diff)
		}

	})

}

func processData(testprogresslogger util.ProgressLogger) {

	valuesChannel := make(chan regv1.Update)
	regObject := []regv1.Update{{Total: 100, Complete: 0, Error: nil}, {Total: 200, Complete: 0, Error: nil}}

	testprogresslogger.Start(context.Background(), valuesChannel)
	for _, val := range regObject {
		valuesChannel <- val
		time.Sleep(100 * time.Millisecond)
	}
	defer testprogresslogger.End()
	defer close(valuesChannel)

}
