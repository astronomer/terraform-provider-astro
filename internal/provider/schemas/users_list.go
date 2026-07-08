package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// UsersListDataSourceSchemaAttributes mirrors UsersDataSourceSchemaAttributes
// but exposes the users collection as a List instead of a Set. A List avoids
// the expensive nested-object hashing the framework performs for Sets, which
// dominates plan time for large organizations. The per-user element schema is
// shared with the Set-based astro_users data source.
func UsersListDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"users": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: UserDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"workspace_id": schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{validators.IsCuid()},
		},
		"deployment_id": schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{validators.IsCuid()},
		},
	}
}
