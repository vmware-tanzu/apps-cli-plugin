/*
Copyright 2019 VMware, Inc.

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

package parsers

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func ObjectReference(str string) corev1.ObjectReference {
	parts := strings.SplitN(str, ":", 4)
	var name = parts[2]
	if len(parts) == 4 {
		name = parts[3]
	}
	return corev1.ObjectReference{
		APIVersion: parts[0],
		Kind:       parts[1],
		Name:       name,
	}
}

func DeletableObjectReference(str string) (corev1.ObjectReference, bool) {
	remove := false
	if strings.HasSuffix(str, "-") {
		str = str[0 : len(str)-1]
		remove = true
	}
	return ObjectReference(str), remove
}

func ObjectReferenceAnnotation(str string) map[string]string {
	parts := strings.SplitN(str, ":", 4)
	if len(parts) == 4 {
		return map[string]string{"namespace": parts[2]}
	}
	return nil
}
