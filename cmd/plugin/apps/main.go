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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	tanzucliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/logs"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/logger"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = cartov1alpha1.AddToScheme(scheme)
	_ = knativeservingv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	p, err := plugin.NewPlugin(&tanzucliv1alpha1.PluginDescriptor{
		Name:           "apps",
		Description:    "Applications on Kubernetes",
		Group:          tanzucliv1alpha1.BuildCmdGroup,
		CompletionType: tanzucliv1alpha1.NativePluginCompletion,
		Aliases:        []string{"app"},
		Version:        buildinfo.Version,
		BuildSHA:       buildinfo.SHA,
	})
	if err != nil {
		log.Fatal(err)
	}

	// deactivate default commands
	p.Cmd.CompletionOptions.DisableDefaultCmd = true // wokeignore:rule=disable

	// setup logs.Tail() for all commands using stern
	ctx = logs.StashTailer(ctx, &logs.SternTailer{})

	c := cli.Initialize(fmt.Sprintf("tanzu %s", p.Cmd.Use), scheme)
	p.AddCommands(
		commands.NewClusterSupplyChainCommand(ctx, c),
		commands.NewWorkloadCommand(ctx, c),

		// hidden commands
		commands.NewDocsCommand(ctx, c),
		commands.NewLocalSourceProxyCommand(ctx, c),
	)

	// add root persistent flags
	// TODO can we normalize all of these flags?
	p.Cmd.PersistentFlags().StringVar(&c.KubeConfigFile, cli.StripDash(flags.KubeConfigFlagName), "", "kubeconfig `file` (default is $HOME/.kube/config)")
	p.Cmd.MarkFlagFilename(cli.StripDash(flags.KubeConfigFlagName))
	p.Cmd.PersistentFlags().StringVar(&c.CurrentContext, cli.StripDash(flags.ContextFlagName), "", "`name` of the kubeconfig context to use (default is current-context defined by kubeconfig)")
	p.Cmd.PersistentFlags().BoolVar(&color.NoColor, cli.StripDash(flags.NoColorFlagName), color.NoColor, "deactivate color, bold, animations, and emoji output")
	p.Cmd.PersistentFlags().Int32VarP(c.Verbose, cli.StripDash(flags.VerboseLevelFlagName), "v", 1, "number for the log level verbosity")

	cobra.OnInitialize(func() {
		// sync config and fatih to deactivate emojis printing
		c.NoColor = color.NoColor
		// set the default logger
		c.SetLogger(logger.NewSinkLogger(c.Name, c.Verbose, c.Stderr))
	})

	// override usage template to add arguments
	p.Cmd.SetUsageTemplate(strings.ReplaceAll(p.Cmd.UsageTemplate(), "{{.UseLine}}", "{{useLine .}}"))
	cobra.AddTemplateFunc("useLine", func(cmd *cobra.Command) string {
		result := cmd.UseLine()
		flags := ""
		if strings.HasSuffix(result, " [flags]") {
			flags = " [flags]"
			result = result[0 : len(result)-len(flags)]
		}
		return result + cli.FormatArgs(cmd) + flags
	})

	// override default colors
	printer.InfoColor = color.New(color.FgCyan, color.Bold)
	printer.SuccessColor = color.New(color.FgGreen, color.Bold)
	printer.WarnColor = color.New(color.FgYellow, color.Bold)
	printer.ErrorColor = color.New(color.FgRed, color.Bold)

	p.Cmd.SilenceErrors = true
	if err := p.Execute(); err != nil {
		// silent errors should not log, but still exit with an error code
		// typically the command has already been logged with more detail
		if !errors.Is(err, cli.SilentError) {
			if aggregate, ok := err.(utilerrors.Aggregate); ok {
				for _, err := range aggregate.Errors() {
					c.Eprintf("%s %s\n", printer.Serrorf("Error:"), err)
				}
			} else if apierrors.IsForbidden(err) {
				c.Eprintf("%s %s: %s\n", printer.Serrorf("Error:"), "Unable to complete command, you do not have permissions for this resource", err)
			} else {
				c.Eprintf("%s %s\n", printer.Serrorf("Error:"), err)
			}
		}
		os.Exit(1)
	}
}
