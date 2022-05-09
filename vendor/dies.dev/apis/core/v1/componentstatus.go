/*
Copyright 2022 the original author or authors.

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
	diemetav1 "dies.dev/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

// +die:object=true
type _ = corev1.ComponentStatus

func (d *ComponentStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *ComponentStatusDie {
	return d.DieStamp(func(r *corev1.ComponentStatus) {
		r.Conditions = make([]corev1.ComponentCondition, len(conditions))
		for i := range conditions {
			c := conditions[i].DieRelease()
			r.Conditions[i] = corev1.ComponentCondition{
				Type:    corev1.ComponentConditionType(c.Type),
				Status:  corev1.ConditionStatus(c.Status),
				Message: c.Message,
			}
		}
	})
}
