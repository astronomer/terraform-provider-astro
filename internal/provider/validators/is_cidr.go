package validators

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = isCidrValidator{}

type isCidrValidator struct {
}

func (v isCidrValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v isCidrValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid CIDR range (e.g. `203.0.113.0/24`)"
}

func (v isCidrValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	// If the value is unknown or null we can't validate it (it may resolve from a variable).
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if _, _, err := net.ParseCIDR(value); err == nil {
		return
	}

	response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
		request.Path,
		v.Description(ctx),
		value,
	))
}

func IsCidr() validator.String {
	return isCidrValidator{}
}
