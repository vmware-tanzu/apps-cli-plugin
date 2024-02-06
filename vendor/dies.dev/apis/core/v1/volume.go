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
)

// +die
type _ = corev1.Volume

func (d *VolumeDie) HostPathDie(fn func(d *HostPathVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := HostPathVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.HostPath)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			HostPath: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) EmptyDirDie(fn func(d *EmptyDirVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := EmptyDirVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.EmptyDir)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			EmptyDir: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) GCEPersistentDiskDie(fn func(d *GCEPersistentDiskVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := GCEPersistentDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.GCEPersistentDisk)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			GCEPersistentDisk: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) AWSElasticBlockStoreDie(fn func(d *AWSElasticBlockStoreVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := AWSElasticBlockStoreVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.AWSElasticBlockStore)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			AWSElasticBlockStore: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) GitRepoDie(fn func(d *GitRepoVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := GitRepoVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.GitRepo)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			GitRepo: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) SecretDie(fn func(d *SecretVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := SecretVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Secret)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Secret: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) NFSDie(fn func(d *NFSVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := NFSVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.NFS)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			NFS: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) ISCSIDie(fn func(d *ISCSIVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := ISCSIVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.ISCSI)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			ISCSI: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) GlusterfsDie(fn func(d *GlusterfsVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := GlusterfsVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Glusterfs)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Glusterfs: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) PersistentVolumeClaimDie(fn func(d *PersistentVolumeClaimVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := PersistentVolumeClaimVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.PersistentVolumeClaim)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) RBDDie(fn func(d *RBDVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := RBDVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.RBD)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			RBD: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) FlexVolumeDie(fn func(d *FlexVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := FlexVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.FlexVolume)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			FlexVolume: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) CinderDie(fn func(d *CinderVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := CinderVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Cinder)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Cinder: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) CephFSDie(fn func(d *CephFSVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := CephFSVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.CephFS)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			CephFS: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) FlockerDie(fn func(d *FlockerVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := FlockerVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Flocker)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Flocker: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) DownwardAPIDie(fn func(d *DownwardAPIVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := DownwardAPIVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.DownwardAPI)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			DownwardAPI: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) FCDie(fn func(d *FCVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := FCVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.FC)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			FC: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) AzureFileDie(fn func(d *AzureFileVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := AzureFileVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.AzureFile)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			AzureFile: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) ConfigMapDie(fn func(d *ConfigMapVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := ConfigMapVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.ConfigMap)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			ConfigMap: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) VsphereVolumeDie(fn func(d *VsphereVirtualDiskVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := VsphereVirtualDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.VsphereVolume)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			VsphereVolume: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) QuobyteDie(fn func(d *QuobyteVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := QuobyteVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Quobyte)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Quobyte: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) AzureDiskDie(fn func(d *AzureDiskVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := AzureDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.AzureDisk)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			AzureDisk: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) PhotonPersistentDiskDie(fn func(d *PhotonPersistentDiskVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := PhotonPersistentDiskVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.PhotonPersistentDisk)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			PhotonPersistentDisk: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) ProjectedDie(fn func(d *ProjectedVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := ProjectedVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Projected)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Projected: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) PortworxVolumeDie(fn func(d *PortworxVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := PortworxVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.PortworxVolume)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			PortworxVolume: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) ScaleIODie(fn func(d *ScaleIOVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := ScaleIOVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.ScaleIO)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			ScaleIO: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) StorageOSDie(fn func(d *StorageOSVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := StorageOSVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.StorageOS)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			StorageOS: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) CSIDie(fn func(d *CSIVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := CSIVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.CSI)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			CSI: d.DieReleasePtr(),
		}
	})
}

func (d *VolumeDie) EphemeralDie(fn func(d *EphemeralVolumeSourceDie)) *VolumeDie {
	return d.DieStamp(func(r *corev1.Volume) {
		d := EphemeralVolumeSourceBlank.DieImmutable(false).DieFeedPtr(r.Ephemeral)
		fn(d)
		r.VolumeSource = corev1.VolumeSource{
			Ephemeral: d.DieReleasePtr(),
		}
	})
}

// +die
type _ = corev1.HostPathVolumeSource

// +die
type _ = corev1.EmptyDirVolumeSource

// +die
type _ = corev1.GCEPersistentDiskVolumeSource

// +die
type _ = corev1.AWSElasticBlockStoreVolumeSource

// +die
type _ = corev1.GitRepoVolumeSource

// +die
type _ = corev1.SecretVolumeSource

func (d *SecretVolumeSourceDie) ItemDie(key string, fn func(d *KeyToPathDie)) *SecretVolumeSourceDie {
	return d.DieStamp(func(r *corev1.SecretVolumeSource) {
		for i := range r.Items {
			if key == r.Items[i].Key {
				d := KeyToPathBlank.DieImmutable(false).DieFeed(r.Items[i])
				fn(d)
				r.Items[i] = d.DieRelease()
				return
			}
		}

		d := KeyToPathBlank.DieImmutable(false).DieFeed(corev1.KeyToPath{Key: key})
		fn(d)
		r.Items = append(r.Items, d.DieRelease())
	})
}

// +die
type _ = corev1.NFSVolumeSource

// +die
type _ = corev1.ISCSIVolumeSource

func (d *ISCSIVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *ISCSIVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ISCSIVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.GlusterfsVolumeSource

// +die
type _ = corev1.PersistentVolumeClaimVolumeSource

// +die
type _ = corev1.RBDVolumeSource

func (d *RBDVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *RBDVolumeSourceDie {
	return d.DieStamp(func(r *corev1.RBDVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.FlexVolumeSource

func (d *FlexVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *FlexVolumeSourceDie {
	return d.DieStamp(func(r *corev1.FlexVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.CinderVolumeSource

func (d *CinderVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *CinderVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CinderVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.CephFSVolumeSource

func (d *CephFSVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *CephFSVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CephFSVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.FlockerVolumeSource

// +die
type _ = corev1.DownwardAPIVolumeSource

func (d *DownwardAPIVolumeSourceDie) ItemDie(path string, fn func(d *DownwardAPIVolumeFileDie)) *DownwardAPIVolumeSourceDie {
	return d.DieStamp(func(r *corev1.DownwardAPIVolumeSource) {
		for i := range r.Items {
			if path == r.Items[i].Path {
				d := DownwardAPIVolumeFileBlank.DieImmutable(false).DieFeed(r.Items[i])
				fn(d)
				r.Items[i] = d.DieRelease()
				return
			}
		}

		d := DownwardAPIVolumeFileBlank.DieImmutable(false).DieFeed(corev1.DownwardAPIVolumeFile{Path: path})
		fn(d)
		r.Items = append(r.Items, d.DieRelease())
	})
}

// +die
type _ = corev1.DownwardAPIVolumeFile

func (d *DownwardAPIVolumeFileDie) FieldRefDie(fn func(d *ObjectFieldSelectorDie)) *DownwardAPIVolumeFileDie {
	return d.DieStamp(func(r *corev1.DownwardAPIVolumeFile) {
		d := ObjectFieldSelectorBlank.DieImmutable(false).DieFeedPtr(r.FieldRef)
		fn(d)
		r.FieldRef = d.DieReleasePtr()
	})
}

func (d *DownwardAPIVolumeFileDie) ResourceFieldRefDie(fn func(d *ResourceFieldSelectorDie)) *DownwardAPIVolumeFileDie {
	return d.DieStamp(func(r *corev1.DownwardAPIVolumeFile) {
		d := ResourceFieldSelectorBlank.DieImmutable(false).DieFeedPtr(r.ResourceFieldRef)
		fn(d)
		r.ResourceFieldRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.FCVolumeSource

// +die
type _ = corev1.AzureFileVolumeSource

// +die
type _ = corev1.ConfigMapVolumeSource

func (d *ConfigMapVolumeSourceDie) Name(v string) *ConfigMapVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ConfigMapVolumeSource) {
		r.Name = v
	})
}

func (d *ConfigMapVolumeSourceDie) ItemDie(key string, fn func(d *KeyToPathDie)) *ConfigMapVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ConfigMapVolumeSource) {
		for i := range r.Items {
			if key == r.Items[i].Key {
				d := KeyToPathBlank.DieImmutable(false).DieFeed(r.Items[i])
				fn(d)
				r.Items[i] = d.DieRelease()
				return
			}
		}

		d := KeyToPathBlank.DieImmutable(false).DieFeed(corev1.KeyToPath{Key: key})
		fn(d)
		r.Items = append(r.Items, d.DieRelease())
	})
}

// +die
type _ = corev1.VsphereVirtualDiskVolumeSource

// +die
type _ = corev1.QuobyteVolumeSource

// +die
type _ = corev1.AzureDiskVolumeSource

// +die
type _ = corev1.PhotonPersistentDiskVolumeSource

// +die
type _ = corev1.ProjectedVolumeSource

func (d *ProjectedVolumeSourceDie) SourcesDie(sources ...*VolumeProjectionDie) *ProjectedVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ProjectedVolumeSource) {
		r.Sources = make([]corev1.VolumeProjection, len(sources))
		for i := range sources {
			r.Sources[i] = sources[i].DieRelease()
		}
	})
}

// +die
type _ = corev1.VolumeProjection

func (d *VolumeProjectionDie) SecretDie(fn func(d *SecretProjectionDie)) *VolumeProjectionDie {
	return d.DieStamp(func(r *corev1.VolumeProjection) {
		d := SecretProjectionBlank.DieImmutable(false).DieFeedPtr(r.Secret)
		fn(d)
		r.Secret = d.DieReleasePtr()
	})
}

func (d *VolumeProjectionDie) DownwardAPIDie(fn func(d *DownwardAPIProjectionDie)) *VolumeProjectionDie {
	return d.DieStamp(func(r *corev1.VolumeProjection) {
		d := DownwardAPIProjectionBlank.DieImmutable(false).DieFeedPtr(r.DownwardAPI)
		fn(d)
		r.DownwardAPI = d.DieReleasePtr()
	})
}

func (d *VolumeProjectionDie) ConfigMapDie(fn func(d *ConfigMapProjectionDie)) *VolumeProjectionDie {
	return d.DieStamp(func(r *corev1.VolumeProjection) {
		d := ConfigMapProjectionBlank.DieImmutable(false).DieFeedPtr(r.ConfigMap)
		fn(d)
		r.ConfigMap = d.DieReleasePtr()
	})
}

func (d *VolumeProjectionDie) ServiceAccountTokenDie(fn func(d *ServiceAccountTokenProjectionDie)) *VolumeProjectionDie {
	return d.DieStamp(func(r *corev1.VolumeProjection) {
		d := ServiceAccountTokenProjectionBlank.DieImmutable(false).DieFeedPtr(r.ServiceAccountToken)
		fn(d)
		r.ServiceAccountToken = d.DieReleasePtr()
	})
}

func (d *VolumeProjectionDie) ClusterTrustBundleDie(fn func(d *ClusterTrustBundleProjectionDie)) *VolumeProjectionDie {
	return d.DieStamp(func(r *corev1.VolumeProjection) {
		d := ClusterTrustBundleProjectionBlank.DieImmutable(false).DieFeedPtr(r.ClusterTrustBundle)
		fn(d)
		r.ClusterTrustBundle = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.SecretProjection

func (d *SecretProjectionDie) Name(v string) *SecretProjectionDie {
	return d.DieStamp(func(r *corev1.SecretProjection) {
		r.Name = v
	})
}

func (d *SecretProjectionDie) ItemDie(key string, fn func(d *KeyToPathDie)) *SecretProjectionDie {
	return d.DieStamp(func(r *corev1.SecretProjection) {
		for i := range r.Items {
			if key == r.Items[i].Key {
				d := KeyToPathBlank.DieImmutable(false).DieFeed(r.Items[i])
				fn(d)
				r.Items[i] = d.DieRelease()
				return
			}
		}

		d := KeyToPathBlank.DieImmutable(false).DieFeed(corev1.KeyToPath{Key: key})
		fn(d)
		r.Items = append(r.Items, d.DieRelease())
	})
}

// +die
type _ = corev1.DownwardAPIProjection

func (d *DownwardAPIProjectionDie) ItemDie(path string, fn func(d *DownwardAPIVolumeFileDie)) *DownwardAPIProjectionDie {
	return d.DieStamp(func(r *corev1.DownwardAPIProjection) {
		for i := range r.Items {
			if path == r.Items[i].Path {
				d := DownwardAPIVolumeFileBlank.DieImmutable(false).DieFeed(r.Items[i])
				fn(d)
				r.Items[i] = d.DieRelease()
				return
			}
		}

		d := DownwardAPIVolumeFileBlank.DieImmutable(false).DieFeed(corev1.DownwardAPIVolumeFile{Path: path})
		fn(d)
		r.Items = append(r.Items, d.DieRelease())
	})
}

// +die
type _ = corev1.ConfigMapProjection

func (d *ConfigMapProjectionDie) Name(v string) *ConfigMapProjectionDie {
	return d.DieStamp(func(r *corev1.ConfigMapProjection) {
		r.Name = v
	})
}

func (d *ConfigMapProjectionDie) ItemDie(key string, fn func(d *KeyToPathDie)) *ConfigMapProjectionDie {
	return d.DieStamp(func(r *corev1.ConfigMapProjection) {
		for i := range r.Items {
			if key == r.Items[i].Key {
				d := KeyToPathBlank.DieImmutable(false).DieFeed(r.Items[i])
				fn(d)
				r.Items[i] = d.DieRelease()
				return
			}
		}

		d := KeyToPathBlank.DieImmutable(false).DieFeed(corev1.KeyToPath{Key: key})
		fn(d)
		r.Items = append(r.Items, d.DieRelease())
	})
}

// +die
type _ = corev1.ServiceAccountTokenProjection

// +die
type _ = corev1.ClusterTrustBundleProjection

func (d *ClusterTrustBundleProjectionDie) LabelSelectorDie(fn func(d *diemetav1.LabelSelectorDie)) *ClusterTrustBundleProjectionDie {
	return d.DieStamp(func(r *corev1.ClusterTrustBundleProjection) {
		d := diemetav1.LabelSelectorBlank.DieImmutable(false).DieFeedPtr(r.LabelSelector)
		fn(d)
		r.LabelSelector = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.PortworxVolumeSource

// +die
type _ = corev1.ScaleIOVolumeSource

func (d *ScaleIOVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *ScaleIOVolumeSourceDie {
	return d.DieStamp(func(r *corev1.ScaleIOVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.StorageOSVolumeSource

func (d *StorageOSVolumeSourceDie) SecretRefDie(fn func(d *LocalObjectReferenceDie)) *StorageOSVolumeSourceDie {
	return d.DieStamp(func(r *corev1.StorageOSVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.SecretRef)
		fn(d)
		r.SecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.CSIVolumeSource

func (d *CSIVolumeSourceDie) VolumeAttribute(key, value string) *CSIVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIVolumeSource) {
		r.VolumeAttributes[key] = value
	})
}

func (d *CSIVolumeSourceDie) NodePublishSecretRefDie(fn func(d *LocalObjectReferenceDie)) *CSIVolumeSourceDie {
	return d.DieStamp(func(r *corev1.CSIVolumeSource) {
		d := LocalObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.NodePublishSecretRef)
		fn(d)
		r.NodePublishSecretRef = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.EphemeralVolumeSource

func (d *EphemeralVolumeSourceDie) VolumeClaimTemplateDie(fn func(d *PersistentVolumeClaimTemplateDie)) *EphemeralVolumeSourceDie {
	return d.DieStamp(func(r *corev1.EphemeralVolumeSource) {
		d := PersistentVolumeClaimTemplateBlank.DieImmutable(false).DieFeedPtr(r.VolumeClaimTemplate)
		fn(d)
		r.VolumeClaimTemplate = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.KeyToPath
