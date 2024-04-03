package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// AstronomerProviderModel describes the provider data model.
type AstronomerProviderModel struct {
	Token          types.String `tfsdk:"token"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Host           types.String `tfsdk:"host"`
}
