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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/vmware-tanzu/difflib"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type OutputFormat string

const (
	OutputFormatJson = "json"
	OutputFormatYaml = "yaml"
	OutputFormatYml  = "yml"
)

type Object interface {
	client.Object
	metav1.ObjectMetaAccessor
	schema.ObjectKind
}

func ExportResource(obj Object, format OutputFormat, scheme *runtime.Scheme) (string, error) {
	copy := obj.DeepCopyObject().(Object)

	// force apiVersion and kind to be set
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return "", err
	}
	copy.SetGroupVersionKind(gvks[0])

	// pune ObjectMeta to generateName or name, annotations, and labels
	copy.GetObjectMeta().(*metav1.ObjectMeta).Reset()
	if obj.GetGenerateName() != "" {
		copy.SetGenerateName(obj.GetGenerateName())
	} else {
		copy.SetName(obj.GetName())
	}
	copy.SetNamespace(obj.GetNamespace())
	copy.SetAnnotations(obj.GetAnnotations())
	copy.SetLabels(obj.GetLabels())

	// remove status and other nuisance fields
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(copy)
	if err != nil {
		return "", err
	}

	unstructured.RemoveNestedField(u, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(u, "status")

	return printObject(u, format)
}

func setGVK(obj Object, scheme *runtime.Scheme) (Object, error) {
	copy := obj.DeepCopyObject().(Object)

	// force apiVersion and kind to be set
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	copy.SetGroupVersionKind(gvks[0])
	return copy, nil
}

func OutputResource(obj Object, format OutputFormat, scheme *runtime.Scheme) (string, error) {
	copy, err := setGVK(obj, scheme)
	if err != nil {
		return "", err
	}
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(copy)
	if err != nil {
		return "", err
	}

	unstructured.RemoveNestedField(u, "metadata", "managedFields")

	return printObject(u, format)
}

func OutputResources(objList []Object, format OutputFormat, scheme *runtime.Scheme) (string, error) {
	updatedList := []Object{}

	for _, o := range objList {
		copy, err := setGVK(o, scheme)
		if err != nil {
			return "", err
		}
		updatedList = append(updatedList, copy)
	}
	return printObject(updatedList, format)
}

func printObject(obj interface{}, format OutputFormat) (string, error) {
	// render according to desired format
	switch format {
	case OutputFormatJson:
		b, err := json.MarshalIndent(obj, "", "\t")
		return strings.TrimSpace(string(b)), err
	case OutputFormatYaml, OutputFormatYml:
		b, err := yaml.Marshal(obj)
		return fmt.Sprintf("---\n%s", strings.TrimSpace(string(b))), err
	default:
		return "", fmt.Errorf("unknown output format %q", format)
	}
}

var (
	DiffAdditionColor    = color.New(color.FgGreen)
	DiffSubtractionColor = color.New(color.FgRed)
	DiffUnchangedColor   = color.New()
	DiffContextToShow    = 4
)

// ResourceDiff returns the results of diffing left and right as an pretty
// printed string. It will display all the lines of both the sequences
// that are being compared.
// When the left is different from right it will prepend a " - |" before
// the line.
// When the right is different from left it will prepend a " + |" before
// the line.
// When the right and left are equal it will prepend a "   |" before
// the line.
func ResourceDiff(left, right Object, scheme *runtime.Scheme) (string, bool, error) {
	leftLines, err := yamlLines(left, scheme)
	if err != nil {
		return "", false, err
	}
	rightLines, err := yamlLines(right, scheme)
	if err != nil {
		return "", false, err
	}

	diff := difflib.Diff(leftLines, rightLines)

	var sb strings.Builder
	inElipsis := false
	hasDiff := false

	for lineNum, record := range diff {
		switch record.Delta {
		case difflib.RightOnly:
			inElipsis = false
			hasDiff = true
			sb.WriteString(DiffAdditionColor.Sprintf("%3s %3d + |%s\n", "", record.LineRight+1, record.Payload))
		case difflib.LeftOnly:
			inElipsis = false
			hasDiff = true
			sb.WriteString(DiffSubtractionColor.Sprintf("%3d %3s - |%s\n", record.LineLeft+1, "", record.Payload))
		case difflib.Common:
			if !inContext(lineNum, diff) {
				if !inElipsis {
					sb.WriteString(DiffUnchangedColor.Sprintf("...\n"))
					inElipsis = true
				}
				continue
			}
			sb.WriteString(DiffUnchangedColor.Sprintf("%3d,%3d   |%s\n", record.LineLeft+1, record.LineRight+1, record.Payload))
		}
	}

	return sb.String(), !hasDiff, nil
}

func inContext(lineNum int, diff []difflib.DiffRecord) bool {
	start := max(0, lineNum-DiffContextToShow)
	end := min(len(diff), lineNum+DiffContextToShow+1)
	for _, record := range diff[start:end] {
		if record.Delta != difflib.Common {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func yamlLines(obj Object, scheme *runtime.Scheme) ([]string, error) {
	if obj == nil || reflect.ValueOf(obj).IsNil() {
		return []string{}, nil
	}
	yaml, err := ExportResource(obj, OutputFormatYaml, scheme)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(strings.TrimSpace(string(yaml)), "\n"), nil

}
