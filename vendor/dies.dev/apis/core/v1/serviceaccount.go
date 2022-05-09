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

// +die:object=true
type _ = corev1.ServiceAccount

func (d *ServiceAccountDie) SecretsDie(secrets ...*ObjectReferenceDie) *ServiceAccountDie {
	return d.DieStamp(func(r *corev1.ServiceAccount) {
		r.Secrets = make([]corev1.ObjectReference, len(secrets))
		for i := range secrets {
			r.Secrets[i] = secrets[i].DieRelease()
		}
	})
}

func (d *ServiceAccountDie) ImagePullSecretsDie(secrets ...*LocalObjectReferenceDie) *ServiceAccountDie {
	return d.DieStamp(func(r *corev1.ServiceAccount) {
		r.ImagePullSecrets = make([]corev1.LocalObjectReference, len(secrets))
		for i := range secrets {
			r.ImagePullSecrets[i] = secrets[i].DieRelease()
		}
	})
}
