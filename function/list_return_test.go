// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestListReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListReturn
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.ListReturn{
				ElementType: basetypes.StringType{},
			},
			expected: basetypes.ListType{
				ElemType: basetypes.StringType{},
			},
		},
		"CustomType": {
			parameter: function.ListReturn{
				CustomType: testtypes.ListType{
					ListType: basetypes.ListType{
						ElemType: basetypes.StringType{},
					},
				},
			},
			expected: testtypes.ListType{
				ListType: basetypes.ListType{
					ElemType: basetypes.StringType{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListReturnValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		returnDef function.ListReturn
		request   function.ValidateReturnImplementationRequest
		expected  *function.ValidateReturnImplementationResponse
	}{
		"customtype": {
			returnDef: function.ListReturn{
				CustomType: testtypes.ListType{},
			},
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"elementtype": {
			returnDef: function.ListReturn{
				ElementType: types.StringType,
			},
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"elementtype-dynamic": {
			returnDef: function.ListReturn{
				ElementType: types.DynamicType,
			},
			request: function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Return contains a collection type with a nested dynamic type. "+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &function.ValidateReturnImplementationResponse{}
			testCase.returnDef.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
