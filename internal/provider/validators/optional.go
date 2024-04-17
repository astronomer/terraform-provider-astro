package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = optionalValidator{}

type optionalValidator struct {
	Validator validator.String
}

func (v optionalValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v optionalValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("value must be empty or %v", v.Validator.MarkdownDescription(ctx))
}

func (v optionalValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() || request.ConfigValue.ValueString() == "" {
		return
	}

	v.Validator.ValidateString(ctx, request, response)
}

func OptionalString(v validator.String) validator.String {
	return optionalValidator{
		Validator: v,
	}
}
