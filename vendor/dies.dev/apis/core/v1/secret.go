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

// +die:object=true,ignore={Data,StringData}
type Secret = corev1.Secret

func (d *SecretDie) Data(v map[string][]byte) *SecretDie {
	return d.DieStamp(func(r *corev1.Secret) {
		for k := range v {
			delete(r.StringData, k)
		}
		r.Data = v
	})
}

func (d *SecretDie) AddData(key string, value []byte) *SecretDie {
	return d.DieStamp(func(r *corev1.Secret) {
		if r.Data == nil {
			r.Data = map[string][]byte{}
		}
		delete(d.r.StringData, key)
		r.Data[key] = value
	})
}

func (d *SecretDie) StringData(v map[string]string) *SecretDie {
	return d.DieStamp(func(r *corev1.Secret) {
		for k := range v {
			delete(r.Data, k)
		}
		r.StringData = v
	})
}

func (d *SecretDie) AddStringData(key string, value string) *SecretDie {
	return d.DieStamp(func(r *corev1.Secret) {
		if r.StringData == nil {
			r.StringData = map[string]string{}
		}
		delete(d.r.StringData, key)
		r.StringData[key] = value
	})
}
