package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// AstroProviderModel describes the provider data model.
type AstroProviderModel struct {
	Token          types.String `tfsdk:"token"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Host           types.String `tfsdk:"host"`
}
