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

func EnvVar(env, field string) FieldErrors {
	return KeyValue(env, field)
}

func DeletableEnvVar(env, field string) FieldErrors {
	return DeletableKeyValue(env, field)
}

func EnvVars(envs []string, field string) FieldErrors {
	return KeyValues(envs, field)
}

func DeletableEnvVars(envs []string, field string) FieldErrors {
	return DeletableKeyValues(envs, field)
}

func EnvVarFrom(env, field string) FieldErrors {
	errs := FieldErrors{}

	parts := strings.SplitN(env, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		errs = errs.Also(ErrInvalidValue(env, field))
	} else {
		value := strings.SplitN(parts[1], ":", 3)
		if len(value) != 3 {
			errs = errs.Also(ErrInvalidValue(env, field))
		} else if value[0] != "configMapKeyRef" && value[0] != "secretKeyRef" {
			errs = errs.Also(ErrInvalidValue(env, field))
		} else if value[1] == "" {
			errs = errs.Also(ErrInvalidValue(env, field))
		}
	}

	return errs
}

func EnvVarFroms(envs []string, field string) FieldErrors {
	errs := FieldErrors{}

	for i, env := range envs {
		errs = errs.Also(EnvVarFrom(env, CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}
