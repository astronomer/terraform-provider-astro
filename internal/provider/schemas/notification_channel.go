package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NotificationChannelDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
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
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "API Token type",
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
	}
}
