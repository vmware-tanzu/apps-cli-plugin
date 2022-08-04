/*
Copyright 2021 VMware, Inc.

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

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	servicesv1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/services/v1alpha1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
)

const (
	WorkloadConditionReady  = "Ready"
	WorkloadAnnotationParam = "annotations"
	WorkloadMavenParam      = "maven"
)

type MavenSource struct {
	ArtifactId string  `json:"artifactId"`
	GroupId    string  `json:"groupId"`
	Version    string  `json:"version"`
	Type       *string `json:"type"`
}

func (w *Workload) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Workload")
}

func (w *Workload) Load(in io.Reader) error {
	if err := w.loadAndValidateDocuments(in); err != nil {
		return err
	}

	if apiVersion, kind := SchemeGroupVersion.Identifier(), "Workload"; w.APIVersion != apiVersion || w.Kind != kind {
		return fmt.Errorf("file must contain resource with API Version %q and Kind %q", apiVersion, kind)
	}
	w.APIVersion = ""
	w.Kind = ""
	return nil
}

func (w *Workload) loadAndValidateDocuments(in io.Reader) error {
	d := yaml.NewYAMLOrJSONDecoder(in, 4096)
	documents := 0
	for {
		var workload *Workload
		if err := d.Decode(&workload); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if workload == nil {
			continue
		}
		if documents > 0 {
			return fmt.Errorf("files containing multiple workload descriptions are not supported")
		}
		workload.DeepCopyInto(w)
		documents++
	}
	return nil
}

func (w *WorkloadSpec) MergeServiceAccountName(serviceAccountName string) {
	var serviceAccountNamePtr *string
	if serviceAccountName != "" {
		serviceAccountNamePtr = &serviceAccountName
	}
	w.ServiceAccountName = serviceAccountNamePtr
}

func (w *Workload) Merge(updates *Workload) {
	for k, v := range updates.Annotations {
		w.MergeAnnotations(k, v)
	}
	for k, v := range updates.Labels {
		w.MergeLabels(k, v)
	}
	w.Spec.Merge(&updates.Spec)
}

func (w *WorkloadSpec) GetMavenSource() *MavenSource {
	currentMaven := &MavenSource{}
	w.GetParam("maven", currentMaven)
	return currentMaven
}

func (w *Workload) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	errs = errs.Also(validation.K8sName(w.Name, cli.NameArgumentName))
	errs = errs.Also(validation.K8sName(w.Namespace, flags.NamespaceFlagName))
	errs = errs.Also(w.Spec.Validate())

	return errs
}

func (w *Workload) IsSourceFound() bool {
	if w.Spec.Source == nil && w.Spec.Image == "" {
		return false
	}
	if w.Spec.Source != nil {
		if w.Spec.Source.Git == nil && w.Spec.Source.Image == "" {
			return false
		}
	}
	return true
}

func (w *WorkloadSpec) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}
	if w.Source != nil {
		if w.Image != "" && w.Source.Subpath == "" {
			errs = errs.Also(validation.ErrMultipleOneOf(flags.GitFlagWildcard, flags.SourceImageFlagName, flags.ImageFlagName))
		}

		errs = errs.Also(w.Source.Validate())

		if w.Source.Git == nil && w.Source.Image == "" && w.Image == "" && w.Source.Subpath != "" {
			errs = errs.Also(validation.ErrInvalidValue(w.Source.Subpath, flags.SubPathFlagName))
		}
	}

	errs = errs.Also(w.ValidateMavenSource())

	return errs
}

func (w *WorkloadSpec) ValidateMavenSource() validation.FieldErrors {
	errs := validation.FieldErrors{}

	mavenParam := MavenSource{}
	w.GetParam(WorkloadMavenParam, &mavenParam)
	if !(MavenSource{} == mavenParam) {
		if mavenParam.ArtifactId == "" {
			errs = errs.Also(validation.ErrMissingField(cli.StripDash(flags.MavenArtifactFlagName)))
		}
		if mavenParam.GroupId == "" {
			errs = errs.Also(validation.ErrMissingField(cli.StripDash(flags.MavenGroupFlagName)))
		}
		if mavenParam.Version == "" {
			errs = errs.Also(validation.ErrMissingField(cli.StripDash(flags.MavenVersionFlagName)))
		}
	}

	return errs
}

func (w *Source) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	if w.Git != nil && w.Image != "" {
		errs = errs.Also(validation.ErrMultipleOneOf(flags.GitFlagWildcard, flags.SourceImageFlagName))
	}

	if w.Git != nil {
		errs = errs.Also(w.Git.Validate())
	}

	return errs
}

func (w *GitSource) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	if w.URL == "" {
		errs = errs.Also(validation.ErrMissingField(flags.GitRepoFlagName))
	}

	if w.Ref.Branch == "" && w.Ref.Tag == "" {
		errs = errs.Also(validation.ErrMissingOneOf(flags.GitBranchFlagName, flags.GitTagFlagName))
	}

	return errs
}

func (w *WorkloadSpec) Merge(updates *WorkloadSpec) {
	for _, p := range updates.Params {
		w.MergeParams(p.Name, p.Value)
	}
	if updates.Image != "" {
		w.MergeImage(updates.Image)
	}
	if s := updates.Source; s != nil {
		if s.Git != nil {
			w.MergeGit(*s.Git)
		}
		if s.Image != "" {
			w.MergeSourceImage(s.Image)
		}
		if s.Subpath != "" {
			w.MergeSubPath(s.Subpath)
		}
	}
	for _, e := range updates.Env {
		w.MergeEnv(e)
	}
	w.MergeResources(updates.Resources)
	for _, s := range updates.ServiceClaims {
		w.MergeServiceClaim(s)
	}
	if updates.Build != nil {
		for _, e := range updates.Build.Env {
			w.MergeBuildEnv(e)
		}
	}
}

func (w *WorkloadSpec) MergeParams(key string, value interface{}) {
	b, _ := json.Marshal(value)
	param := Param{
		Name:  key,
		Value: apiextensionsv1.JSON{Raw: b},
	}
	for i := range w.Params {
		if w.Params[i].Name == param.Name {
			w.Params[i] = param
			return
		}
	}
	w.Params = append(w.Params, param)
}

func (w *WorkloadSpec) RemoveParam(name string) {
	params := []Param{}
	for i := range w.Params {
		if w.Params[i].Name != name {
			params = append(params, w.Params[i])
		}
	}
	w.Params = params
}

func (w *WorkloadSpec) GetParam(key string, value interface{}) {
	for _, p := range w.Params {
		if p.Name == key {
			json.Unmarshal(p.Value.Raw, value)
		}
	}
}

func (w *WorkloadSpec) MergeAnnotationParams(key string, value string) {
	annotations := make(map[string]string)
	w.GetParam(WorkloadAnnotationParam, &annotations)
	annotations[key] = value
	w.MergeParams(WorkloadAnnotationParam, annotations)
}

func (w *WorkloadSpec) RemoveAnnotationParams(name string) {
	annotations := make(map[string]string)
	w.GetParam(WorkloadAnnotationParam, &annotations)
	delete(annotations, name)
	if len(annotations) == 0 {
		w.RemoveParam(WorkloadAnnotationParam)
	} else {
		w.MergeParams(WorkloadAnnotationParam, annotations)
	}
}

func (w *WorkloadSpec) MergeMavenSource(updates MavenSource) {
	currentMaven := w.GetMavenSource()
	if updates.ArtifactId != "" {
		currentMaven.ArtifactId = updates.ArtifactId
	}
	if updates.GroupId != "" {
		currentMaven.GroupId = updates.GroupId
	}
	if updates.Version != "" {
		currentMaven.Version = updates.Version
	}
	w.MergeParams(WorkloadMavenParam, currentMaven)
}

func (w *WorkloadSpec) ResetSource() {
	w.Source = nil
	w.Image = ""
}

func (w *WorkloadSpec) MergeGit(git GitSource) {
	stash := w.Source
	w.ResetSource()

	w.Source = &Source{
		Git: &git,
	}
	if stash != nil && stash.Git != nil {
		if w.Source.Git.URL == "" {
			w.Source.Git.URL = stash.Git.URL
		}
		if w.Source.Git.Ref.Branch == "" {
			w.Source.Git.Ref.Branch = stash.Git.Ref.Branch
		}
	}
}

func (w *WorkloadSpec) MergeSourceImage(image string) {
	w.ResetSource()

	w.Source = &Source{
		Image: image,
	}
}

func (w *WorkloadSpec) MergeSubPath(subPath string) {
	if w.Source == nil {
		w.Source = &Source{}
	}

	w.Source.Subpath = subPath
}

func (w *WorkloadSpec) MergeImage(image string) {
	w.ResetSource()

	w.Image = image
}

func (w *WorkloadSpec) MergeEnv(env corev1.EnvVar) {
	for i := range w.Env {
		if w.Env[i].Name == env.Name {
			w.Env[i] = env
			return
		}
	}
	w.Env = append(w.Env, env)
}

func (w *WorkloadSpec) RemoveEnv(name string) {
	env := []corev1.EnvVar{}
	for i := range w.Env {
		if w.Env[i].Name != name {
			env = append(env, w.Env[i])
		}
	}
	w.Env = env
}

func (w *WorkloadSpec) MergeResources(r *corev1.ResourceRequirements) {
	if r == nil {
		return
	}
	if w.Resources == nil {
		w.Resources = &corev1.ResourceRequirements{}
	}
	if r.Limits != nil {
		if w.Resources.Limits == nil {
			w.Resources.Limits = corev1.ResourceList{}
		}
		for k, v := range r.Limits {
			w.Resources.Limits[k] = v
		}
	}
	if r.Requests != nil {
		if w.Resources.Requests == nil {
			w.Resources.Requests = corev1.ResourceList{}
		}
		for k, v := range r.Requests {
			w.Resources.Requests[k] = v
		}
	}
}

func (w *Workload) MergeAnnotations(key, value string) {
	if w.Annotations == nil {
		w.Annotations = map[string]string{}
	}
	w.Annotations[key] = value
}

func (w *Workload) MergeLabels(key, value string) {
	if w.Labels == nil {
		w.Labels = map[string]string{}
	}
	w.Labels[key] = value
}

func (w *Workload) MergeServiceClaimAnnotation(name string, value interface{}) {
	annotationServiceClaims, err := servicesv1alpha1.NewServiceClaimWorkloadConfigFromAnnotation(w.GetAnnotations()[apis.ServiceClaimAnnotationName])
	if err != nil {
		return
	}

	annotationServiceClaims.AddServiceClaim(name, value)
	if len(annotationServiceClaims.Spec.ServiceClaims) > 0 {
		w.MergeAnnotations(apis.ServiceClaimAnnotationName, annotationServiceClaims.Annotation())
	}
}

func (w *Workload) DeleteServiceClaimAnnotation(name string) {
	annotationServiceClaims, err := servicesv1alpha1.NewServiceClaimWorkloadConfigFromAnnotation(w.GetAnnotations()[apis.ServiceClaimAnnotationName])
	if err != nil {
		return
	}

	sc := servicesv1alpha1.NewServiceClaimWorkloadConfig()

	for claimName, claimValue := range annotationServiceClaims.Spec.ServiceClaims {
		if claimName != name {
			sc.AddServiceClaim(claimName, claimValue)
		}
	}

	if len(sc.Spec.ServiceClaims) > 0 {
		w.MergeAnnotations(apis.ServiceClaimAnnotationName, sc.Annotation())
	} else {
		delete(w.GetAnnotations(), apis.ServiceClaimAnnotationName)
	}
}

func (w *WorkloadSpec) DeleteServiceClaim(name string) {
	serviceClaims := []WorkloadServiceClaim{}
	for _, sc := range w.ServiceClaims {
		if sc.Name != name {
			serviceClaims = append(serviceClaims, sc)
		}
	}
	w.ServiceClaims = serviceClaims
}

func (w *WorkloadSpec) MergeServiceClaim(sc WorkloadServiceClaim) {
	for i := range w.ServiceClaims {
		if sc.Name == w.ServiceClaims[i].Name {
			w.ServiceClaims[i] = sc
			return
		}
	}
	w.ServiceClaims = append(w.ServiceClaims, sc)
}

func NewServiceClaim(name string, serviceRef corev1.ObjectReference) WorkloadServiceClaim {
	return WorkloadServiceClaim{
		Name: name,
		Ref: &WorkloadServiceClaimReference{
			APIVersion: serviceRef.APIVersion,
			Kind:       serviceRef.Kind,
			Name:       serviceRef.Name,
		},
	}
}

func (w *WorkloadSpec) RemoveBuildEnv(name string) {
	env := []corev1.EnvVar{}
	if w.Build != nil {
		for i := range w.Build.Env {
			if w.Build.Env[i].Name != name {
				env = append(env, w.Build.Env[i])
			}
		}
		if len(env) > 0 {
			w.Build.Env = env
		} else {
			w.Build = nil
		}
	}
}

func (w *WorkloadSpec) MergeBuildEnv(env corev1.EnvVar) {
	if w.Build == nil {
		w.Build = &WorkloadBuild{}
	}
	for i := range w.Build.Env {
		if w.Build.Env[i].Name == env.Name {
			w.Build.Env[i] = env
			return
		}
	}
	w.Build.Env = append(w.Build.Env, env)
}

func WorkloadReadyConditionFunc(target client.Object) (bool, error) {
	obj, ok := target.(*Workload)
	if !ok {
		return false, nil
	}
	if obj.Generation != obj.Status.ObservedGeneration {
		return false, nil
	}
	for _, cond := range obj.Status.Conditions {
		if cond.Type == WorkloadConditionReady {
			if cond.Status == metav1.ConditionTrue {
				return true, nil
			}
			if cond.Status == metav1.ConditionFalse {
				return true, fmt.Errorf("Failed to become ready: %s", cond.Message)
			}
		}
	}
	return false, nil
}

func (w *Workload) DeprecationWarnings() []string {
	warnings := []string{}
	var serviceClaimDeprecationWarningMsg = "Cross namespace service claims are deprecated. Please use `tanzu service claim create` instead."

	if sc := w.GetAnnotations()[apis.ServiceClaimAnnotationName]; sc != "" {
		warnings = append(warnings, serviceClaimDeprecationWarningMsg)
	}
	return warnings
}

type workloadNoticeStashKey struct{}

func StashWorkloadNotice(ctx context.Context, notice string) context.Context {
	res := RetrieveStashNotice(ctx)
	if res != nil {
		res = append(res, notice)
	} else {
		res = []string{notice}
	}
	return context.WithValue(ctx, workloadNoticeStashKey{}, res)
}

func RetrieveStashNotice(ctx context.Context) []string {
	noticeSet, ok := ctx.Value(workloadNoticeStashKey{}).([]string)
	if !ok {
		return nil
	}
	return noticeSet
}

func (w *Workload) GetNoticeMsgs(ctx context.Context) []string {
	msgs := []string{}
	if res := RetrieveStashNotice(ctx); res != nil {
		msgs = append(msgs, res...)
	}

	return msgs
}
