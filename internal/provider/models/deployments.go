package models

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeploymentsDataSource describes the data source data model.
type DeploymentsDataSource struct {
	Deployments   types.List `tfsdk:"deployments"`
	WorkspaceIds  types.List `tfsdk:"workspace_ids"`  // query parameter
	DeploymentIds types.List `tfsdk:"deployment_ids"` // query parameter
	Names         types.List `tfsdk:"names"`          // query parameter
}

func (data *DeploymentsDataSource) ReadFromResponse(
	ctx context.Context,
	deployments []platform.Deployment,
) diag.Diagnostics {
	values := make([]attr.Value, len(deployments))
	for i, deployment := range deployments {
		var singleDeploymentData DeploymentDataSource
		diags := singleDeploymentData.ReadFromResponse(ctx, &deployment)
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
	data.Deployments, diags = types.ListValue(types.ObjectType{AttrTypes: schemas.DeploymentsElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
