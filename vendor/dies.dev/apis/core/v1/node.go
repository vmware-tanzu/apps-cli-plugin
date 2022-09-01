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
	diemetav1 "dies.dev/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// +die:object=true
type _ = corev1.Node

// +die
type _ = corev1.NodeSpec

func (d *NodeSpecDie) TaintDie(key string, fn func(d *TaintDie)) *NodeSpecDie {
	return d.DieStamp(func(r *corev1.NodeSpec) {
		for i := range r.Taints {
			if key == r.Taints[i].Key {
				d := TaintBlank.DieImmutable(false).DieFeed(r.Taints[i])
				fn(d)
				r.Taints[i] = d.DieRelease()
				return
			}
		}

		d := TaintBlank.DieImmutable(false).DieFeed(corev1.Taint{Key: key})
		fn(d)
		r.Taints = append(r.Taints, d.DieRelease())
	})
}

func (d *NodeSpecDie) ConfigSourceDie(fn func(d *NodeConfigSourceDie)) *NodeSpecDie {
	return d.DieStamp(func(r *corev1.NodeSpec) {
		d := NodeConfigSourceBlank.DieImmutable(false).DieFeedPtr(r.ConfigSource)
		fn(d)
		r.ConfigSource = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.Taint

// +die
type _ = corev1.NodeConfigSource

func (d *NodeConfigSourceDie) ConfigMapDie(fn func(d *ConfigMapNodeConfigSourceDie)) *NodeConfigSourceDie {
	return d.DieStamp(func(r *corev1.NodeConfigSource) {
		d := ConfigMapNodeConfigSourceBlank.DieImmutable(false).DieFeedPtr(r.ConfigMap)
		fn(d)
		r.ConfigMap = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ConfigMapNodeConfigSource

// +die
type _ = corev1.NodeStatus

func (d *NodeStatusDie) AddCapacity(name corev1.ResourceName, quantity resource.Quantity) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		if r.Capacity == nil {
			r.Capacity = corev1.ResourceList{}
		}
		r.Capacity[name] = quantity
	})
}

func (d *NodeStatusDie) AddCapacityString(name corev1.ResourceName, quantity string) *NodeStatusDie {
	return d.AddCapacity(name, resource.MustParse(quantity))
}

func (d *NodeStatusDie) AddAllocatable(name corev1.ResourceName, quantity resource.Quantity) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		if r.Allocatable == nil {
			r.Allocatable = corev1.ResourceList{}
		}
		r.Allocatable[name] = quantity
	})
}

func (d *NodeStatusDie) AddAllocatableString(name corev1.ResourceName, quantity string) *NodeStatusDie {
	return d.AddAllocatable(name, resource.MustParse(quantity))
}

func (d *NodeStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		r.Conditions = make([]corev1.NodeCondition, len(conditions))
		for i := range conditions {
			c := conditions[i].DieRelease()
			r.Conditions[i] = corev1.NodeCondition{
				Type:               corev1.NodeConditionType(c.Type),
				Status:             corev1.ConditionStatus(c.Status),
				Reason:             c.Reason,
				Message:            c.Message,
				LastTransitionTime: c.LastTransitionTime,
			}
		}
	})
}

func (d *NodeStatusDie) AddresssDie(addresses ...*NodeAddressDie) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		r.Addresses = make([]corev1.NodeAddress, len(addresses))
		for i := range addresses {
			r.Addresses[i] = addresses[i].DieRelease()
		}
	})
}

func (d *NodeStatusDie) DaemonEndpointsDie(fn func(d *NodeDaemonEndpointsDie)) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		d := NodeDaemonEndpointsBlank.DieImmutable(false).DieFeed(r.DaemonEndpoints)
		fn(d)
		r.DaemonEndpoints = d.DieRelease()
	})
}

func (d *NodeStatusDie) NodeInfoDie(fn func(d *NodeSystemInfoDie)) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		d := NodeSystemInfoBlank.DieImmutable(false).DieFeed(r.NodeInfo)
		fn(d)
		r.NodeInfo = d.DieRelease()
	})
}

func (d *NodeStatusDie) ImagesDie(images ...*ContainerImageDie) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		r.Images = make([]corev1.ContainerImage, len(images))
		for i := range images {
			r.Images[i] = images[i].DieRelease()
		}
	})
}

func (d *NodeStatusDie) VolumeAttachedDie(name corev1.UniqueVolumeName, fn func(d *AttachedVolumeDie)) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		for i := range r.VolumesAttached {
			if name == r.VolumesAttached[i].Name {
				d := AttachedVolumeBlank.DieImmutable(false).DieFeed(r.VolumesAttached[i])
				fn(d)
				r.VolumesAttached[i] = d.DieRelease()
				return
			}
		}

		d := AttachedVolumeBlank.DieImmutable(false).DieFeed(corev1.AttachedVolume{Name: name})
		fn(d)
		r.VolumesAttached = append(r.VolumesAttached, d.DieRelease())
	})
}

func (d *NodeStatusDie) ConfigDie(fn func(d *NodeConfigStatusDie)) *NodeStatusDie {
	return d.DieStamp(func(r *corev1.NodeStatus) {
		d := NodeConfigStatusBlank.DieImmutable(false).DieFeedPtr(r.Config)
		fn(d)
		r.Config = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.NodeAddress

// +die
type _ = corev1.NodeDaemonEndpoints

func (d *NodeDaemonEndpointsDie) KubeletEndpointDie(fn func(d *DaemonEndpointDie)) *NodeDaemonEndpointsDie {
	return d.DieStamp(func(r *corev1.NodeDaemonEndpoints) {
		d := DaemonEndpointBlank.DieImmutable(false).DieFeed(r.KubeletEndpoint)
		fn(d)
		r.KubeletEndpoint = d.DieRelease()
	})
}

// +die
type _ = corev1.DaemonEndpoint

// +die
type _ = corev1.NodeSystemInfo

// +die
type _ = corev1.ContainerImage

// +die
type _ = corev1.AttachedVolume

// +die
type _ = corev1.NodeConfigStatus

func (d *NodeConfigStatusDie) AssignedDie(fn func(d *NodeConfigSourceDie)) *NodeConfigStatusDie {
	return d.DieStamp(func(r *corev1.NodeConfigStatus) {
		d := NodeConfigSourceBlank.DieImmutable(false).DieFeedPtr(r.Assigned)
		fn(d)
		r.Assigned = d.DieReleasePtr()
	})
}

func (d *NodeConfigStatusDie) ActiveDie(fn func(d *NodeConfigSourceDie)) *NodeConfigStatusDie {
	return d.DieStamp(func(r *corev1.NodeConfigStatus) {
		d := NodeConfigSourceBlank.DieImmutable(false).DieFeedPtr(r.Active)
		fn(d)
		r.Active = d.DieReleasePtr()
	})
}

func (d *NodeConfigStatusDie) LastKnownGoodDie(fn func(d *NodeConfigSourceDie)) *NodeConfigStatusDie {
	return d.DieStamp(func(r *corev1.NodeConfigStatus) {
		d := NodeConfigSourceBlank.DieImmutable(false).DieFeedPtr(r.LastKnownGood)
		fn(d)
		r.LastKnownGood = d.DieReleasePtr()
	})
}
