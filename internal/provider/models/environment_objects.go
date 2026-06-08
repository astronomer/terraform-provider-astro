package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentObjects struct {
	WorkspaceId        types.String `tfsdk:"workspace_id"`
	DeploymentId       types.String `tfsdk:"deployment_id"`
	ObjectType         types.String `tfsdk:"object_type"`
	ObjectKey          types.String `tfsdk:"object_key"`
	EnvironmentObjects types.List   `tfsdk:"environment_objects"`
}

func (data *EnvironmentObjects) ReadFromResponse(ctx context.Context, objects []platform.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics

	envObjAttrTypes := schemas.EnvironmentObjectDataSourceSchemaAttributes()
	attrTypes := make(map[string]attr.Type)
	for k, v := range envObjAttrTypes {
		attrTypes[k] = v.GetType()
	}

	if len(objects) == 0 {
		data.EnvironmentObjects = types.ListNull(types.ObjectType{AttrTypes: attrTypes})
		return nil
	}

	envObjValues := make([]attr.Value, len(objects))
	for i, obj := range objects {
		var envObj EnvironmentObject
		diags = envObj.ReadFromResponse(ctx, &obj)
		if diags.HasError() {
			return diags
		}

		objVal, d := types.ObjectValueFrom(ctx, attrTypes, &envObj)
		if d.HasError() {
			return d
		}
		envObjValues[i] = objVal
	}

	data.EnvironmentObjects, diags = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, envObjValues)
	return diags
}
