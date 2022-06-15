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
	json "encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var TypeMetaBlank = (&TypeMetaDie{}).DieFeed(metav1.TypeMeta{})

type TypeMetaDie struct {
	mutable bool
	r       metav1.TypeMeta
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *TypeMetaDie) DieImmutable(immutable bool) *TypeMetaDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *TypeMetaDie) DieFeed(r metav1.TypeMeta) *TypeMetaDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &TypeMetaDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *TypeMetaDie) DieFeedPtr(r *metav1.TypeMeta) *TypeMetaDie {
	if r == nil {
		r = &metav1.TypeMeta{}
	}
	return d.DieFeed(*r)
}

// DieFeedRawExtension returns the resource managed by the die as an raw extension.
func (d *TypeMetaDie) DieFeedRawExtension(raw runtime.RawExtension) *TypeMetaDie {
	b, _ := json.Marshal(raw)
	r := metav1.TypeMeta{}
	_ = json.Unmarshal(b, &r)
	return d.DieFeed(r)
}

// DieRelease returns the resource managed by the die.
func (d *TypeMetaDie) DieRelease() metav1.TypeMeta {
	if d.mutable {
		return d.r
	}
	return metav1.TypeMeta{
		APIVersion: d.r.APIVersion,
		Kind:       d.r.Kind,
	}
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *TypeMetaDie) DieReleasePtr() *metav1.TypeMeta {
	r := d.DieRelease()
	return &r
}

// DieReleaseRawExtension returns the resource managed by the die as an raw extension.
func (d *TypeMetaDie) DieReleaseRawExtension() runtime.RawExtension {
	r := d.DieReleasePtr()
	b, _ := json.Marshal(r)
	raw := runtime.RawExtension{}
	_ = json.Unmarshal(b, &raw)
	return raw
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *TypeMetaDie) DieStamp(fn func(r *metav1.TypeMeta)) *TypeMetaDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *TypeMetaDie) DeepCopy() *TypeMetaDie {
	return &TypeMetaDie{
		mutable: d.mutable,
		r: metav1.TypeMeta{
			APIVersion: d.r.APIVersion,
			Kind:       d.r.Kind,
		},
	}
}

// Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
func (d *TypeMetaDie) Kind(v string) *TypeMetaDie {
	return d.DieStamp(func(r *metav1.TypeMeta) {
		r.Kind = v
	})
}

// APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
func (d *TypeMetaDie) APIVersion(v string) *TypeMetaDie {
	return d.DieStamp(func(r *metav1.TypeMeta) {
		r.APIVersion = v
	})
}
