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

package flags

import "strings"

var (
	EnvironmentVariablePrefix = "TANZU_APPS"
	EnvVarAllowedList         = map[string]bool{
		FlagToEnvVar(RegistryCertFlagName):     true,
		FlagToEnvVar(RegistryPasswordFlagName): true,
		FlagToEnvVar(RegistryTokenFlagName):    true,
		FlagToEnvVar(RegistryUsernameFlagName): true,
		FlagToEnvVar(SourceImageFlagName):      true,
		FlagToEnvVar(TypeFlagName):             true,
	}
)

func FlagToEnvVar(name string) string {
	ev := strings.ToUpper(name)
	if strings.HasPrefix(ev, "--") {
		ev = ev[2:]
	}
	ev = strings.ReplaceAll(ev, "-", "_")
	return EnvironmentVariablePrefix + "_" + strings.ToUpper(ev)
}
