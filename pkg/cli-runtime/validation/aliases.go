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
	"github.com/vmware-labs/reconciler-runtime/validation"
)

const CurrentField = validation.CurrentField

type FieldValidator = validation.FieldValidator
type FieldErrors = validation.FieldErrors
type Validatable = validation.Validatable

var ErrDisallowedFields = validation.ErrDisallowedFields
var ErrInvalidArrayValue = validation.ErrInvalidArrayValue
var ErrInvalidValue = validation.ErrInvalidValue
var ErrDuplicateValue = validation.ErrDuplicateValue
var ErrMissingField = validation.ErrMissingField
var ErrMissingOneOf = validation.ErrMissingOneOf
var ErrMultipleOneOf = validation.ErrMultipleOneOf
