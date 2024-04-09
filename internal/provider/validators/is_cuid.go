package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/lucsky/cuid"
)

var _ validator.String = isCuidValidator{}

type isCuidValidator struct {
}

func (v isCuidValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v isCuidValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a cuid"
}

func (v isCuidValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.String()
	if err := cuid.IsCuid(value); err == nil {
		return
	}

	response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
		request.Path,
		v.Description(ctx),
		value,
	))
}

func IsCuid() validator.String {
	return isCuidValidator{}
}

var _ validator.List = listIsCuidsValidator{}

type listIsCuidsValidator struct{}

func (v listIsCuidsValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v listIsCuidsValidator) MarkdownDescription(_ context.Context) string {
	return "each value in list must be a cuid"
}

func (v listIsCuidsValidator) ValidateList(
	ctx context.Context,
	request validator.ListRequest,
	response *validator.ListResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.Elements()
	for i, elem := range value {
		if elem.IsNull() || elem.IsUnknown() {
			response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
				request.Path.AtListIndex(i),
				v.Description(ctx),
				elem.String(),
			))
		}

		if err := cuid.IsCuid(elem.String()); err != nil {
			response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
				request.Path.AtListIndex(i),
				v.Description(ctx),
				elem.String(),
			))
		}
	}
}

func ListIsCuids() validator.List {
	return listIsCuidsValidator{}
}
