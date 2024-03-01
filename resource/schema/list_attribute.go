// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the desired interfaces.
var (
	_ Attribute                                    = ListAttribute{}
	_ fwschema.AttributeWithValidateImplementation = ListAttribute{}
	_ fwschema.AttributeWithListDefaultValue       = ListAttribute{}
	_ fwxschema.AttributeWithListPlanModifiers     = ListAttribute{}
	_ fwxschema.AttributeWithListValidators        = ListAttribute{}
)

// ListAttribute represents a schema attribute that is a list with a single
// element type. When retrieving the value for this attribute, use types.List
// as the value type unless the CustomType field is set. The ElementType field
// must be set.
//
// Use ListNestedAttribute if the underlying elements should be objects and
// require definition beyond type information.
//
// Terraform configurations configure this attribute using expressions that
// return a list or directly via square brace syntax.
//
//	# list of strings
//	example_attribute = ["first", "second"]
//
// Terraform configurations reference this attribute using expressions that
// accept a list or an element directly via square brace 0-based index syntax:
//
//	# first known element
//	.example_attribute[0]
type ListAttribute struct {
	// ElementType is the type for all elements of the list. This field must be
	// set.
	ElementType attr.Type

	// CustomType enables the use of a custom attribute type in place of the
	// default basetypes.ListType. When retrieving data, the basetypes.ListValuable
	// associated with this custom type must be used in place of types.List.
	CustomType basetypes.ListTypable

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
	Required bool

	// Optional indicates whether the practitioner can choose to enter a value
	// for this attribute or not. Optional and Required cannot both be true.
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this Attribute or not. Required and Computed cannot both be true. If
	// Required and Optional are both false, Computed must be true, and the
	// attribute will be considered "read only" for the practitioner, with
	// only the provider able to set its value.
	Computed bool

	// Sensitive indicates whether the value of this attribute should be
	// considered sensitive data. Setting it to true will obscure the value
	// in CLI output. Sensitive does not impact how values are stored, and
	// practitioners are encouraged to store their state as if the entire
	// file is sensitive.
	Sensitive bool

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// DeprecationMessage defines warning diagnostic details to display when
	// practitioner configurations use this Attribute. The warning diagnostic
	// summary is automatically set to "Attribute Deprecated" along with
	// configuration source file and line information.
	//
	// Set this field to a practitioner actionable message such as:
	//
	//  - "Configure other_attribute instead. This attribute will be removed
	//    in the next major version of the provider."
	//  - "Remove this attribute's configuration as it no longer is used and
	//    the attribute will be removed in the next major version of the
	//    provider."
	//
	// In Terraform 1.2.7 and later, this warning diagnostic is displayed any
	// time a practitioner attempts to configure a value for this attribute and
	// certain scenarios where this attribute is referenced.
	//
	// In Terraform 1.2.6 and earlier, this warning diagnostic is only
	// displayed when the Attribute is Required or Optional, and if the
	// practitioner configuration sets the value to a known or unknown value
	// (which may eventually be null). It has no effect when the Attribute is
	// Computed-only (read-only; not Required or Optional).
	//
	// Across any Terraform version, there are no warnings raised for
	// practitioner configuration values set directly to null, as there is no
	// way for the framework to differentiate between an unset and null
	// configuration due to how Terraform sends configuration information
	// across the protocol.
	//
	// Additional information about deprecation enhancements for read-only
	// attributes can be found in:
	//
	//  - https://github.com/hashicorp/terraform/issues/7569
	//
	DeprecationMessage string

	// Validators define value validation functionality for the attribute. All
	// elements of the slice of AttributeValidator are run, regardless of any
	// previous error diagnostics.
	//
	// Many common use case validators can be found in the
	// github.com/hashicorp/terraform-plugin-framework-validators Go module.
	//
	// If the Type field points to a custom type that implements the
	// xattr.TypeWithValidate interface, the validators defined in this field
	// are run in addition to the validation defined by the type.
	Validators []validator.List

	// PlanModifiers defines a sequence of modifiers for this attribute at
	// plan time. Schema-based plan modifications occur before any
	// resource-level plan modifications.
	//
	// Schema-based plan modifications can adjust Terraform's plan by:
	//
	//  - Requiring resource recreation. Typically used for configuration
	//    updates which cannot be done in-place.
	//  - Setting the planned value. Typically used for enhancing the plan
	//    to replace unknown values. Computed must be true or Terraform will
	//    return an error. If the plan value is known due to a known
	//    configuration value, the plan value cannot be changed or Terraform
	//    will return an error.
	//
	// Any errors will prevent further execution of this sequence or modifiers.
	PlanModifiers []planmodifier.List

	// Default defines a proposed new state (plan) value for the attribute
	// if the configuration value is null. Default prevents the framework
	// from automatically marking the value as unknown during planning when
	// other proposed new state changes are detected. If the attribute is
	// computed and the value could be altered by other changes then a default
	// should be avoided and a plan modifier should be used instead.
	Default defaults.List
}

// ApplyTerraform5AttributePathStep returns the result of stepping into a list
// index or an error.
func (a ListAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal returns true if the given Attribute is a ListAttribute
// and all fields are equal.
func (a ListAttribute) Equal(o fwschema.Attribute) bool {
	if _, ok := o.(ListAttribute); !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage returns the DeprecationMessage field value.
func (a ListAttribute) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription returns the Description field value.
func (a ListAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (a ListAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType returns types.ListType or the CustomType field value if defined.
func (a ListAttribute) GetType() attr.Type {
	if a.CustomType != nil {
		return a.CustomType
	}

	return types.ListType{
		ElemType: a.ElementType,
	}
}

// IsComputed returns the Computed field value.
func (a ListAttribute) IsComputed() bool {
	return a.Computed
}

// IsOptional returns the Optional field value.
func (a ListAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired returns the Required field value.
func (a ListAttribute) IsRequired() bool {
	return a.Required
}

// IsSensitive returns the Sensitive field value.
func (a ListAttribute) IsSensitive() bool {
	return a.Sensitive
}

// ListDefaultValue returns the Default field value.
func (a ListAttribute) ListDefaultValue() defaults.List {
	return a.Default
}

// ListPlanModifiers returns the PlanModifiers field value.
func (a ListAttribute) ListPlanModifiers() []planmodifier.List {
	return a.PlanModifiers
}

// ListValidators returns the Validators field value.
func (a ListAttribute) ListValidators() []validator.List {
	return a.Validators
}

// ValidateImplementation contains logic for validating the
// provider-defined implementation of the attribute to prevent unexpected
// errors or panics. This logic runs during the GetProviderSchema RPC and
// should never include false positives.
func (a ListAttribute) ValidateImplementation(ctx context.Context, req fwschema.ValidateImplementationRequest, resp *fwschema.ValidateImplementationResponse) {
	if a.CustomType == nil && a.ElementType == nil {
		resp.Diagnostics.Append(fwschema.AttributeMissingElementTypeDiag(req.Path))
	}

	if a.ElementType != nil {
		resp.Diagnostics.Append(checkAttrTypeForDynamics(req.Path, a.ElementType))
	}

	if a.ListDefaultValue() != nil {
		if !a.IsComputed() {
			resp.Diagnostics.Append(nonComputedAttributeWithDefaultDiag(req.Path))
		}

		// Validate Default implementation. This is safe unless the framework
		// ever allows more dynamic Default implementations at which the
		// implementation would be required to be validated at runtime.
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		defaultReq := defaults.ListRequest{
			Path: req.Path,
		}
		defaultResp := &defaults.ListResponse{}

		a.ListDefaultValue().DefaultList(ctx, defaultReq, defaultResp)

		resp.Diagnostics.Append(defaultResp.Diagnostics...)

		if defaultResp.Diagnostics.HasError() {
			return
		}

		if a.ElementType != nil && !a.ElementType.Equal(defaultResp.PlanValue.ElementType(ctx)) {
			resp.Diagnostics.Append(fwschema.AttributeDefaultElementTypeMismatchDiag(req.Path, a.ElementType, defaultResp.PlanValue.ElementType(ctx)))
		}
	}
}

// TODO: Not sure if there is a better package for this function, but it definitely needs to go somewhere else. `attr` package?
//
// checkAttrTypeForDynamics is a helper that will return a diagnostic if an attr.Type contains any children with a dynamic attr.Type
func checkAttrTypeForDynamics(attrPath path.Path, typ attr.Type) diag.Diagnostic {
	switch attrType := typ.(type) {
	case attr.TypeWithDynamicValue:
		return fwschema.CollectionWithDynamicTypeDiag(attrPath)
	// Lists, maps, sets
	case attr.TypeWithElementType:
		return checkAttrTypeForDynamics(attrPath, attrType.ElementType())
	// Tuples
	case attr.TypeWithElementTypes:
		for _, elemType := range attrType.ElementTypes() {
			diag := checkAttrTypeForDynamics(attrPath, elemType)
			if diag != nil {
				return diag
			}
		}
		return nil
	// Objects
	case attr.TypeWithAttributeTypes:
		for _, objAttrType := range attrType.AttributeTypes() {
			diag := checkAttrTypeForDynamics(attrPath, objAttrType)
			if diag != nil {
				return diag
			}
		}
		return nil
	// Missing or unsupported type
	default:
		return nil
	}
}
