package schemas

import (
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ProviderSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"token": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Astro API Token. Can be set with an `ASTRO_API_TOKEN` env var.",
		},
		"organization_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Organization ID this provider will operate on.",
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"host": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "API host to use for the provider. Default is `https://api.astronomer.io`",
		},
	}
}
