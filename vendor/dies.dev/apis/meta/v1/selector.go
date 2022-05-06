/*
Copyright 2021 the original author or authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die
type _ = metav1.LabelSelector

func (d *LabelSelectorDie) AddMatchLabel(key, value string) *LabelSelectorDie {
	return d.DieStamp(func(r *metav1.LabelSelector) {
		if r.MatchLabels == nil {
			r.MatchLabels = map[string]string{}
		}
		r.MatchLabels[key] = value
	})
}

func (d *LabelSelectorDie) AddMatchExpression(key string, operator metav1.LabelSelectorOperator, values ...string) *LabelSelectorDie {
	return d.DieStamp(func(r *metav1.LabelSelector) {
		lsr := metav1.LabelSelectorRequirement{
			Key:      key,
			Operator: operator,
			Values:   values,
		}

		found := false
		for i := range r.MatchExpressions {
			if lsr.Key == r.MatchExpressions[i].Key {
				found = true
				r.MatchExpressions[i] = lsr
			}
		}
		if !found {
			r.MatchExpressions = append(r.MatchExpressions, lsr)
		}
	})
}
