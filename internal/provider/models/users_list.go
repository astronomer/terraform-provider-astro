package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UsersList describes the astro_users_list data source data model. It mirrors
// Users but represents the users collection as an ordered List instead of a Set
// for significantly better plan performance on large organizations. The
// per-user element type is shared with the Set-based astro_users data source.
type UsersList struct {
	Users        types.List   `tfsdk:"users"`
	WorkspaceId  types.String `tfsdk:"workspace_id"`  // query parameter
	DeploymentId types.String `tfsdk:"deployment_id"` // query parameter
}

func (data *UsersList) ReadFromResponse(ctx context.Context, users []iam.User) diag.Diagnostics {
	values := make([]attr.Value, len(users))
	for i, user := range users {
		var singleUserData User
		diags := singleUserData.ReadFromResponse(ctx, &user)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.UsersElementAttributeTypes(), singleUserData)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.Users, diags = types.ListValue(types.ObjectType{AttrTypes: schemas.UsersElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
