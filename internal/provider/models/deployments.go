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
	DeploymentIds types.List `tfsdk:"deployment_ids"` // query parameter
	Names         types.List `tfsdk:"names"`          // query parameter
}

func (data *DeploymentsDataSource) ReadFromResponse(
	ctx context.Context,
	deployments []platform.Deployment,
) diag.Diagnostics {
	if len(deployments) == 0 {
		types.ListNull(types.ObjectType{AttrTypes: schemas.DeploymentsElementAttributeTypes()})
	}

	values := make([]attr.Value, len(deployments))
	for i, deployment := range deployments {
		var data DeploymentDataSource
		diags := data.ReadFromResponse(ctx, &deployment)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.DeploymentsElementAttributeTypes(), data)
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
