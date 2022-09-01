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
	"k8s.io/apimachinery/pkg/api/resource"
)

// +die:object=true
type _ = corev1.ResourceQuota

// +die
type _ = corev1.ResourceQuotaSpec

func (d *ResourceQuotaSpecDie) AddHard(name corev1.ResourceName, quantity resource.Quantity) *ResourceQuotaSpecDie {
	return d.DieStamp(func(r *corev1.ResourceQuotaSpec) {
		if r.Hard == nil {
			r.Hard = corev1.ResourceList{}
		}
		r.Hard[name] = quantity
	})
}

func (d *ResourceQuotaSpecDie) AddHardString(name corev1.ResourceName, quantity string) *ResourceQuotaSpecDie {
	return d.AddHard(name, resource.MustParse(quantity))
}

func (d *ResourceQuotaSpecDie) ScopeSelectorDie(fn func(d *ScopeSelectorDie)) *ResourceQuotaSpecDie {
	return d.DieStamp(func(r *corev1.ResourceQuotaSpec) {
		d := ScopeSelectorBlank.DieImmutable(false).DieFeedPtr(r.ScopeSelector)
		fn(d)
		r.ScopeSelector = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ScopeSelector

func (d *ScopeSelectorDie) MatchExpressionDie(scope corev1.ResourceQuotaScope, fn func(d *ScopedResourceSelectorRequirementDie)) *ScopeSelectorDie {
	return d.DieStamp(func(r *corev1.ScopeSelector) {
		for i := range r.MatchExpressions {
			if scope == r.MatchExpressions[i].ScopeName {
				d := ScopedResourceSelectorRequirementBlank.DieImmutable(false).DieFeed(r.MatchExpressions[i])
				fn(d)
				r.MatchExpressions[i] = d.DieRelease()
				return
			}
		}

		d := ScopedResourceSelectorRequirementBlank.DieImmutable(false).DieFeed(corev1.ScopedResourceSelectorRequirement{ScopeName: scope})
		fn(d)
		r.MatchExpressions = append(r.MatchExpressions, d.DieRelease())
	})
}

// +die
type _ = corev1.ScopedResourceSelectorRequirement

// +die
type _ = corev1.ResourceQuotaStatus

func (d *ResourceQuotaStatusDie) AddHard(name corev1.ResourceName, quantity resource.Quantity) *ResourceQuotaStatusDie {
	return d.DieStamp(func(r *corev1.ResourceQuotaStatus) {
		if r.Hard == nil {
			r.Hard = corev1.ResourceList{}
		}
		r.Hard[name] = quantity
	})
}

func (d *ResourceQuotaStatusDie) AddHardString(name corev1.ResourceName, quantity string) *ResourceQuotaStatusDie {
	return d.AddHard(name, resource.MustParse(quantity))
}

func (d *ResourceQuotaStatusDie) AddUsed(name corev1.ResourceName, quantity resource.Quantity) *ResourceQuotaStatusDie {
	return d.DieStamp(func(r *corev1.ResourceQuotaStatus) {
		if r.Used == nil {
			r.Used = corev1.ResourceList{}
		}
		r.Used[name] = quantity
	})
}

func (d *ResourceQuotaStatusDie) AddUsedString(name corev1.ResourceName, quantity string) *ResourceQuotaStatusDie {
	return d.AddUsed(name, resource.MustParse(quantity))
}
