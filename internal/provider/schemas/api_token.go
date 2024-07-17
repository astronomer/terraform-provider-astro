package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
				stringvalidator.OneOf(string(iam.ApiTokenTypeORGANIZATION),
					string(iam.ApiTokenTypeWORKSPACE),
					string(iam.ApiTokenRoleEntityTypeDEPLOYMENT),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
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
			MarkdownDescription: "The roles assigned to the API Token",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"token": resourceSchema.StringAttribute{
			MarkdownDescription: "API Token value",
			Computed:            true,
			Sensitive:           true,
		},
	}
}
