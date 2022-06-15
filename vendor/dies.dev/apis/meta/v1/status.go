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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die
type _ = metav1.Status

// Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
func (d *StatusDie) Kind(v string) *StatusDie {
	return d.DieStamp(func(r *metav1.Status) {
		r.Kind = v
	})
}

// APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
func (d *StatusDie) APIVersion(v string) *StatusDie {
	return d.DieStamp(func(r *metav1.Status) {
		r.APIVersion = v
	})
}

func (d *StatusDie) ListMetaDie(fn func(d *ListMetaDie)) *StatusDie {
	return d.DieStamp(func(r *metav1.Status) {
		d := ListMetaBlank.DieImmutable(false).DieFeed(r.ListMeta)
		fn(d)
		r.ListMeta = d.DieRelease()
	})
}

func (d *StatusDie) DetailDie(fn func(d *StatusDetailsDie)) *StatusDie {
	return d.DieStamp(func(r *metav1.Status) {
		d := StatusDetailsBlank.DieImmutable(false).DieFeedPtr(r.Details)
		fn(d)
		r.Details = d.DieReleasePtr()
	})
}

// +die
type _ = metav1.StatusDetails

func (d *StatusDetailsDie) CausesDie(causes ...*StatusCauseDie) *StatusDetailsDie {
	return d.DieStamp(func(r *metav1.StatusDetails) {
		r.Causes = make([]metav1.StatusCause, len(causes))
		for i := range causes {
			c := causes[i].DieRelease()
			r.Causes[i] = c
		}
	})
}

// +die
type _ = metav1.StatusCause
