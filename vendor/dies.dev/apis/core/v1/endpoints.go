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
type _ = corev1.Endpoints

func (d *EndpointsDie) SubsetsDie(subsets ...*EndpointSubsetDie) *EndpointsDie {
	return d.DieStamp(func(r *corev1.Endpoints) {
		r.Subsets = make([]corev1.EndpointSubset, len(subsets))
		for i := range subsets {
			r.Subsets[i] = subsets[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.EndpointSubset

func (d *EndpointSubsetDie) AddressesDie(addresses ...*EndpointAddressDie) *EndpointSubsetDie {
	return d.DieStamp(func(r *corev1.EndpointSubset) {
		r.Addresses = make([]corev1.EndpointAddress, len(addresses))
		for i := range addresses {
			r.Addresses[i] = addresses[i].DieRelease()
		}
	})
}

func (d *EndpointSubsetDie) NotReadyAddressesDie(addresses ...*EndpointAddressDie) *EndpointSubsetDie {
	return d.DieStamp(func(r *corev1.EndpointSubset) {
		r.NotReadyAddresses = make([]corev1.EndpointAddress, len(addresses))
		for i := range addresses {
			r.NotReadyAddresses[i] = addresses[i].DieRelease()
		}
	})
}

func (d *EndpointSubsetDie) PortsDie(ports ...*EndpointPortDie) *EndpointSubsetDie {
	return d.DieStamp(func(r *corev1.EndpointSubset) {
		r.Ports = make([]corev1.EndpointPort, len(ports))
		for i := range ports {
			r.Ports[i] = ports[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.EndpointAddress

func (d *EndpointAddressDie) TargetRefDie(fn func(d *ObjectReferenceDie)) *EndpointAddressDie {
	return d.DieStamp(func(r *corev1.EndpointAddress) {
		d := ObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.TargetRef)
		fn(d)
		r.TargetRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.EndpointPort
