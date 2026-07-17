package schemas

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AllowedIpAddressRangesResourceSchemaAttributes returns the attributes for the
// astro_allowed_ip_address_ranges resource.
func AllowedIpAddressRangesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"ip_address_ranges": resourceSchema.SetAttribute{
			MarkdownDescription: "The organization's allowed IP address ranges, in CIDR format (e.g. `203.0.113.0/24`). " +
				"This resource authoritatively manages the organization's full IP access list - ranges not included here " +
				"are removed on apply. An empty set removes all restrictions.",
			Required:    true,
			ElementType: types.StringType,
		},
	}
}
