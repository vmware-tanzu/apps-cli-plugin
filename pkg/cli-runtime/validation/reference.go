/*
Copyright 2019 VMware, Inc.

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

package validation

import (
	"strings"
)

func ObjectReference(ref, field string) FieldErrors {
	errs := FieldErrors{}

	parts := strings.Split(ref, ":")
	if len(parts) != 3 && len(parts) != 4 {
		errs = errs.Also(ErrInvalidValue(ref, field))
		return errs
	}
	if parts[0] != "v1" && !strings.Contains(parts[0], "/") {
		errs = errs.Also(ErrInvalidValue(ref, field))
	}
	if parts[1] == "" {
		errs = errs.Also(ErrInvalidValue(ref, field))
	}
	errs = errs.Also(K8sName(parts[2], field))

	if len(parts) == 4 {
		errs = errs.Also(K8sName(parts[3], field))
	}
	return errs
}

func DeletableObjectReference(ref, field string) FieldErrors {
	if strings.HasSuffix(ref, "-") {
		ref = ref[0 : len(ref)-1]
	}
	return ObjectReference(ref, field)
}

func ObjectReferences(refs []string, field string) FieldErrors {
	errs := FieldErrors{}

	for i, ref := range refs {
		errs = errs.Also(ObjectReference(ref, CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}

func DeletableObjectReferences(kvs []string, field string) FieldErrors {
	errs := FieldErrors{}

	for i, kv := range kvs {
		errs = errs.Also(DeletableObjectReference(kv, CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}

func DeletableKeyObjectReference(ref, field string) FieldErrors {
	parts := strings.Split(ref, "=")
	errs := FieldErrors{}
	if len(parts) == 2 {
		errs = errs.Also(K8sName(parts[0], field))
		errs = errs.Also(ObjectReference(parts[1], field))
	} else if !strings.HasSuffix(ref, "-") {
		errs = errs.Also(ErrInvalidValue(ref, field))
	}
	return errs
}

func DeletableKeyObjectReferences(kvs []string, field string) FieldErrors {
	errs := FieldErrors{}

	for i, kv := range kvs {
		errs = errs.Also(DeletableKeyObjectReference(kv, CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}
