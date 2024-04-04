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
