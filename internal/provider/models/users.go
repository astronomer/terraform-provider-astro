package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Users describes the data source data model.
type Users struct {
	Users        types.Set    `tfsdk:"users"`
	WorkspaceId  types.String `tfsdk:"workspace_id"`  // query parameter
	DeploymentId types.String `tfsdk:"deployment_id"` // query parameter
}

func (data *Users) ReadFromResponse(ctx context.Context, users []iam.User) diag.Diagnostics {
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
	data.Users, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.UsersElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
