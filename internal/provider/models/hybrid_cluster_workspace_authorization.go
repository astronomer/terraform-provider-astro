package models

import (
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HybridClusterWorkspaceAuthorizationResource struct {
	ClusterId    types.String `tfsdk:"cluster_id"`
	WorkspaceIds types.Set    `tfsdk:"workspace_ids"`
}

func (data *HybridClusterWorkspaceAuthorizationResource) ReadFromResponse(
	cluster *platform_v1.Cluster,
) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ClusterId = types.StringValue(cluster.Id)
	if cluster.WorkspaceIds == nil || len(*cluster.WorkspaceIds) == 0 {
		data.WorkspaceIds = types.SetNull(types.StringType)
	} else {
		data.WorkspaceIds, diags = utils.StringSet(cluster.WorkspaceIds)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}
