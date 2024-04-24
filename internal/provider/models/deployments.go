package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Deployments describes the data source data model.
type Deployments struct {
	Deployments   types.Set `tfsdk:"deployments"`
	WorkspaceIds  types.Set `tfsdk:"workspace_ids"`  // query parameter
	DeploymentIds types.Set `tfsdk:"deployment_ids"` // query parameter
	Names         types.Set `tfsdk:"names"`          // query parameter
}

func (data *Deployments) ReadFromResponse(
	ctx context.Context,
	deployments []platform.Deployment,
) diag.Diagnostics {
	values := make([]attr.Value, len(deployments))
	for i, deployment := range deployments {
		var singleDeploymentData Deployment
		diags := singleDeploymentData.ReadFromResponse(ctx, &deployment, false)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.DeploymentsElementAttributeTypes(), singleDeploymentData)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.Deployments, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.DeploymentsElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
