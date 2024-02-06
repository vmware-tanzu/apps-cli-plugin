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

// +die
type _ = corev1.Container

func (d *ContainerDie) PortsDie(ports ...*ContainerPortDie) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		r.Ports = make([]corev1.ContainerPort, len(ports))
		for i := range ports {
			r.Ports[i] = ports[i].DieRelease()
		}
	})
}

func (d *ContainerDie) EnvFromDie(prefix string, fn func(d *EnvFromSourceDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		for i := range r.EnvFrom {
			if prefix == r.EnvFrom[i].Prefix {
				d := EnvFromSourceBlank.DieImmutable(false).DieFeed(r.EnvFrom[i])
				fn(d)
				r.EnvFrom[i] = d.DieRelease()
				return
			}
		}

		d := EnvFromSourceBlank.DieImmutable(false).DieFeed(corev1.EnvFromSource{Prefix: prefix})
		fn(d)
		r.EnvFrom = append(r.EnvFrom, d.DieRelease())
	})
}

func (d *ContainerDie) EnvDie(name string, fn func(d *EnvVarDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		for i := range r.Env {
			if name == r.Env[i].Name {
				d := EnvVarBlank.DieImmutable(false).DieFeed(r.Env[i])
				fn(d)
				r.Env[i] = d.DieRelease()
				return
			}
		}

		d := EnvVarBlank.DieImmutable(false).DieFeed(corev1.EnvVar{Name: name})
		fn(d)
		r.Env = append(r.Env, d.DieRelease())
	})
}

func (d *ContainerDie) ResourcesDie(fn func(d *ResourceRequirementsDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		d := ResourceRequirementsBlank.DieImmutable(false).DieFeed(r.Resources)
		fn(d)
		r.Resources = d.DieRelease()
	})
}

func (d *ContainerDie) ResizePolicyDie(name corev1.ResourceName, fn func(d *ContainerResizePolicyDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		for i := range r.ResizePolicy {
			if name == r.ResizePolicy[i].ResourceName {
				d := ContainerResizePolicyBlank.DieImmutable(false).DieFeed(r.ResizePolicy[i])
				fn(d)
				r.ResizePolicy[i] = d.DieRelease()
				return
			}
		}

		d := ContainerResizePolicyBlank.DieImmutable(false).DieFeed(corev1.ContainerResizePolicy{ResourceName: name})
		fn(d)
		r.ResizePolicy = append(r.ResizePolicy, d.DieRelease())
	})
}

func (d *ContainerDie) VolumeMountDie(name string, fn func(d *VolumeMountDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		for i := range r.VolumeMounts {
			if name == r.VolumeMounts[i].Name {
				d := VolumeMountBlank.DieImmutable(false).DieFeed(r.VolumeMounts[i])
				fn(d)
				r.VolumeMounts[i] = d.DieRelease()
				return
			}
		}

		d := VolumeMountBlank.DieImmutable(false).DieFeed(corev1.VolumeMount{Name: name})
		fn(d)
		r.VolumeMounts = append(r.VolumeMounts, d.DieRelease())
	})
}

func (d *ContainerDie) VolumeDeviceDie(name string, fn func(d *VolumeDeviceDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		for i := range r.VolumeDevices {
			if name == r.VolumeDevices[i].Name {
				d := VolumeDeviceBlank.DieImmutable(false).DieFeed(r.VolumeDevices[i])
				fn(d)
				r.VolumeDevices[i] = d.DieRelease()
				return
			}
		}

		d := VolumeDeviceBlank.DieImmutable(false).DieFeed(corev1.VolumeDevice{Name: name})
		fn(d)
		r.VolumeDevices = append(r.VolumeDevices, d.DieRelease())
	})
}

func (d *ContainerDie) LivenessProbeDie(fn func(d *ProbeDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		d := ProbeBlank.DieImmutable(false).DieFeedPtr(r.LivenessProbe)
		fn(d)
		r.LivenessProbe = d.DieReleasePtr()
	})
}

func (d *ContainerDie) ReadinessProbeDie(fn func(d *ProbeDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		d := ProbeBlank.DieImmutable(false).DieFeedPtr(r.ReadinessProbe)
		fn(d)
		r.ReadinessProbe = d.DieReleasePtr()
	})
}

func (d *ContainerDie) StartupProbeDie(fn func(d *ProbeDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		d := ProbeBlank.DieImmutable(false).DieFeedPtr(r.StartupProbe)
		fn(d)
		r.StartupProbe = d.DieReleasePtr()
	})
}

func (d *ContainerDie) LifecycleDie(fn func(d *LifecycleDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		d := LifecycleBlank.DieImmutable(false).DieFeedPtr(r.Lifecycle)
		fn(d)
		r.Lifecycle = d.DieReleasePtr()
	})
}

func (d *ContainerDie) SecurityContextDie(fn func(d *SecurityContextDie)) *ContainerDie {
	return d.DieStamp(func(r *corev1.Container) {
		d := SecurityContextBlank.DieImmutable(false).DieFeedPtr(r.SecurityContext)
		fn(d)
		r.SecurityContext = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ContainerPort

// +die
type _ = corev1.EnvFromSource

func (d *EnvFromSourceDie) ConfigMapRefDie(fn func(d *ConfigMapEnvSourceDie)) *EnvFromSourceDie {
	return d.DieStamp(func(r *corev1.EnvFromSource) {
		d := ConfigMapEnvSourceBlank.DieImmutable(false).DieFeedPtr(r.ConfigMapRef)
		fn(d)
		r.ConfigMapRef = d.DieReleasePtr()
	})
}

func (d *EnvFromSourceDie) SecretRefDie(fn func(d *SecretEnvSourceDie)) *EnvFromSourceDie {
	return d.DieStamp(func(r *corev1.EnvFromSource) {
		d := SecretEnvSourceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ConfigMapEnvSource

func (d *ConfigMapEnvSourceDie) Name(v string) *ConfigMapEnvSourceDie {
	return d.DieStamp(func(r *corev1.ConfigMapEnvSource) {
		r.Name = v
	})
}

// +die
type _ = corev1.SecretEnvSource

func (d *SecretEnvSourceDie) Name(v string) *SecretEnvSourceDie {
	return d.DieStamp(func(r *corev1.SecretEnvSource) {
		r.Name = v
	})
}

// +die
type _ = corev1.EnvVar

func (d *EnvVarDie) ValueFromDie(fn func(d *EnvVarSourceDie)) *EnvVarDie {
	return d.DieStamp(func(r *corev1.EnvVar) {
		d := EnvVarSourceBlank.DieImmutable(false).DieFeedPtr(r.ValueFrom)
		fn(d)
		r.ValueFrom = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.EnvVarSource

func (d *EnvVarSourceDie) FieldRefDie(fn func(d *ObjectFieldSelectorDie)) *EnvVarSourceDie {
	return d.DieStamp(func(r *corev1.EnvVarSource) {
		d := ObjectFieldSelectorBlank.DieImmutable(false).DieFeedPtr(r.FieldRef)
		fn(d)
		r.FieldRef = d.DieReleasePtr()
	})
}

func (d *EnvVarSourceDie) ResourceFieldRefDie(fn func(d *ResourceFieldSelectorDie)) *EnvVarSourceDie {
	return d.DieStamp(func(r *corev1.EnvVarSource) {
		d := ResourceFieldSelectorBlank.DieImmutable(false).DieFeedPtr(r.ResourceFieldRef)
		fn(d)
		r.ResourceFieldRef = d.DieReleasePtr()
	})
}

func (d *EnvVarSourceDie) ConfigMapKeyRefDie(fn func(d *ConfigMapKeySelectorDie)) *EnvVarSourceDie {
	return d.DieStamp(func(r *corev1.EnvVarSource) {
		d := ConfigMapKeySelectorBlank.DieImmutable(false).DieFeedPtr(r.ConfigMapKeyRef)
		fn(d)
		r.ConfigMapKeyRef = d.DieReleasePtr()
	})
}

func (d *EnvVarSourceDie) SecretKeyRefDie(fn func(d *SecretKeySelectorDie)) *EnvVarSourceDie {
	return d.DieStamp(func(r *corev1.EnvVarSource) {
		d := SecretKeySelectorBlank.DieImmutable(false).DieFeedPtr(r.SecretKeyRef)
		fn(d)
		r.SecretKeyRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ObjectFieldSelector

// +die
type _ = corev1.ResourceFieldSelector

// +die
type _ = corev1.ConfigMapKeySelector

func (d *ConfigMapKeySelectorDie) Name(v string) *ConfigMapKeySelectorDie {
	return d.DieStamp(func(r *corev1.ConfigMapKeySelector) {
		r.Name = v
	})
}

// +die
type _ = corev1.SecretKeySelector

func (d *SecretKeySelectorDie) Name(v string) *SecretKeySelectorDie {
	return d.DieStamp(func(r *corev1.SecretKeySelector) {
		r.Name = v
	})
}

// +die
type _ = corev1.ResourceRequirements

func (d *ResourceRequirementsDie) AddLimit(name corev1.ResourceName, quantity resource.Quantity) *ResourceRequirementsDie {
	return d.DieStamp(func(r *corev1.ResourceRequirements) {
		if r.Limits == nil {
			r.Limits = corev1.ResourceList{}
		}
		r.Limits[name] = quantity
	})
}

func (d *ResourceRequirementsDie) AddLimitString(name corev1.ResourceName, quantity string) *ResourceRequirementsDie {
	return d.AddLimit(name, resource.MustParse(quantity))
}

func (d *ResourceRequirementsDie) AddRequest(name corev1.ResourceName, quantity resource.Quantity) *ResourceRequirementsDie {
	return d.DieStamp(func(r *corev1.ResourceRequirements) {
		if r.Requests == nil {
			r.Requests = corev1.ResourceList{}
		}
		r.Requests[name] = quantity
	})
}

func (d *ResourceRequirementsDie) AddRequestString(name corev1.ResourceName, quantity string) *ResourceRequirementsDie {
	return d.AddRequest(name, resource.MustParse(quantity))
}

func (d *ResourceRequirementsDie) ClaimsDie(claims ...*ResourceClaimDie) *ResourceRequirementsDie {
	return d.DieStamp(func(r *corev1.ResourceRequirements) {
		r.Claims = make([]corev1.ResourceClaim, len(claims))
		for i := range claims {
			r.Claims[i] = claims[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.ResourceClaim

// +die
type _ = corev1.ContainerResizePolicy

// +die
type _ = corev1.VolumeMount

// +die
type _ = corev1.VolumeDevice

// +die
type _ = corev1.Probe

func (d *ProbeDie) ProbeHandlerDie(fn func(d *ProbeHandlerDie)) *ProbeDie {
	return d.DieStamp(func(r *corev1.Probe) {
		d := ProbeHandlerBlank.DieImmutable(false).DieFeed(r.ProbeHandler)
		fn(d)
		r.ProbeHandler = d.DieRelease()
	})
}

func (d *ProbeDie) ExecDie(fn func(d *ExecActionDie)) *ProbeDie {
	return d.DieStamp(func(r *corev1.Probe) {
		d := ExecActionBlank.DieImmutable(false).DieFeedPtr(r.Exec)
		fn(d)
		r.ProbeHandler = corev1.ProbeHandler{
			Exec: d.DieReleasePtr(),
		}
	})
}

func (d *ProbeDie) HTTPGetDie(fn func(d *HTTPGetActionDie)) *ProbeDie {
	return d.DieStamp(func(r *corev1.Probe) {
		d := HTTPGetActionBlank.DieImmutable(false).DieFeedPtr(r.HTTPGet)
		fn(d)
		r.ProbeHandler = corev1.ProbeHandler{
			HTTPGet: d.DieReleasePtr(),
		}
	})
}

func (d *ProbeDie) TCPSocketDie(fn func(d *TCPSocketActionDie)) *ProbeDie {
	return d.DieStamp(func(r *corev1.Probe) {
		d := TCPSocketActionBlank.DieImmutable(false).DieFeedPtr(r.TCPSocket)
		fn(d)
		r.ProbeHandler = corev1.ProbeHandler{
			TCPSocket: d.DieReleasePtr(),
		}
	})
}

// +die
type _ = corev1.Lifecycle

func (d *LifecycleDie) PostStartDie(fn func(d *LifecycleHandlerDie)) *LifecycleDie {
	return d.DieStamp(func(r *corev1.Lifecycle) {
		d := LifecycleHandlerBlank.DieImmutable(false).DieFeedPtr(r.PostStart)
		fn(d)
		r.PostStart = d.DieReleasePtr()
	})
}

func (d *LifecycleDie) PreStopDie(fn func(d *LifecycleHandlerDie)) *LifecycleDie {
	return d.DieStamp(func(r *corev1.Lifecycle) {
		d := LifecycleHandlerBlank.DieImmutable(false).DieFeedPtr(r.PreStop)
		fn(d)
		r.PreStop = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.LifecycleHandler

func (d *LifecycleHandlerDie) ExecDie(fn func(d *ExecActionDie)) *LifecycleHandlerDie {
	return d.DieStamp(func(r *corev1.LifecycleHandler) {
		d := ExecActionBlank.DieImmutable(false).DieFeedPtr(r.Exec)
		fn(d)
		r.Exec = d.DieReleasePtr()
	})
}

func (d *LifecycleHandlerDie) HTTPGetDie(fn func(d *HTTPGetActionDie)) *LifecycleHandlerDie {
	return d.DieStamp(func(r *corev1.LifecycleHandler) {
		d := HTTPGetActionBlank.DieImmutable(false).DieFeedPtr(r.HTTPGet)
		fn(d)
		r.HTTPGet = d.DieReleasePtr()
	})
}

func (d *LifecycleHandlerDie) TCPSocketDie(fn func(d *TCPSocketActionDie)) *LifecycleHandlerDie {
	return d.DieStamp(func(r *corev1.LifecycleHandler) {
		d := TCPSocketActionBlank.DieImmutable(false).DieFeedPtr(r.TCPSocket)
		fn(d)
		r.TCPSocket = d.DieReleasePtr()
	})
}

func (d *LifecycleHandlerDie) SleepDie(fn func(d *SleepActionDie)) *LifecycleHandlerDie {
	return d.DieStamp(func(r *corev1.LifecycleHandler) {
		d := SleepActionBlank.DieImmutable(false).DieFeedPtr(r.Sleep)
		fn(d)
		r.Sleep = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ProbeHandler

func (d *ProbeHandlerDie) ExecDie(fn func(d *ExecActionDie)) *ProbeHandlerDie {
	return d.DieStamp(func(r *corev1.ProbeHandler) {
		d := ExecActionBlank.DieImmutable(false).DieFeedPtr(r.Exec)
		fn(d)
		r.Exec = d.DieReleasePtr()
	})
}

func (d *ProbeHandlerDie) HTTPGetDie(fn func(d *HTTPGetActionDie)) *ProbeHandlerDie {
	return d.DieStamp(func(r *corev1.ProbeHandler) {
		d := HTTPGetActionBlank.DieImmutable(false).DieFeedPtr(r.HTTPGet)
		fn(d)
		r.HTTPGet = d.DieReleasePtr()
	})
}

func (d *ProbeHandlerDie) TCPSocketDie(fn func(d *TCPSocketActionDie)) *ProbeHandlerDie {
	return d.DieStamp(func(r *corev1.ProbeHandler) {
		d := TCPSocketActionBlank.DieImmutable(false).DieFeedPtr(r.TCPSocket)
		fn(d)
		r.TCPSocket = d.DieReleasePtr()
	})
}

func (d *ProbeHandlerDie) GRPCDie(fn func(d *GRPCActionDie)) *ProbeHandlerDie {
	return d.DieStamp(func(r *corev1.ProbeHandler) {
		d := GRPCActionBlank.DieImmutable(false).DieFeedPtr(r.GRPC)
		fn(d)
		r.GRPC = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ExecAction

// +die
type _ = corev1.HTTPGetAction

func (d *HTTPGetActionDie) HTTPHeadersDie(headers ...*HTTPHeaderDie) *HTTPGetActionDie {
	return d.DieStamp(func(r *corev1.HTTPGetAction) {
		r.HTTPHeaders = make([]corev1.HTTPHeader, len(headers))
		for i := range headers {
			r.HTTPHeaders[i] = headers[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.HTTPHeader

// +die
type _ = corev1.TCPSocketAction

// +die
type _ = corev1.GRPCAction

// +die
type _ = corev1.SleepAction

// +die
type _ = corev1.SecurityContext

func (d *SecurityContextDie) CapabilitiesDie(fn func(d *CapabilitiesDie)) *SecurityContextDie {
	return d.DieStamp(func(r *corev1.SecurityContext) {
		d := CapabilitiesBlank.DieImmutable(false).DieFeedPtr(r.Capabilities)
		fn(d)
		r.Capabilities = d.DieReleasePtr()
	})
}

func (d *SecurityContextDie) SELinuxOptionsDie(fn func(d *SELinuxOptionsDie)) *SecurityContextDie {
	return d.DieStamp(func(r *corev1.SecurityContext) {
		d := SELinuxOptionsBlank.DieImmutable(false).DieFeedPtr(r.SELinuxOptions)
		fn(d)
		r.SELinuxOptions = d.DieReleasePtr()
	})
}

func (d *SecurityContextDie) WindowsOptionsDie(fn func(d *WindowsSecurityContextOptionsDie)) *SecurityContextDie {
	return d.DieStamp(func(r *corev1.SecurityContext) {
		d := WindowsSecurityContextOptionsBlank.DieImmutable(false).DieFeedPtr(r.WindowsOptions)
		fn(d)
		r.WindowsOptions = d.DieReleasePtr()
	})
}

func (d *SecurityContextDie) SeccompProfileDie(fn func(d *SeccompProfileDie)) *SecurityContextDie {
	return d.DieStamp(func(r *corev1.SecurityContext) {
		d := SeccompProfileBlank.DieImmutable(false).DieFeedPtr(r.SeccompProfile)
		fn(d)
		r.SeccompProfile = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.Capabilities

// +die
type _ = corev1.SELinuxOptions

// +die
type _ = corev1.WindowsSecurityContextOptions

// +die
type _ = corev1.SeccompProfile

// +die
type _ = corev1.ContainerStatus

func (d *ContainerStatusDie) StateDie(fn func(d *ContainerStateDie)) *ContainerStatusDie {
	return d.DieStamp(func(r *corev1.ContainerStatus) {
		d := ContainerStateBlank.DieImmutable(false).DieFeed(r.State)
		fn(d)
		r.State = d.DieRelease()
	})
}

func (d *ContainerStatusDie) LastTerminationStateDie(fn func(d *ContainerStateDie)) *ContainerStatusDie {
	return d.DieStamp(func(r *corev1.ContainerStatus) {
		d := ContainerStateBlank.DieImmutable(false).DieFeed(r.LastTerminationState)
		fn(d)
		r.LastTerminationState = d.DieRelease()
	})
}

func (d *ContainerStatusDie) AddAllocatedResource(name corev1.ResourceName, quantity resource.Quantity) *ContainerStatusDie {
	return d.DieStamp(func(r *corev1.ContainerStatus) {
		if r.AllocatedResources == nil {
			r.AllocatedResources = corev1.ResourceList{}
		}
		r.AllocatedResources[name] = quantity
	})
}

func (d *ContainerStatusDie) AddAllocatedResourceString(name corev1.ResourceName, quantity string) *ContainerStatusDie {
	return d.AddAllocatedResource(name, resource.MustParse(quantity))
}

func (d *ContainerStatusDie) ResourcesDie(fn func(d *ResourceRequirementsDie)) *ContainerStatusDie {
	return d.DieStamp(func(r *corev1.ContainerStatus) {
		d := ResourceRequirementsBlank.DieImmutable(false).DieFeedPtr(r.Resources)
		fn(d)
		r.Resources = d.DieReleasePtr()
	})
}

// ADD

// +die
type _ = corev1.ContainerState

func (d *ContainerStateDie) WaitingDie(fn func(d *ContainerStateWaitingDie)) *ContainerStateDie {
	return d.DieStamp(func(r *corev1.ContainerState) {
		d := ContainerStateWaitingBlank.DieImmutable(false).DieFeedPtr(r.Waiting)
		fn(d)
		r.Waiting = d.DieReleasePtr()
	})
}

func (d *ContainerStateDie) RunningDie(fn func(d *ContainerStateRunningDie)) *ContainerStateDie {
	return d.DieStamp(func(r *corev1.ContainerState) {
		d := ContainerStateRunningBlank.DieImmutable(false).DieFeedPtr(r.Running)
		fn(d)
		r.Running = d.DieReleasePtr()
	})
}

func (d *ContainerStateDie) TerminatedDie(fn func(d *ContainerStateTerminatedDie)) *ContainerStateDie {
	return d.DieStamp(func(r *corev1.ContainerState) {
		d := ContainerStateTerminatedBlank.DieImmutable(false).DieFeedPtr(r.Terminated)
		fn(d)
		r.Terminated = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ContainerStateWaiting

// +die
type _ = corev1.ContainerStateRunning

// +die
type _ = corev1.ContainerStateTerminated
