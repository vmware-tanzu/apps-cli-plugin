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
type _ = corev1.PersistentVolume

// +die
type _ = corev1.PersistentVolumeSpec

func (d *PersistentVolumeSpecDie) AddCapacity(name corev1.ResourceName, quantity resource.Quantity) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		r.Capacity[name] = quantity
	})
}

func (d *PersistentVolumeSpecDie) AddCapacityString(name corev1.ResourceName, quantity string) *PersistentVolumeSpecDie {
	return d.AddCapacity(name, resource.MustParse(quantity))
}

func (d *PersistentVolumeSpecDie) GCEPersistentDiskDie(fn func(d *GCEPersistentDiskVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := GCEPersistentDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.GCEPersistentDisk)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			GCEPersistentDisk: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) AWSElasticBlockStoreDie(fn func(d *AWSElasticBlockStoreVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := AWSElasticBlockStoreVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.AWSElasticBlockStore)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			AWSElasticBlockStore: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) HostPathDie(fn func(d *HostPathVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := HostPathVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.HostPath)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			HostPath: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) GlusterfsDie(fn func(d *GlusterfsPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := GlusterfsPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Glusterfs)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			Glusterfs: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) NFSDie(fn func(d *NFSVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := NFSVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.NFS)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			NFS: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) RBDDie(fn func(d *RBDPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := RBDPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.RBD)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			RBD: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) ISCSIDie(fn func(d *ISCSIPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := ISCSIPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.ISCSI)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			ISCSI: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) CinderDie(fn func(d *CinderPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := CinderPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Cinder)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			Cinder: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) CephFSDie(fn func(d *CephFSPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := CephFSPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.CephFS)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			CephFS: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) FCDie(fn func(d *FCVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := FCVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.FC)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			FC: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) FlockerDie(fn func(d *FlockerVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := FlockerVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Flocker)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			Flocker: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) FlexVolumeDie(fn func(d *FlexPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := FlexPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.FlexVolume)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			FlexVolume: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) AzureFileDie(fn func(d *AzureFilePersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := AzureFilePersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.AzureFile)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			AzureFile: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) VsphereVolumeDie(fn func(d *VsphereVirtualDiskVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := VsphereVirtualDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.VsphereVolume)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			VsphereVolume: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) QuobyteDie(fn func(d *QuobyteVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := QuobyteVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Quobyte)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			Quobyte: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) AzureDiskDie(fn func(d *AzureDiskVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := AzureDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.AzureDisk)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			AzureDisk: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) PhotonPersistentDiskDie(fn func(d *PhotonPersistentDiskVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := PhotonPersistentDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.PhotonPersistentDisk)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			PhotonPersistentDisk: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) PortworxVolumeDie(fn func(d *PortworxVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := PortworxVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.PortworxVolume)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			PortworxVolume: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) ScaleIODie(fn func(d *ScaleIOPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := ScaleIOPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.ScaleIO)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			ScaleIO: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) LocalDie(fn func(d *LocalVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := LocalVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Local)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			Local: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) StorageOSDie(fn func(d *StorageOSPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := StorageOSPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.StorageOS)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			StorageOS: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) CSIDie(fn func(d *CSIPersistentVolumeSourceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := CSIPersistentVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.CSI)
		fn(d)
		r.PersistentVolumeSource = corev1.PersistentVolumeSource{
			CSI: d.DieReleasePtr(),
		}
	})
}

func (d *PersistentVolumeSpecDie) ClaimRefDie(fn func(d *ObjectReferenceDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := ObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.ClaimRef)
		fn(d)
		r.ClaimRef = d.DieReleasePtr()
	})
}

func (d *PersistentVolumeSpecDie) NodeAffinityDie(fn func(d *VolumeNodeAffinityDie)) *PersistentVolumeSpecDie {
	return d.DieStamp(func(r *corev1.PersistentVolumeSpec) {
		d := VolumeNodeAffinityBlank.DieImmutable(false).DieFeedPtr(r.NodeAffinity)
		fn(d)
		r.NodeAffinity = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.PersistentVolumeStatus

// +die
type _ = corev1.GlusterfsPersistentVolumeSource

// +die
type _ = corev1.RBDPersistentVolumeSource

func (d *RBDPersistentVolumeSourceDie) SecretRefDie(fn func(d *SecretReferenceDie)) *RBDPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.RBDPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.ISCSIPersistentVolumeSource

func (d *ISCSIPersistentVolumeSourceDie) SecretRefDie(fn func(d *SecretReferenceDie)) *ISCSIPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ISCSIPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.CinderPersistentVolumeSource

func (d *CinderPersistentVolumeSourceDie) SecretRefDie(fn func(d *SecretReferenceDie)) *CinderPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CinderPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.CephFSPersistentVolumeSource

func (d *CephFSPersistentVolumeSourceDie) SecretRefDie(fn func(d *SecretReferenceDie)) *CephFSPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CephFSPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.FlexPersistentVolumeSource

func (d *FlexPersistentVolumeSourceDie) SecretRefDie(fn func(d *SecretReferenceDie)) *FlexPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.FlexPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.AzureFilePersistentVolumeSource

// +die
type _ = corev1.ScaleIOPersistentVolumeSource

func (d *ScaleIOPersistentVolumeSourceDie) SecretRefDie(fn func(d *SecretReferenceDie)) *ScaleIOPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ScaleIOPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.LocalVolumeSource

// +die
type _ = corev1.StorageOSPersistentVolumeSource

func (d *StorageOSPersistentVolumeSourceDie) SecretRefDie(fn func(d *ObjectReferenceDie)) *StorageOSPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.StorageOSPersistentVolumeSource) {
		d := ObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.CSIPersistentVolumeSource

func (d *CSIPersistentVolumeSourceDie) AddVolumeAttributes(key, value string) *CSIPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIPersistentVolumeSource) {
		if r.VolumeAttributes == nil {
			r.VolumeAttributes = map[string]string{}
		}
		r.VolumeAttributes[key] = value
	})
}

func (d *CSIPersistentVolumeSourceDie) ControllerPublishSecretRefDie(fn func(d *SecretReferenceDie)) *CSIPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.ControllerPublishSecretRef)
		fn(d)
		r.ControllerPublishSecretRef = d.DieReleasePtr()
	})
}

func (d *CSIPersistentVolumeSourceDie) NodeStageSecretRefDie(fn func(d *SecretReferenceDie)) *CSIPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.NodeStageSecretRef)
		fn(d)
		r.NodeStageSecretRef = d.DieReleasePtr()
	})
}

func (d *CSIPersistentVolumeSourceDie) NodePublishSecretRefDie(fn func(d *SecretReferenceDie)) *CSIPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.NodePublishSecretRef)
		fn(d)
		r.NodePublishSecretRef = d.DieReleasePtr()
	})
}

func (d *CSIPersistentVolumeSourceDie) ControllerExpandSecretRefDie(fn func(d *SecretReferenceDie)) *CSIPersistentVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIPersistentVolumeSource) {
		d := SecretReferenceBlank.DieImmutable(false).DieFeedPtr(r.ControllerExpandSecretRef)
		fn(d)
		r.ControllerExpandSecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.VolumeNodeAffinity

func (d *VolumeNodeAffinityDie) RequiredDie(fn func(d *NodeSelectorDie)) *VolumeNodeAffinityDie {
	return d.DieStamp(func(r *corev1.VolumeNodeAffinity) {
		d := NodeSelectorBlank.DieImmutable(false).DieFeedPtr(r.Required)
		fn(d)
		r.Required = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.NodeSelector

func (d *NodeSelectorDie) NodeSelectorTermsDie(terms ...*NodeSelectorTermDie) *NodeSelectorDie {
	return d.DieStamp(func(r *corev1.NodeSelector) {
		r.NodeSelectorTerms = make([]corev1.NodeSelectorTerm, len(terms))
		for i := range terms {
			r.NodeSelectorTerms[i] = terms[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.NodeSelectorTerm

func (d *NodeSelectorTermDie) MatchExpressionDie(key string, fn func(d *NodeSelectorRequirementDie)) *NodeSelectorTermDie {
	return d.DieStamp(func(r *corev1.NodeSelectorTerm) {
		for i := range r.MatchExpressions {
			if key == r.MatchExpressions[i].Key {
				d := NodeSelectorRequirementBlank.DieImmutable(false).DieFeed(r.MatchExpressions[i])
				fn(d)
				r.MatchExpressions[i] = d.DieRelease()
				return
			}
		}

		d := NodeSelectorRequirementBlank.DieImmutable(false).DieFeed(corev1.NodeSelectorRequirement{Key: key})
		fn(d)
		r.MatchExpressions = append(r.MatchExpressions, d.DieRelease())
	})
}

func (d *NodeSelectorTermDie) MatchFieldsDie(key string, fn func(d *NodeSelectorRequirementDie)) *NodeSelectorTermDie {
	return d.DieStamp(func(r *corev1.NodeSelectorTerm) {
		for i := range r.MatchFields {
			if key == r.MatchFields[i].Key {
				d := NodeSelectorRequirementBlank.DieImmutable(false).DieFeed(r.MatchFields[i])
				fn(d)
				r.MatchFields[i] = d.DieRelease()
				return
			}
		}

		d := NodeSelectorRequirementBlank.DieImmutable(false).DieFeed(corev1.NodeSelectorRequirement{Key: key})
		fn(d)
		r.MatchFields = append(r.MatchFields, d.DieRelease())
	})
}

// +die
type _ = corev1.NodeSelectorRequirement
