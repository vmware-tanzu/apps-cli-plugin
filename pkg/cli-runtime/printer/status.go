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

package printer

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func ResourceStatus(name string, condition *metav1.Condition) string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("# %s: %s\n", name, ConditionStatus(condition)))
	if condition != nil {
		s, _ := yaml.Marshal(condition)
		b.WriteString("---\n")
		b.Write(s)
	}
	return b.String()
}

func FindCondition(conditions []metav1.Condition, ct string) *metav1.Condition {
	for _, c := range conditions {
		if c.Type == ct {
			return &c
		}
	}
	return nil
}
