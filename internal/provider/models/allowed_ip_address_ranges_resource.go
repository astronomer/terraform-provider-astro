package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// AllowedIpAddressRangesResource describes the astro_allowed_ip_address_ranges resource data
// model. The resource authoritatively manages the organization's full IP access list as a single
// set of CIDR ranges.
type AllowedIpAddressRangesResource struct {
	Id              types.String `tfsdk:"id"`
	IpAddressRanges types.Set    `tfsdk:"ip_address_ranges"`
}
