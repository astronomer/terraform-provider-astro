package schemas

import (
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func OrganizationDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization identifier",
			Computed:            true, // This is computed because we retrieve it from the provider configuration
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization name",
			Computed:            true,
		},
		"support_plan": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization support plan",
			Computed:            true,
		},
		"product": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization product type",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Organization creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Organization updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"trial_expires_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization trial expiration timestamp",
			Computed:            true,
		},
		"status": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization status",
			Computed:            true,
		},
		"payment_method": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization payment method",
			Computed:            true,
		},
		"is_scim_enabled": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether SCIM is enabled for the organization",
			Computed:            true,
		},
		"billing_email": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization billing email",
			Computed:            true,
		},
	}
}
