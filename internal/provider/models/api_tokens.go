package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApiTokens describes the data source data model.
type ApiTokens struct {
	ApiTokens                     types.Set    `tfsdk:"api_tokens"`
	WorkspaceId                   types.String `tfsdk:"workspace_id"`                     // query parameter
	DeploymentId                  types.String `tfsdk:"deployment_id"`                    // query parameter
	IncludeOnlyOrganizationTokens types.Bool   `tfsdk:"include_only_organization_tokens"` // query parameter
}

func (data *ApiTokens) ReadFromResponse(ctx context.Context, apiTokens []iam.ApiToken) diag.Diagnostics {
	values := make([]attr.Value, len(apiTokens))
	for i, apiToken := range apiTokens {
		var singleApiTokenData ApiToken
		diags := singleApiTokenData.ReadFromResponse(ctx, &apiToken)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.ApiTokensElementAttributeTypes(), singleApiTokenData)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.ApiTokens, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.ApiTokensElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
