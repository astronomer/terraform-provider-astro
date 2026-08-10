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
	return "value must be a valid CIDR range (e.g. `10.0.0.0/8`)"
}

func (v isCidrValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	// If the value is unknown or null, we can't validate it - it may be resolved from a variable.
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// net.ParseCIDR accepts host bits (e.g. "10.1.2.3/8"), which the API also accepts and stores
	// verbatim, so those round-trip cleanly and must not be rejected here. It does reject surrounding
	// whitespace, which matters: the API trims whitespace before storing, so an untrimmed value would
	// come back in a different form and trip Terraform's "inconsistent result after apply" check.
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
