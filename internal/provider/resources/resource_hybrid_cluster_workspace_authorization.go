package resources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var _ resource.Resource = &hybridClusterWorkspaceAuthorizationResource{}
var _ resource.ResourceWithImportState = &hybridClusterWorkspaceAuthorizationResource{}
var _ resource.ResourceWithConfigure = &hybridClusterWorkspaceAuthorizationResource{}

func NewHybridClusterWorkspaceAuthorizationResource() resource.Resource {
	return &hybridClusterWorkspaceAuthorizationResource{}
}

// hybridClusterWorkspaceAuthorizationResource represents a hybrid cluster workspace authorization resource.
type hybridClusterWorkspaceAuthorizationResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *hybridClusterWorkspaceAuthorizationResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_hybrid_cluster_workspace_authorization"
}

func (r *hybridClusterWorkspaceAuthorizationResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Hybrid cluster workspace authorization resource",
		Attributes:          schemas.ResourceHybridClusterWorkspaceAuthorizationSchemaAttributes(),
	}
}

func (r *hybridClusterWorkspaceAuthorizationResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.ResourceApiClientConfigureError(ctx, req, resp)
		return
	}

	r.platformClient = apiClients.PlatformClient
	r.organizationId = apiClients.OrganizationId
}

func (r *hybridClusterWorkspaceAuthorizationResource) MutateRoles(
	ctx context.Context,
	data *models.HybridClusterWorkspaceAuthorizationResource,
) diag.Diagnostics {
	diags := diag.Diagnostics{}
	var updateClusterRequest platform.UpdateClusterRequest
	updateHybridClusterRequest := platform.UpdateHybridClusterRequest{
		ClusterType: platform.UpdateHybridClusterRequestClusterTypeHYBRID,
	}

	// workspaceIds
	if !data.WorkspaceIds.IsNull() {
		workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
		updateHybridClusterRequest.WorkspaceIds = &workspaceIds
		if diags.HasError() {
			return diags
		}
	}

	err := updateClusterRequest.FromUpdateHybridClusterRequest(updateHybridClusterRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to mutate hybrid cluster workspace authorization error: %v", err))
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Failed to mutate hybrid cluster workspace authorization, got error: %s", err),
		)
		return diags
	}

	cluster, err := r.platformClient.UpdateClusterWithResponse(ctx, r.organizationId, data.ClusterId.ValueString(), updateClusterRequest)
	if err != nil {
		tflog.Error(ctx, "failed to mutate hybrid cluster workspace authorization", map[string]interface{}{"error": err})
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to mutate hybrid cluster workspace authorization, got error: %s", err),
		)
		return diags
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
	if diagnostic != nil {
		diags.Append(diagnostic)
		return diags
	}

	// Wait for the cluster to be updated (or fail)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED)},
		Refresh:    ClusterResourceRefreshFunc(ctx, r.platformClient, r.organizationId, cluster.JSON200.Id),
		Timeout:    1 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		diags.AddError("Hybrid cluster authorization mutation", err.Error())
		return diags
	}

	diags = data.ReadFromResponse(readyCluster.(*platform.Cluster))
	if diags.HasError() {
		return diags
	}

	return nil
}

func (r *hybridClusterWorkspaceAuthorizationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.HybridClusterWorkspaceAuthorizationResource

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.MutateRoles(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Created hybrid cluster workspace authorization for cluster: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridClusterWorkspaceAuthorizationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.HybridClusterWorkspaceAuthorizationResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.platformClient.GetClusterWithResponse(ctx, r.organizationId, data.ClusterId.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to get cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get cluster, got error: %s", err))
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
	// If the resource no longer exists, it is recommended to ignore the errors
	// and call RemoveResource to remove the resource from the state. The next Terraform plan will recreate the resource.
	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(cluster.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Read cluster workspace authorization for: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridClusterWorkspaceAuthorizationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.HybridClusterWorkspaceAuthorizationResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.MutateRoles(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Updated hybrid cluster workspace authorization for cluster: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridClusterWorkspaceAuthorizationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.HybridClusterWorkspaceAuthorizationResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	var updateClusterRequest platform.UpdateClusterRequest
	updateHybridClusterRequest := platform.UpdateHybridClusterRequest{
		ClusterType:  platform.UpdateHybridClusterRequestClusterTypeHYBRID,
		WorkspaceIds: nil,
	}

	err := updateClusterRequest.FromUpdateHybridClusterRequest(updateHybridClusterRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to delete hybrid cluster workspace authorization error: %v", err))
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete hybrid cluster workspace authorization, got error: %s", err),
		)
		return
	}

	cluster, err := r.platformClient.UpdateClusterWithResponse(ctx, r.organizationId, data.ClusterId.ValueString(), updateClusterRequest)
	if err != nil {
		tflog.Error(ctx, "failed to delete hybrid cluster workspace authorization", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete hybrid cluster workspace authorization, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Wait for the cluster to be updated (or fail)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED)},
		Refresh:    ClusterResourceRefreshFunc(ctx, r.platformClient, r.organizationId, cluster.JSON200.Id),
		Timeout:    1 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Hybrid cluster workspace authorization delete failed", err.Error())
		return
	}

	diags = data.ReadFromResponse(readyCluster.(*platform.Cluster))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted hybrid cluster workspace authorization for cluster: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridClusterWorkspaceAuthorizationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("cluster_id"), req, resp)
}
