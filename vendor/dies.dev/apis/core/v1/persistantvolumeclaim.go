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
type _ = corev1.PersistentVolumeClaim

// +die
type _ = corev1.PersistentVolumeClaimSpec

func (d *PersistentVolumeClaimSpecDie) SelectorDie(fn func(d *diemetav1.LabelSelectorDie)) *PersistentVolumeClaimSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimSpec) {
		d := diemetav1.LabelSelectorBlank.DieImmutable(false).DieFeedPtr(r.Selector)
		fn(d)
		r.Selector = d.DieReleasePtr()
	})
}

func (d *PersistentVolumeClaimSpecDie) ResourcesDie(fn func(d *VolumeResourceRequirementsDie)) *PersistentVolumeClaimSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimSpec) {
		d := VolumeResourceRequirementsBlank.DieImmutable(false).DieFeed(r.Resources)
		fn(d)
		r.Resources = d.DieRelease()
	})
}

func (d *PersistentVolumeClaimSpecDie) DataSourceDie(fn func(d *TypedLocalObjectReferenceDie)) *PersistentVolumeClaimSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimSpec) {
		d := TypedLocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.DataSource)
		fn(d)
		r.DataSource = d.DieReleasePtr()
	})
}

func (d *PersistentVolumeClaimSpecDie) DataSourceRefDie(fn func(d *TypedObjectReferenceDie)) *PersistentVolumeClaimSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimSpec) {
		d := TypedObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.DataSourceRef)
		fn(d)
		r.DataSourceRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.VolumeResourceRequirements

func (d *VolumeResourceRequirementsDie) AddLimit(name corev1.ResourceName, quantity resource.Quantity) *VolumeResourceRequirementsDie {
	return d.DieStamp(func(r *corev1.VolumeResourceRequirements) {
		if r.Limits == nil {
			r.Limits = corev1.ResourceList{}
		}
		r.Limits[name] = quantity
	})
}

func (d *VolumeResourceRequirementsDie) AddLimitString(name corev1.ResourceName, quantity string) *VolumeResourceRequirementsDie {
	return d.AddLimit(name, resource.MustParse(quantity))
}

func (d *VolumeResourceRequirementsDie) AddRequest(name corev1.ResourceName, quantity resource.Quantity) *VolumeResourceRequirementsDie {
	return d.DieStamp(func(r *corev1.VolumeResourceRequirements) {
		if r.Requests == nil {
			r.Requests = corev1.ResourceList{}
		}
		r.Requests[name] = quantity
	})
}

func (d *VolumeResourceRequirementsDie) AddRequestString(name corev1.ResourceName, quantity string) *VolumeResourceRequirementsDie {
	return d.AddRequest(name, resource.MustParse(quantity))
}

// +die:ignore={AllocatedResourceStatuses}
type _ = corev1.PersistentVolumeClaimStatus

func (d *PersistentVolumeClaimStatusDie) AddCapacity(name corev1.ResourceName, quantity resource.Quantity) *PersistentVolumeClaimStatusDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimStatus) {
		if r.Capacity == nil {
			r.Capacity = corev1.ResourceList{}
		}
		r.Capacity[name] = quantity
	})
}

func (d *PersistentVolumeClaimStatusDie) AddCapacityString(name corev1.ResourceName, quantity string) *PersistentVolumeClaimStatusDie {
	return d.AddCapacity(name, resource.MustParse(quantity))
}

func (d *PersistentVolumeClaimStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *PersistentVolumeClaimStatusDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimStatus) {
		r.Conditions = make([]corev1.PersistentVolumeClaimCondition, len(conditions))
		for i := range conditions {
			c := conditions[i].DieRelease()
			r.Conditions[i] = corev1.PersistentVolumeClaimCondition{
				Type:               corev1.PersistentVolumeClaimConditionType(c.Type),
				Status:             corev1.ConditionStatus(c.Status),
				Reason:             c.Reason,
				Message:            c.Message,
				LastTransitionTime: c.LastTransitionTime,
			}
		}
	})
}

func (d *PersistentVolumeClaimStatusDie) AddAllocatedResources(name corev1.ResourceName, quantity resource.Quantity) *PersistentVolumeClaimStatusDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimStatus) {
		if r.AllocatedResources == nil {
			r.AllocatedResources = corev1.ResourceList{}
		}
		r.AllocatedResources[name] = quantity
	})
}

func (d *PersistentVolumeClaimStatusDie) AddAllocatedResourcesString(name corev1.ResourceName, quantity string) *PersistentVolumeClaimStatusDie {
	return d.AddAllocatedResources(name, resource.MustParse(quantity))
}

// allocatedResourceStatuses stores status of resource being resized for the given PVC.
// Key names follow standard Kubernetes label syntax. Valid values are either:
//   - Un-prefixed keys:
//   - storage - the capacity of the volume.
//   - Custom resources must use implementation-defined prefixed names such as "example.com/my-custom-resource"
//
// Apart from above values - keys that are unprefixed or have kubernetes.io prefix are considered
// reserved and hence may not be used.
//
// ClaimResourceStatus can be in any of following states:
//   - ControllerResizeInProgress:
//     State set when resize controller starts resizing the volume in control-plane.
//   - ControllerResizeFailed:
//     State set when resize has failed in resize controller with a terminal error.
//   - NodeResizePending:
//     State set when resize controller has finished resizing the volume but further resizing of
//     volume is needed on the node.
//   - NodeResizeInProgress:
//     State set when kubelet starts resizing the volume.
//   - NodeResizeFailed:
//     State set when resizing has failed in kubelet with a terminal error. Transient errors don't set
//     NodeResizeFailed.
//
// For example: if expanding a PVC for more capacity - this field can be one of the following states:
//   - pvc.status.allocatedResourceStatus['storage'] = "ControllerResizeInProgress"
//   - pvc.status.allocatedResourceStatus['storage'] = "ControllerResizeFailed"
//   - pvc.status.allocatedResourceStatus['storage'] = "NodeResizePending"
//   - pvc.status.allocatedResourceStatus['storage'] = "NodeResizeInProgress"
//   - pvc.status.allocatedResourceStatus['storage'] = "NodeResizeFailed"
//
// When this field is not set, it means that no resize operation is in progress for the given PVC.
//
// A controller that receives PVC update with previously unknown resourceName or ClaimResourceStatus
// should ignore the update for the purpose it was designed. For example - a controller that
// only is responsible for resizing capacity of the volume, should ignore PVC updates that change other valid
// resources associated with PVC.
//
// This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
func (d *PersistentVolumeClaimStatusDie) AllocatedResourceStatuses(v map[corev1.ResourceName]corev1.ClaimResourceStatus) *PersistentVolumeClaimStatusDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimStatus) {
		r.AllocatedResourceStatuses = v
	})
}

func (d *PersistentVolumeClaimStatusDie) AddAllocatedResourceStatus(name corev1.ResourceName, status corev1.ClaimResourceStatus) *PersistentVolumeClaimStatusDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimStatus) {
		if r.AllocatedResourceStatuses == nil {
			r.AllocatedResourceStatuses = map[corev1.ResourceName]corev1.ClaimResourceStatus{}
		}
		r.AllocatedResourceStatuses[name] = status
	})
}

func (d *PersistentVolumeClaimStatusDie) ModifyVolumeStatusDie(fn func(d *ModifyVolumeStatusDie)) *PersistentVolumeClaimStatusDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimStatus) {
		d := ModifyVolumeStatusBlank.DieImmutable(false).DieFeedPtr(r.ModifyVolumeStatus)
		fn(d)
		r.ModifyVolumeStatus = d.DieReleasePtr()
	})
}

// +die
type _ corev1.ModifyVolumeStatus

// +die
type _ corev1.PersistentVolumeClaimTemplate

func (d *PersistentVolumeClaimTemplateDie) MetadataDie(fn func(d *diemetav1.ObjectMetaDie)) *PersistentVolumeClaimTemplateDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimTemplate) {
		d := diemetav1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

func (d *PersistentVolumeClaimTemplateDie) SpecDie(fn func(d *PersistentVolumeClaimSpecDie)) *PersistentVolumeClaimTemplateDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeClaimTemplate) {
		d := PersistentVolumeClaimSpecBlank.DieImmutable(false).DieFeed(r.Spec)
		fn(d)
		r.Spec = d.DieRelease()
	})
}
