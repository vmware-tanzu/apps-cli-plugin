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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +die:object=true
type _ = corev1.Service

// +die
type _ = corev1.ServiceSpec

func (d *ServiceSpecDie) PortDie(protocol corev1.Protocol, port int32, fn func(d *ServicePortDie)) *ServiceSpecDie {
	return d.DieStamp(func(r *corev1.ServiceSpec) {
		for i := range r.Ports {
			if protocol == r.Ports[i].Protocol && port == r.Ports[i].Port {
				d := ServicePortBlank.DieImmutable(false).DieFeed(r.Ports[i])
				fn(d)
				r.Ports[i] = d.DieRelease()
				return
			}
		}

		d := ServicePortBlank.DieImmutable(false).DieFeed(corev1.ServicePort{Protocol: protocol, Port: port})
		fn(d)
		r.Ports = append(r.Ports, d.DieRelease())
	})
}

func (d *ServiceSpecDie) AddSelector(key, value string) *ServiceSpecDie {
	return d.DieStamp(func(r *corev1.ServiceSpec) {
		r.Selector[key] = value
	})
}

func (d *ServiceSpecDie) SessionAffinityConfigDie(fn func(d *SessionAffinityConfigDie)) *ServiceSpecDie {
	return d.DieStamp(func(r *corev1.ServiceSpec) {
		d := SessionAffinityConfigBlank.DieImmutable(false).DieFeedPtr(r.SessionAffinityConfig)
		fn(d)
		r.SessionAffinityConfig = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ServicePort

// +die
type _ = corev1.SessionAffinityConfig

func (d *SessionAffinityConfigDie) ClientIPDie(fn func(d *ClientIPConfigDie)) *SessionAffinityConfigDie {
	return d.DieStamp(func(r *corev1.SessionAffinityConfig) {
		d := ClientIPConfigBlank.DieImmutable(false).DieFeedPtr(r.ClientIP)
		fn(d)
		r.ClientIP = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ClientIPConfig

// +die
type _ = corev1.ServiceStatus

func (d *ServiceStatusDie) LoadBalancerDie(fn func(d *LoadBalancerStatusDie)) *ServiceStatusDie {
	return d.DieStamp(func(r *corev1.ServiceStatus) {
		d := LoadBalancerStatusBlank.DieImmutable(false).DieFeed(r.LoadBalancer)
		fn(d)
		r.LoadBalancer = d.DieRelease()
	})
}

func (d *ServiceStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *ServiceStatusDie {
	return d.DieStamp(func(r *corev1.ServiceStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i, c := range conditions {
			r.Conditions[i] = c.DieRelease()
		}
	})
}

// +die
type _ = corev1.LoadBalancerStatus

func (d *LoadBalancerStatusDie) LoadBalancerDie(ingress ...*LoadBalancerIngressDie) *LoadBalancerStatusDie {
	return d.DieStamp(func(r *corev1.LoadBalancerStatus) {
		r.Ingress = make([]corev1.LoadBalancerIngress, len(ingress))
		for i := range ingress {
			r.Ingress[i] = ingress[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.LoadBalancerIngress

func (d *LoadBalancerIngressDie) PortsDie(ports ...*PortStatusDie) *LoadBalancerIngressDie {
	return d.DieStamp(func(r *corev1.LoadBalancerIngress) {
		r.Ports = make([]corev1.PortStatus, len(ports))
		for i := range ports {
			r.Ports[i] = ports[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.PortStatus
