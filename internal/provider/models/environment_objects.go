package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObjects describes the data source data model.
type EnvironmentObjects struct {
	WorkspaceId        types.String `tfsdk:"workspace_id"`
	DeploymentId       types.String `tfsdk:"deployment_id"`
	ObjectType         types.String `tfsdk:"object_type"`
	ObjectKey          types.String `tfsdk:"object_key"`
	ShowSecrets        types.Bool   `tfsdk:"show_secrets"`
	ResolveLinked      types.Bool   `tfsdk:"resolve_linked"`
	EnvironmentObjects types.Set    `tfsdk:"environment_objects"`
}

func (data *EnvironmentObjects) ReadFromResponse(ctx context.Context, objects []platform.EnvironmentObject) diag.Diagnostics {
	values := make([]attr.Value, len(objects))
	for i, obj := range objects {
		var single EnvironmentObject
		diags := single.ReadFromResponse(ctx, &obj, nil)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.EnvironmentObjectsElementAttributeTypes(), single)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}

	var diags diag.Diagnostics
	data.EnvironmentObjects, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.EnvironmentObjectsElementAttributeTypes()}, values)
	return diags
}
