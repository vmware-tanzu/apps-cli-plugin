package cli

import (
	"fmt"

	"github.com/vito/go-interact/interact"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

// NewConfirmSurvey create a survey asking for [yN] confirmation when `Resolve` is called
func NewConfirmSurvey(c *Config, format string, a ...any) *interact.Interaction {
	i := interact.NewInteraction(fmt.Sprintf("%s %s", printer.Ssuccessf("?"), printer.Sboldf(fmt.Sprintf(format, a...))))
	i.Input = c.Stdin
	i.Output = c.Stdout
	return &i
}
