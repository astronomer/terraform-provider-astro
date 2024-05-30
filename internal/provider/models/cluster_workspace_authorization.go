package models

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClusterWorkspaceAuthorizationResource struct {
	ClusterId    types.String `tfsdk:"cluster_id"`
	WorkspaceIds types.Set    `tfsdk:"workspace_ids"`
}

func (data *ClusterWorkspaceAuthorizationResource) ReadFromResponse(
	cluster *platform.Cluster,
) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ClusterId = types.StringValue(cluster.Id)

	workspaceIds := make([]attr.Value, len(*cluster.WorkspaceIds))
	for i, id := range *cluster.WorkspaceIds {
		workspaceIds[i] = types.StringValue(id)
	}
	data.WorkspaceIds, diags = types.SetValue(types.StringType, workspaceIds)
	if diags.HasError() {
		return diags
	}

	return nil
}
