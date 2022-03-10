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

func KeyValue(kv, field string) FieldErrors {
	errs := FieldErrors{}

	if strings.HasPrefix(kv, "=") || !strings.Contains(kv, "=") {
		errs = errs.Also(ErrInvalidValue(kv, field))
	}

	return errs
}

func DeletableKeyValue(kv, field string) FieldErrors {
	if strings.Contains(kv, "=") {
		return KeyValue(kv, field)
	}

	errs := FieldErrors{}

	if !strings.HasSuffix(kv, "-") {
		errs = errs.Also(ErrInvalidValue(kv, field))
	}

	return errs
}

func KeyValues(kvs []string, field string) FieldErrors {
	errs := FieldErrors{}

	for i, kv := range kvs {
		errs = errs.Also(KeyValue(kv, CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}

func DeletableKeyValues(kvs []string, field string) FieldErrors {
	errs := FieldErrors{}

	for i, kv := range kvs {
		errs = errs.Also(DeletableKeyValue(kv, CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}
