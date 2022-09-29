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

package printer

import (
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

type Object = printer.Object

var ExportResource = printer.ExportResource
var OutputResource = printer.OutputResource
var FindCondition = printer.FindCondition
var ResourceDiff = printer.ResourceDiff
var ResourceStatus = printer.ResourceStatus
var Serrorf = printer.Serrorf
var SortByNamespaceAndName = printer.SortByNamespaceAndName

type OutputFormat = printer.OutputFormat

var OutputFormatJson = printer.OutputFormatJson
var OutputFormatYaml = printer.OutputFormatYaml
var OutputFormatYml = printer.OutputFormatYml
