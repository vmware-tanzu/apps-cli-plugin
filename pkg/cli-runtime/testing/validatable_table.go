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

package testing

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
)

type ValidatableTestSuite []ValidatableTestCase

type ValidatableTestCase struct {
	// Name is used to identify the record in the test results. A sub-test is created for each
	// record with this name.
	Name string
	// Skip suppresses the execution of this test record.
	Skip bool
	// Focus executes this record skipping all unfocused records. The containing test will fail to
	// prevent accidental check-in.
	Focus bool

	// inputs

	// Validatable to validate
	Validatable validation.Validatable

	// outputs

	// ExpectFieldErrors are the errors that should be returned from the validator.
	ExpectFieldErrors validation.FieldErrors

	// ShouldValidate is true if the validatable object is valid
	ShouldValidate bool

	// lifecycle

	// Prepare is called before the command is executed. It is intended to prepare that broader
	// environment before the specific table record is executed. For example, changing the working
	// directory or setting mock expectations.
	Prepare func(context.Context, *testing.T) (context.Context, error)
	// CleanUp is called after the table record is finished and all defined assertions complete.
	// It is intended to clean up any state created in the Prepare step or during the test
	// execution, or to make assertions for mocks.
	CleanUp func(context.Context, *testing.T) error
}

func (ts ValidatableTestSuite) Run(t *testing.T) {
	ctx := context.Background()
	focused := ValidatableTestSuite{}
	for _, tc := range ts {
		if tc.Focus && !tc.Skip {
			focused = append(focused, tc)
		}
	}
	if len(focused) != 0 {
		for _, tc := range focused {
			tc.Run(t, ctx)
		}
		t.Errorf("test run focused on %d record(s), skipped %d record(s)", len(focused), len(ts)-len(focused))
		return
	}

	for _, tc := range ts {
		tc.Run(t, ctx)
	}
}

func (tc ValidatableTestCase) Run(t *testing.T, ctx context.Context) {
	t.Run(tc.Name, func(t *testing.T) {
		if tc.Skip {
			t.SkipNow()
		}

		if tc.Prepare != nil {
			var err error
			if ctx, err = tc.Prepare(ctx, t); err != nil {
				t.Errorf("unexpected error in prepare: %v", err)
			}
		}

		errs := tc.Validatable.Validate(ctx)

		if tc.ExpectFieldErrors != nil {
			actual := errs
			expected := tc.ExpectFieldErrors
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected errors (-expected, +actual): %s", diff)
			}
		}

		if expected, actual := tc.ShouldValidate, len(errs) == 0; expected != actual {
			if expected {
				t.Errorf("expected to validate, actually %q", errs)
			} else {
				t.Errorf("expected not to validate, actually %q", errs)
			}
		}

		if !tc.ShouldValidate && len(tc.ExpectFieldErrors) == 0 {
			t.Error("one of ShouldValidate=true or ExpectFieldErrors is required")
		}

		if tc.CleanUp != nil {
			var err error
			if err = tc.CleanUp(ctx, t); err != nil {
				t.Errorf("unexpected error in clean up: %v", err)
			}
		}
	})
}
