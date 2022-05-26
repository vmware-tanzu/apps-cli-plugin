//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 VMware, Inc.

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

// Code generated by diegen. DO NOT EDIT.

package v1alpha1

import (
	json "encoding/json"
	fmtx "fmt"

	v1 "dies.dev/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"

	cartographerv1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
)

var ClusterSupplyChainBlank = (&ClusterSupplyChainDie{}).DieFeed(cartographerv1alpha1.ClusterSupplyChain{})

type ClusterSupplyChainDie struct {
	v1.FrozenObjectMeta
	mutable bool
	r       cartographerv1alpha1.ClusterSupplyChain
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *ClusterSupplyChainDie) DieImmutable(immutable bool) *ClusterSupplyChainDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *ClusterSupplyChainDie) DieFeed(r cartographerv1alpha1.ClusterSupplyChain) *ClusterSupplyChainDie {
	if d.mutable {
		d.FrozenObjectMeta = v1.FreezeObjectMeta(r.ObjectMeta)
		d.r = r
		return d
	}
	return &ClusterSupplyChainDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *ClusterSupplyChainDie) DieFeedPtr(r *cartographerv1alpha1.ClusterSupplyChain) *ClusterSupplyChainDie {
	if r == nil {
		r = &cartographerv1alpha1.ClusterSupplyChain{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *ClusterSupplyChainDie) DieRelease() cartographerv1alpha1.ClusterSupplyChain {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *ClusterSupplyChainDie) DieReleasePtr() *cartographerv1alpha1.ClusterSupplyChain {
	r := d.DieRelease()
	return &r
}

// DieReleaseUnstructured returns the resource managed by the die as an unstructured object.
func (d *ClusterSupplyChainDie) DieReleaseUnstructured() runtime.Unstructured {
	r := d.DieReleasePtr()
	u, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	return &unstructured.Unstructured{
		Object: u,
	}
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *ClusterSupplyChainDie) DieStamp(fn func(r *cartographerv1alpha1.ClusterSupplyChain)) *ClusterSupplyChainDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *ClusterSupplyChainDie) DeepCopy() *ClusterSupplyChainDie {
	r := *d.r.DeepCopy()
	return &ClusterSupplyChainDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

var _ runtime.Object = (*ClusterSupplyChainDie)(nil)

func (d *ClusterSupplyChainDie) DeepCopyObject() runtime.Object {
	return d.r.DeepCopy()
}

func (d *ClusterSupplyChainDie) GetObjectKind() schema.ObjectKind {
	r := d.DieRelease()
	return r.GetObjectKind()
}

func (d *ClusterSupplyChainDie) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.r)
}

func (d *ClusterSupplyChainDie) UnmarshalJSON(b []byte) error {
	if d == ClusterSupplyChainBlank {
		return fmtx.Errorf("cannot unmarshal into the blank die, create a copy first")
	}
	if !d.mutable {
		return fmtx.Errorf("cannot unmarshal into immutable dies, create a mutable version first")
	}
	r := &cartographerv1alpha1.ClusterSupplyChain{}
	err := json.Unmarshal(b, r)
	*d = *d.DieFeed(*r)
	return err
}

// MetadataDie stamps the resource's ObjectMeta field with a mutable die.
func (d *ClusterSupplyChainDie) MetadataDie(fn func(d *v1.ObjectMetaDie)) *ClusterSupplyChainDie {
	return d.DieStamp(func(r *cartographerv1alpha1.ClusterSupplyChain) {
		d := v1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

// SpecDie stamps the resource's spec field with a mutable die.
func (d *ClusterSupplyChainDie) SpecDie(fn func(d *SupplyChainSpecDie)) *ClusterSupplyChainDie {
	return d.DieStamp(func(r *cartographerv1alpha1.ClusterSupplyChain) {
		d := SupplyChainSpecBlank.DieImmutable(false).DieFeed(r.Spec)
		fn(d)
		r.Spec = d.DieRelease()
	})
}

// StatusDie stamps the resource's status field with a mutable die.
func (d *ClusterSupplyChainDie) StatusDie(fn func(d *SupplyChainStatusDie)) *ClusterSupplyChainDie {
	return d.DieStamp(func(r *cartographerv1alpha1.ClusterSupplyChain) {
		d := SupplyChainStatusBlank.DieImmutable(false).DieFeed(r.Status)
		fn(d)
		r.Status = d.DieRelease()
	})
}

func (d *ClusterSupplyChainDie) Spec(v cartographerv1alpha1.SupplyChainSpec) *ClusterSupplyChainDie {
	return d.DieStamp(func(r *cartographerv1alpha1.ClusterSupplyChain) {
		r.Spec = v
	})
}

func (d *ClusterSupplyChainDie) Status(v cartographerv1alpha1.SupplyChainStatus) *ClusterSupplyChainDie {
	return d.DieStamp(func(r *cartographerv1alpha1.ClusterSupplyChain) {
		r.Status = v
	})
}

var SupplyChainSpecBlank = (&SupplyChainSpecDie{}).DieFeed(cartographerv1alpha1.SupplyChainSpec{})

type SupplyChainSpecDie struct {
	mutable bool
	r       cartographerv1alpha1.SupplyChainSpec
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *SupplyChainSpecDie) DieImmutable(immutable bool) *SupplyChainSpecDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *SupplyChainSpecDie) DieFeed(r cartographerv1alpha1.SupplyChainSpec) *SupplyChainSpecDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &SupplyChainSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *SupplyChainSpecDie) DieFeedPtr(r *cartographerv1alpha1.SupplyChainSpec) *SupplyChainSpecDie {
	if r == nil {
		r = &cartographerv1alpha1.SupplyChainSpec{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *SupplyChainSpecDie) DieRelease() cartographerv1alpha1.SupplyChainSpec {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *SupplyChainSpecDie) DieReleasePtr() *cartographerv1alpha1.SupplyChainSpec {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *SupplyChainSpecDie) DieStamp(fn func(r *cartographerv1alpha1.SupplyChainSpec)) *SupplyChainSpecDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *SupplyChainSpecDie) DeepCopy() *SupplyChainSpecDie {
	r := *d.r.DeepCopy()
	return &SupplyChainSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

func (d *SupplyChainSpecDie) Resources(v ...cartographerv1alpha1.SupplyChainResource) *SupplyChainSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainSpec) {
		r.Resources = v
	})
}

func (d *SupplyChainSpecDie) Params(v ...cartographerv1alpha1.DelegatableParam) *SupplyChainSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainSpec) {
		r.Params = v
	})
}

func (d *SupplyChainSpecDie) ServiceAccountRef(v cartographerv1alpha1.ServiceAccountRef) *SupplyChainSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainSpec) {
		r.ServiceAccountRef = v
	})
}

func (d *SupplyChainSpecDie) Selector(v map[string]string) *SupplyChainSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainSpec) {
		r.Selector = v
	})
}

func (d *SupplyChainSpecDie) SelectorMatchExpressions(v ...metav1.LabelSelectorRequirement) *SupplyChainSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainSpec) {
		r.SelectorMatchExpressions = v
	})
}

func (d *SupplyChainSpecDie) SelectorMatchFields(v ...cartographerv1alpha1.FieldSelectorRequirement) *SupplyChainSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainSpec) {
		r.SelectorMatchFields = v
	})
}

var SupplyChainStatusBlank = (&SupplyChainStatusDie{}).DieFeed(cartographerv1alpha1.SupplyChainStatus{})

type SupplyChainStatusDie struct {
	mutable bool
	r       cartographerv1alpha1.SupplyChainStatus
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *SupplyChainStatusDie) DieImmutable(immutable bool) *SupplyChainStatusDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *SupplyChainStatusDie) DieFeed(r cartographerv1alpha1.SupplyChainStatus) *SupplyChainStatusDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &SupplyChainStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *SupplyChainStatusDie) DieFeedPtr(r *cartographerv1alpha1.SupplyChainStatus) *SupplyChainStatusDie {
	if r == nil {
		r = &cartographerv1alpha1.SupplyChainStatus{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *SupplyChainStatusDie) DieRelease() cartographerv1alpha1.SupplyChainStatus {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *SupplyChainStatusDie) DieReleasePtr() *cartographerv1alpha1.SupplyChainStatus {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *SupplyChainStatusDie) DieStamp(fn func(r *cartographerv1alpha1.SupplyChainStatus)) *SupplyChainStatusDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *SupplyChainStatusDie) DeepCopy() *SupplyChainStatusDie {
	r := *d.r.DeepCopy()
	return &SupplyChainStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

func (d *SupplyChainStatusDie) Conditions(v ...metav1.Condition) *SupplyChainStatusDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainStatus) {
		r.Conditions = v
	})
}

func (d *SupplyChainStatusDie) ObservedGeneration(v int64) *SupplyChainStatusDie {
	return d.DieStamp(func(r *cartographerv1alpha1.SupplyChainStatus) {
		r.ObservedGeneration = v
	})
}

var WorkloadBlank = (&WorkloadDie{}).DieFeed(cartographerv1alpha1.Workload{})

type WorkloadDie struct {
	v1.FrozenObjectMeta
	mutable bool
	r       cartographerv1alpha1.Workload
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *WorkloadDie) DieImmutable(immutable bool) *WorkloadDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *WorkloadDie) DieFeed(r cartographerv1alpha1.Workload) *WorkloadDie {
	if d.mutable {
		d.FrozenObjectMeta = v1.FreezeObjectMeta(r.ObjectMeta)
		d.r = r
		return d
	}
	return &WorkloadDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *WorkloadDie) DieFeedPtr(r *cartographerv1alpha1.Workload) *WorkloadDie {
	if r == nil {
		r = &cartographerv1alpha1.Workload{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *WorkloadDie) DieRelease() cartographerv1alpha1.Workload {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *WorkloadDie) DieReleasePtr() *cartographerv1alpha1.Workload {
	r := d.DieRelease()
	return &r
}

// DieReleaseUnstructured returns the resource managed by the die as an unstructured object.
func (d *WorkloadDie) DieReleaseUnstructured() runtime.Unstructured {
	r := d.DieReleasePtr()
	u, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	return &unstructured.Unstructured{
		Object: u,
	}
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *WorkloadDie) DieStamp(fn func(r *cartographerv1alpha1.Workload)) *WorkloadDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *WorkloadDie) DeepCopy() *WorkloadDie {
	r := *d.r.DeepCopy()
	return &WorkloadDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

var _ runtime.Object = (*WorkloadDie)(nil)

func (d *WorkloadDie) DeepCopyObject() runtime.Object {
	return d.r.DeepCopy()
}

func (d *WorkloadDie) GetObjectKind() schema.ObjectKind {
	r := d.DieRelease()
	return r.GetObjectKind()
}

func (d *WorkloadDie) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.r)
}

func (d *WorkloadDie) UnmarshalJSON(b []byte) error {
	if d == WorkloadBlank {
		return fmtx.Errorf("cannot unmarshal into the blank die, create a copy first")
	}
	if !d.mutable {
		return fmtx.Errorf("cannot unmarshal into immutable dies, create a mutable version first")
	}
	r := &cartographerv1alpha1.Workload{}
	err := json.Unmarshal(b, r)
	*d = *d.DieFeed(*r)
	return err
}

// MetadataDie stamps the resource's ObjectMeta field with a mutable die.
func (d *WorkloadDie) MetadataDie(fn func(d *v1.ObjectMetaDie)) *WorkloadDie {
	return d.DieStamp(func(r *cartographerv1alpha1.Workload) {
		d := v1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

// SpecDie stamps the resource's spec field with a mutable die.
func (d *WorkloadDie) SpecDie(fn func(d *WorkloadSpecDie)) *WorkloadDie {
	return d.DieStamp(func(r *cartographerv1alpha1.Workload) {
		d := WorkloadSpecBlank.DieImmutable(false).DieFeed(r.Spec)
		fn(d)
		r.Spec = d.DieRelease()
	})
}

// StatusDie stamps the resource's status field with a mutable die.
func (d *WorkloadDie) StatusDie(fn func(d *WorkloadStatusDie)) *WorkloadDie {
	return d.DieStamp(func(r *cartographerv1alpha1.Workload) {
		d := WorkloadStatusBlank.DieImmutable(false).DieFeed(r.Status)
		fn(d)
		r.Status = d.DieRelease()
	})
}

func (d *WorkloadDie) Spec(v cartographerv1alpha1.WorkloadSpec) *WorkloadDie {
	return d.DieStamp(func(r *cartographerv1alpha1.Workload) {
		r.Spec = v
	})
}

func (d *WorkloadDie) Status(v cartographerv1alpha1.WorkloadStatus) *WorkloadDie {
	return d.DieStamp(func(r *cartographerv1alpha1.Workload) {
		r.Status = v
	})
}

var WorkloadSpecBlank = (&WorkloadSpecDie{}).DieFeed(cartographerv1alpha1.WorkloadSpec{})

type WorkloadSpecDie struct {
	mutable bool
	r       cartographerv1alpha1.WorkloadSpec
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *WorkloadSpecDie) DieImmutable(immutable bool) *WorkloadSpecDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *WorkloadSpecDie) DieFeed(r cartographerv1alpha1.WorkloadSpec) *WorkloadSpecDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &WorkloadSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *WorkloadSpecDie) DieFeedPtr(r *cartographerv1alpha1.WorkloadSpec) *WorkloadSpecDie {
	if r == nil {
		r = &cartographerv1alpha1.WorkloadSpec{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *WorkloadSpecDie) DieRelease() cartographerv1alpha1.WorkloadSpec {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *WorkloadSpecDie) DieReleasePtr() *cartographerv1alpha1.WorkloadSpec {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *WorkloadSpecDie) DieStamp(fn func(r *cartographerv1alpha1.WorkloadSpec)) *WorkloadSpecDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *WorkloadSpecDie) DeepCopy() *WorkloadSpecDie {
	r := *d.r.DeepCopy()
	return &WorkloadSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

func (d *WorkloadSpecDie) Params(v ...cartographerv1alpha1.Param) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.Params = v
	})
}

func (d *WorkloadSpecDie) Source(v *cartographerv1alpha1.Source) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.Source = v
	})
}

func (d *WorkloadSpecDie) Build(v *cartographerv1alpha1.WorkloadBuild) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.Build = v
	})
}

func (d *WorkloadSpecDie) Env(v ...corev1.EnvVar) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.Env = v
	})
}

// Image is a pre-built image in a registry. It is an alternative to defining source code.
func (d *WorkloadSpecDie) Image(v string) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.Image = v
	})
}

func (d *WorkloadSpecDie) Resources(v *corev1.ResourceRequirements) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.Resources = v
	})
}

func (d *WorkloadSpecDie) ServiceAccountName(v *string) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.ServiceAccountName = v
	})
}

func (d *WorkloadSpecDie) ServiceClaims(v ...cartographerv1alpha1.WorkloadServiceClaim) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadSpec) {
		r.ServiceClaims = v
	})
}

var WorkloadStatusBlank = (&WorkloadStatusDie{}).DieFeed(cartographerv1alpha1.WorkloadStatus{})

type WorkloadStatusDie struct {
	mutable bool
	r       cartographerv1alpha1.WorkloadStatus
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *WorkloadStatusDie) DieImmutable(immutable bool) *WorkloadStatusDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *WorkloadStatusDie) DieFeed(r cartographerv1alpha1.WorkloadStatus) *WorkloadStatusDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &WorkloadStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *WorkloadStatusDie) DieFeedPtr(r *cartographerv1alpha1.WorkloadStatus) *WorkloadStatusDie {
	if r == nil {
		r = &cartographerv1alpha1.WorkloadStatus{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *WorkloadStatusDie) DieRelease() cartographerv1alpha1.WorkloadStatus {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *WorkloadStatusDie) DieReleasePtr() *cartographerv1alpha1.WorkloadStatus {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *WorkloadStatusDie) DieStamp(fn func(r *cartographerv1alpha1.WorkloadStatus)) *WorkloadStatusDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *WorkloadStatusDie) DeepCopy() *WorkloadStatusDie {
	r := *d.r.DeepCopy()
	return &WorkloadStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

// ObservedGeneration refers to the metadata.Generation of the spec that resulted in the current `status`.
func (d *WorkloadStatusDie) ObservedGeneration(v int64) *WorkloadStatusDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadStatus) {
		r.ObservedGeneration = v
	})
}

// Conditions describing this resource's reconcile state. The top level condition is of type `Ready`, and follows these Kubernetes conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
func (d *WorkloadStatusDie) Conditions(v ...metav1.Condition) *WorkloadStatusDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadStatus) {
		r.Conditions = v
	})
}

// SupplyChainRef is the Supply Chain resource that was used when this status was set.
func (d *WorkloadStatusDie) SupplyChainRef(v cartographerv1alpha1.ObjectReference) *WorkloadStatusDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadStatus) {
		r.SupplyChainRef = v
	})
}

// Resources contain references to the objects created by the Supply Chain and the templates used to create them. It also contains Inputs and Outputs that were passed between the templates as the Supply Chain was processed.
func (d *WorkloadStatusDie) Resources(v ...cartographerv1alpha1.RealizedResource) *WorkloadStatusDie {
	return d.DieStamp(func(r *cartographerv1alpha1.WorkloadStatus) {
		r.Resources = v
	})
}

var RealizedResourceBlank = (&RealizedResourceDie{}).DieFeed(cartographerv1alpha1.RealizedResource{})

type RealizedResourceDie struct {
	mutable bool
	r       cartographerv1alpha1.RealizedResource
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *RealizedResourceDie) DieImmutable(immutable bool) *RealizedResourceDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *RealizedResourceDie) DieFeed(r cartographerv1alpha1.RealizedResource) *RealizedResourceDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &RealizedResourceDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *RealizedResourceDie) DieFeedPtr(r *cartographerv1alpha1.RealizedResource) *RealizedResourceDie {
	if r == nil {
		r = &cartographerv1alpha1.RealizedResource{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *RealizedResourceDie) DieRelease() cartographerv1alpha1.RealizedResource {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *RealizedResourceDie) DieReleasePtr() *cartographerv1alpha1.RealizedResource {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *RealizedResourceDie) DieStamp(fn func(r *cartographerv1alpha1.RealizedResource)) *RealizedResourceDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *RealizedResourceDie) DeepCopy() *RealizedResourceDie {
	r := *d.r.DeepCopy()
	return &RealizedResourceDie{
		mutable: d.mutable,
		r:       r,
	}
}

// Name is the name of the resource in the blueprint
func (d *RealizedResourceDie) Name(v string) *RealizedResourceDie {
	return d.DieStamp(func(r *cartographerv1alpha1.RealizedResource) {
		r.Name = v
	})
}

// StampedRef is a reference to the object that was created by the resource
func (d *RealizedResourceDie) StampedRef(v *corev1.ObjectReference) *RealizedResourceDie {
	return d.DieStamp(func(r *cartographerv1alpha1.RealizedResource) {
		r.StampedRef = v
	})
}

// TemplateRef is a reference to the template used to create the object in StampedRef
func (d *RealizedResourceDie) TemplateRef(v *corev1.ObjectReference) *RealizedResourceDie {
	return d.DieStamp(func(r *cartographerv1alpha1.RealizedResource) {
		r.TemplateRef = v
	})
}

// Inputs are references to resources that were used to template the object in StampedRef
func (d *RealizedResourceDie) Inputs(v ...cartographerv1alpha1.Input) *RealizedResourceDie {
	return d.DieStamp(func(r *cartographerv1alpha1.RealizedResource) {
		r.Inputs = v
	})
}

// Outputs are values from the object in StampedRef that can be consumed by other resources
func (d *RealizedResourceDie) Outputs(v ...cartographerv1alpha1.Output) *RealizedResourceDie {
	return d.DieStamp(func(r *cartographerv1alpha1.RealizedResource) {
		r.Outputs = v
	})
}

// Conditions describing this resource's reconcile state. The top level condition is of type `Ready`, and follows these Kubernetes conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
func (d *RealizedResourceDie) Conditions(v ...metav1.Condition) *RealizedResourceDie {
	return d.DieStamp(func(r *cartographerv1alpha1.RealizedResource) {
		r.Conditions = v
	})
}
