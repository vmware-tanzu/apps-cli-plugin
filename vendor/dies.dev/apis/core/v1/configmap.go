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
	corev1 "k8s.io/api/core/v1"
)

// +die:object=true,ignore={BinaryData,Data}
type _ = corev1.ConfigMap

func (d *ConfigMapDie) Data(v map[string]string) *ConfigMapDie {
	return d.DieStamp(func(r *corev1.ConfigMap) {
		for k := range v {
			delete(r.BinaryData, k)
		}
		r.Data = v
	})
}

func (d *ConfigMapDie) AddData(key, value string) *ConfigMapDie {
	return d.DieStamp(func(r *corev1.ConfigMap) {
		if r.Data == nil {
			r.Data = map[string]string{}
		}
		delete(d.r.BinaryData, key)
		r.Data[key] = value
	})
}

func (d *ConfigMapDie) BinaryData(v map[string][]byte) *ConfigMapDie {
	return d.DieStamp(func(r *corev1.ConfigMap) {
		for k := range v {
			delete(r.Data, k)
		}
		r.BinaryData = v
	})
}

func (d *ConfigMapDie) AddBinaryData(key, value string) *ConfigMapDie {
	return d.DieStamp(func(r *corev1.ConfigMap) {
		if r.BinaryData == nil {
			r.BinaryData = map[string][]byte{}
		}
		delete(d.r.Data, key)
		r.Data[key] = value
	})
}
