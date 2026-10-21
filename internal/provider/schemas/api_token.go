package schemas

import (
	"context"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ApiTokenDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token name",
			Computed:            true,
		},
		"description": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token description",
			Computed:            true,
		},
		"short_token": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token short token",
			Computed:            true,
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token type",
			Computed:            true,
		},
		"kind": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token kind - 'STANDARD' or 'DIRECT_ACCESS'",
			Computed:            true,
		},
		"start_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "time when the API token will become valid in UTC",
			Computed:            true,
		},
		"end_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "time when the API token will expire in UTC",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "API Token creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "API Token updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"expiry_period_in_days": datasourceSchema.Int64Attribute{
			MarkdownDescription: "API Token expiry period in days",
			Computed:            true,
		},
		"last_used_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token last used timestamp",
			Computed:            true,
		},
		"roles": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DataSourceApiTokenRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the API Token",
		},
	}
}

func ApiTokenResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token name",
			Required:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token description",
			Optional:            true,
		},
		"short_token": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token short token",
			Computed:            true,
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token type - if changing this value, the API Token will be recreated with the new type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(platform_v1.ApiTokenScopeORGANIZATION),
					string(platform_v1.ApiTokenScopeWORKSPACE),
					string(platform_v1.ApiTokenRoleEntityTypeDEPLOYMENT),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"kind": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token kind - 'STANDARD' (default) or 'DIRECT_ACCESS'. 'DIRECT_ACCESS' tokens embed their Airflow permissions in the token itself, so they can call the Airflow API directly without going through the Astro IP access list. The kind cannot be changed after creation - if changing this value, the API Token will be recreated with the new kind",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform_v1.ApiTokenKindSTANDARD),
					string(platform_v1.ApiTokenKindDIRECTACCESS),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"start_at": resourceSchema.StringAttribute{
			MarkdownDescription: "time when the API token will become valid in UTC",
			Computed:            true,
		},
		"end_at": resourceSchema.StringAttribute{
			MarkdownDescription: "time when the API token will expire in UTC",
			Computed:            true,
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token creation timestamp",
			Computed:            true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "API Token creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "API Token updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"expiry_period_in_days": resourceSchema.Int64Attribute{
			MarkdownDescription: "API Token expiry period in days",
			Optional:            true,
		},
		"last_used_at": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token last used timestamp",
			Computed:            true,
		},
		"roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceApiTokenRoleSchemaAttributes(),
			},
			Required:            true,
			MarkdownDescription: "The roles assigned to the API Token. A 'DIRECT_ACCESS' API Token must have exactly one role, and changing it recreates the token since its permissions are baked into the token value",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplaceIf(
					requiresReplaceIfDirectAccessToken,
					directAccessRolesReplaceDescription,
					directAccessRolesReplaceDescription,
				),
			},
		},
		"token": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token value. Warning: This value will be saved in plaintext in the terraform state file.",
			Computed:            true,
			Sensitive:           true,
		},
	}
}

const directAccessRolesReplaceDescription = "The roles of a 'DIRECT_ACCESS' API Token cannot be updated in place because its Airflow permissions are embedded in the token value. Therefore, the API Token will be recreated with the new roles"

// requiresReplaceIfDirectAccessToken forces a replace when the roles of an existing 'DIRECT_ACCESS'
// API Token change, since the API rejects role updates for that kind of token.
func requiresReplaceIfDirectAccessToken(
	ctx context.Context,
	req planmodifier.SetRequest,
	resp *setplanmodifier.RequiresReplaceIfFuncResponse,
) {
	var kind types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("kind"), &kind)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if kind.ValueString() == string(platform_v1.ApiTokenKindDIRECTACCESS) {
		resp.RequiresReplace = true
	}
}
