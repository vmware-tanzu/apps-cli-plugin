/*
Copyright 2023 VMware, Inc.

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

package printer

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/lsp"
)

func PrintLocalSourceProxyStatus(w io.Writer, printType string, s lsp.LSPStatus) error {
	switch printType {
	case "json":
		if m, err := json.MarshalIndent(s, "", "  "); err != nil {
			return err
		} else {
			w.Write(m)
		}
	case "yaml", "yml", "":
		if m, err := yaml.Marshal(s); err != nil {
			return err
		} else {
			w.Write(m)
		}
	default:
		return fmt.Errorf("output not supported, supported formats: 'json', 'yaml', 'yml'")
	}
	return nil
}
