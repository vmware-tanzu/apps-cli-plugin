/*
Copyright 2023 VMware, Inc.

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

package lsp

type HealthStatus struct {
	UserHasPermission     bool   `json:"user_has_permission" yaml:"user_has_permission"`
	Reachable             bool   `json:"reachable" yaml:"reachable"`
	UpstreamAuthenticated bool   `json:"upstream_authenticated" yaml:"upstream_authenticated"`
	OverallHealth         bool   `json:"overall_health" yaml:"overall_health"`
	Message               string `json:"message" yaml:"message"`
}
