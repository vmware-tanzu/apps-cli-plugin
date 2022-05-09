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

// +die:object=true,spec=-
type _ = corev1.PodTemplate

func (d *PodTemplateDie) TemplateDie(fn func(d *PodTemplateSpecDie)) *PodTemplateDie {
	return d.DieStamp(func(r *corev1.PodTemplate) {
		d := PodTemplateSpecBlank.DieImmutable(false).DieFeed(r.Template)
		fn(d)
		r.Template = d.DieRelease()
	})
}

// +die
type _ = corev1.PodTemplateSpec

func (d *PodTemplateSpecDie) MetadataDie(fn func(d *diemetav1.ObjectMetaDie)) *PodTemplateSpecDie {
	return d.DieStamp(func(r *corev1.PodTemplateSpec) {
		d := diemetav1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

func (d *PodTemplateSpecDie) SpecDie(fn func(d *PodSpecDie)) *PodTemplateSpecDie {
	return d.DieStamp(func(r *corev1.PodTemplateSpec) {
		d := PodSpecBlank.DieImmutable(false).DieFeed(r.Spec)
		fn(d)
		r.Spec = d.DieRelease()
	})
}
