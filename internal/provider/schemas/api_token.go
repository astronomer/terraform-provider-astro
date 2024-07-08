package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
			MarkdownDescription: "API Token start timestamp",
			Computed:            true,
		},
		"end_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token end timestamp",
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
			MarkdownDescription: "Team creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Team updater",
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
		"token": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token",
			Computed:            true,
		},
	}
}
