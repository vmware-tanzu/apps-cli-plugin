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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/stern/stern/stern"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

const ansi = "[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]"

var _ Tailer = &SternTailer{}
var re = regexp.MustCompile(ansi)

type SternTailer struct{}

func (s *SternTailer) Tail(ctx context.Context, c *cli.Config, namespace string, selector labels.Selector, containers []string, since time.Duration, timestamps bool) error {
	containerQuery := regexp.MustCompile(".*")
	if len(containers) != 0 {
		escapedContainers := []string{}
		for _, c := range containers {
			escapedContainers = append(escapedContainers, regexp.QuoteMeta(c))
		}
		containerQuery = regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(escapedContainers, "|")))
	}
	t := "{{color .ContainerColor .PodName}}{{color .PodColor \"[\"}}{{color .PodColor .ContainerName}}{{color .PodColor \"]\"}} {{format .Message}}\n"
	funs := map[string]interface{}{
		"json": func(in interface{}) (string, error) {
			b, err := json.Marshal(in)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"format": func(in string) string {
			return stripANSIColor(in)
		},
		"color": func(color color.Color, text string) string {
			return color.SprintFunc()(text)
		},
	}
	template, err := template.New("log").Funcs(funs).Parse(t)
	if err != nil {
		panic(err)
	}

	configStern := stern.Config{
		KubeConfig:     c.KubeConfigFile,
		ContextName:    c.CurrentContext,
		Namespaces:     []string{namespace},
		Timestamps:     timestamps,
		Location:       time.Local,
		LabelSelector:  selector,
		ContainerQuery: containerQuery,
		ContainerStates: []stern.ContainerState{
			stern.RUNNING,
			stern.TERMINATED,
		},
		InitContainers: true,
		Since:          since,

		// PodQuery and FieldSelector are required, but we use LabelSelector instead
		PodQuery:      regexp.MustCompile(""),
		FieldSelector: fields.Everything(),

		Template: template,
		Out:      c.Stdout,
		ErrOut:   c.Stderr,
		Follow:   true,
	}

	return stern.Run(ctx, &configStern)
}

func stripANSIColor(message string) string {
	if color.NoColor {
		return re.ReplaceAllString(message, "")
	} else {
		return message
	}
}
