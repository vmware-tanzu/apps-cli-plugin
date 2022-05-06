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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die
type _ = metav1.Condition

func (d *ConditionDie) True() *ConditionDie {
	return d.DieStamp(func(r *metav1.Condition) {
		r.Status = metav1.ConditionTrue
	})
}

func (d *ConditionDie) False() *ConditionDie {
	return d.DieStamp(func(r *metav1.Condition) {
		r.Status = metav1.ConditionFalse
	})
}

func (d *ConditionDie) Unknown() *ConditionDie {
	return d.DieStamp(func(r *metav1.Condition) {
		r.Status = metav1.ConditionUnknown
	})
}

func (d *ConditionDie) Messagef(format string, a ...interface{}) *ConditionDie {
	return d.DieStamp(func(r *metav1.Condition) {
		r.Message = fmt.Sprintf(format, a...)
	})
}
