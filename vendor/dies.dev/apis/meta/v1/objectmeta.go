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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

// +die
type _ = metav1.ObjectMeta

func (d *ObjectMetaDie) AddLabel(key, value string) *ObjectMetaDie {
	return d.DieStamp(func(r *metav1.ObjectMeta) {
		if r.Labels == nil {
			r.Labels = map[string]string{}
		}
		r.Labels[key] = value
	})
}

func (d *ObjectMetaDie) AddAnnotation(key, value string) *ObjectMetaDie {
	return d.DieStamp(func(r *metav1.ObjectMeta) {
		if r.Annotations == nil {
			r.Annotations = map[string]string{}
		}
		r.Annotations[key] = value
	})
}

func (d *ObjectMetaDie) ControlledBy(obj runtime.Object, scheme *runtime.Scheme) *ObjectMetaDie {
	// create a copy to shed the die, if any
	obj = obj.DeepCopyObject()
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		panic(err)
	}
	return d.OwnerReferences(metav1.OwnerReference{
		APIVersion:         gvks[0].GroupVersion().String(),
		Kind:               gvks[0].Kind,
		Name:               obj.(metav1.Object).GetName(),
		UID:                obj.(metav1.Object).GetUID(),
		BlockOwnerDeletion: pointer.Bool(true),
		Controller:         pointer.Bool(true),
	})
}

func (d *ObjectMetaDie) ManagedFieldsDie(fields ...*ManagedFieldsEntryDie) *ObjectMetaDie {
	return d.DieStamp(func(r *metav1.ObjectMeta) {
		r.ManagedFields = make([]metav1.ManagedFieldsEntry, len(fields))
		for i := range fields {
			r.ManagedFields[i] = fields[i].DieRelease()
		}
	})
}

// +die
type _ = metav1.ManagedFieldsEntry

func FreezeObjectMeta(r metav1.ObjectMeta) FrozenObjectMeta {
	return FrozenObjectMeta{
		r: r,
	}
}

type FrozenObjectMeta struct {
	r metav1.ObjectMeta
}

// methods required to implement metav1.ObjectMetaAccessor

var _ metav1.ObjectMetaAccessor = (*FrozenObjectMeta)(nil)

func (d *FrozenObjectMeta) GetObjectMeta() metav1.Object {
	return d
}

// methods required to implement metav1.Object

var _ metav1.Object = (*FrozenObjectMeta)(nil)

func (d *FrozenObjectMeta) GetNamespace() string {
	return d.r.DeepCopy().GetNamespace()
}

func (d *FrozenObjectMeta) SetNamespace(namespace string) {
	panic(fmt.Errorf("SetNamespace() is not implemented"))
}

func (d *FrozenObjectMeta) GetName() string {
	return d.r.DeepCopy().GetName()
}

func (d *FrozenObjectMeta) SetName(name string) {
	panic(fmt.Errorf("SetName() is not implemented"))
}

func (d *FrozenObjectMeta) GetGenerateName() string {
	return d.r.DeepCopy().GetGenerateName()
}

func (d *FrozenObjectMeta) SetGenerateName(name string) {
	panic(fmt.Errorf("SetGenerateName() is not implemented"))
}

func (d *FrozenObjectMeta) GetUID() types.UID {
	return d.r.DeepCopy().GetUID()
}

func (d *FrozenObjectMeta) SetUID(uid types.UID) {
	panic(fmt.Errorf("SetUID() is not implemented"))
}

func (d *FrozenObjectMeta) GetResourceVersion() string {
	return d.r.DeepCopy().GetResourceVersion()
}

func (d *FrozenObjectMeta) SetResourceVersion(version string) {
	panic(fmt.Errorf("SetResourceVersion() is not implemented"))
}

func (d *FrozenObjectMeta) GetGeneration() int64 {
	return d.r.DeepCopy().GetGeneration()
}

func (d *FrozenObjectMeta) SetGeneration(generation int64) {
	panic(fmt.Errorf("SetGeneration() is not implemented"))
}

func (d *FrozenObjectMeta) GetSelfLink() string {
	return d.r.DeepCopy().GetSelfLink()
}

func (d *FrozenObjectMeta) SetSelfLink(selfLink string) {
	panic(fmt.Errorf("SetSelfLink() is not implemented"))
}

func (d *FrozenObjectMeta) GetCreationTimestamp() metav1.Time {
	return d.r.DeepCopy().GetCreationTimestamp()
}

func (d *FrozenObjectMeta) SetCreationTimestamp(timestamp metav1.Time) {
	panic(fmt.Errorf("SetCreationTimestamp() is not implemented"))
}

func (d *FrozenObjectMeta) GetDeletionTimestamp() *metav1.Time {
	return d.r.DeepCopy().GetDeletionTimestamp()
}

func (d *FrozenObjectMeta) SetDeletionTimestamp(timestamp *metav1.Time) {
	panic(fmt.Errorf("SetDeletionTimestamp() is not implemented"))
}

func (d *FrozenObjectMeta) GetDeletionGracePeriodSeconds() *int64 {
	return d.r.GetDeletionGracePeriodSeconds()
}

func (d *FrozenObjectMeta) SetDeletionGracePeriodSeconds(*int64) {
	panic(fmt.Errorf("SetDeletionGracePeriodSeconds() is not implemented"))
}

func (d *FrozenObjectMeta) GetLabels() map[string]string {
	return d.r.DeepCopy().GetLabels()
}

func (d *FrozenObjectMeta) SetLabels(labels map[string]string) {
	panic(fmt.Errorf("SetLabels() is not implemented"))
}

func (d *FrozenObjectMeta) GetAnnotations() map[string]string {
	return d.r.DeepCopy().GetAnnotations()
}

func (d *FrozenObjectMeta) SetAnnotations(annotations map[string]string) {
	panic(fmt.Errorf("SetAnnotations() is not implemented"))
}

func (d *FrozenObjectMeta) GetFinalizers() []string {
	return d.r.DeepCopy().GetFinalizers()
}

func (d *FrozenObjectMeta) SetFinalizers(finalizers []string) {
	panic(fmt.Errorf("SetFinalizers() is not implemented"))
}

func (d *FrozenObjectMeta) GetOwnerReferences() []metav1.OwnerReference {
	return d.r.DeepCopy().GetOwnerReferences()
}

func (d *FrozenObjectMeta) SetOwnerReferences([]metav1.OwnerReference) {
	panic(fmt.Errorf("SetOwnerReferences() is not implemented"))
}

func (d *FrozenObjectMeta) GetClusterName() string {
	return d.r.DeepCopy().GetClusterName()
}

func (d *FrozenObjectMeta) SetClusterName(clusterName string) {
	panic(fmt.Errorf("SetClusterName() is not implemented"))
}

func (d *FrozenObjectMeta) GetManagedFields() []metav1.ManagedFieldsEntry {
	return d.r.DeepCopy().GetManagedFields()
}

func (d *FrozenObjectMeta) SetManagedFields(managedFields []metav1.ManagedFieldsEntry) {
	panic(fmt.Errorf("SetManagedFields() is not implemented"))
}
