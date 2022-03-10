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
	"fmt"
	"strconv"
)

func Port(port string, field string) FieldErrors {
	portNumber, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return ErrInvalidValue(port, field)
	}
	return PortNumber(int32(portNumber), field)
}

func PortNumber(port int32, field string) FieldErrors {
	errs := FieldErrors{}

	if port < 1 || port > 65535 {
		errs = errs.Also(ErrInvalidValue(fmt.Sprint(port), field))
	}

	return errs
}
