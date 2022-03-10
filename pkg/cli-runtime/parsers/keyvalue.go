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

package parsers

import (
	"strings"
)

func KeyValue(kv string) []string {
	return strings.SplitN(kv, "=", 2)
}

func DeletableKeyValue(kv string) []string {
	if strings.Contains(kv, "=") {
		return KeyValue(kv)
	}
	return []string{kv[0 : len(kv)-1]}
}
