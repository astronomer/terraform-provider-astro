package resources

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

var _ resource.Resource = &ClusterWorkspaceAuthorizationResource{}
var _ resource.ResourceWithConfigure = &ClusterWorkspaceAuthorizationResource{}

func NewClusterWorkspaceAuthorizationResource() resource.Resource {
	return &ClusterWorkspaceAuthorizationResource{}
}

// ClusterWorkspaceAuthorizationResource represents a cluster workspace authorization resource.
type ClusterWorkspaceAuthorizationResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *ClusterWorkspaceAuthorizationResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_cluster_workspace_authorization"
}

func (r *ClusterWorkspaceAuthorizationResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster workspace authorization resource.",
		Attributes:          schemas.ResourceClusterWorkspaceAuthorizationSchemaAttributes(),
	}
}

func (r *ClusterWorkspaceAuthorizationResource) Configure(
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

func (r *ClusterWorkspaceAuthorizationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.ClusterWorkspaceAuthorizationResource

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	var updateClusterRequest platform.UpdateClusterRequest
	updateDedicatedClusterRequest := platform.UpdateDedicatedClusterRequest{}

	// workspaceIds
	if !data.WorkspaceIds.IsNull() {
		workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
		updateDedicatedClusterRequest.WorkspaceIds = &workspaceIds
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	err := updateClusterRequest.FromUpdateDedicatedClusterRequest(updateDedicatedClusterRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to update cluster error: %v", err))
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster request body, got error: %s", err),
		)
		return
	}

	cluster, err := r.platformClient.UpdateClusterWithResponse(ctx, r.organizationId, data.ClusterId.ValueString(), updateClusterRequest)
	if err != nil {
		tflog.Error(ctx, "failed to update cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster, got error: %s", err),
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
		Refresh:    r.resourceRefreshFunc(ctx, cluster.JSON200.Id),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster update failed", err.Error())
		return
	}

	diags = data.ReadFromResponse(readyCluster.(*platform.Cluster))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Created cluster workspace authorization for: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterWorkspaceAuthorizationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.ClusterWorkspaceAuthorizationResource

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

func (r *ClusterWorkspaceAuthorizationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.ClusterWorkspaceAuthorizationResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	var updateClusterRequest platform.UpdateClusterRequest
	updateDedicatedClusterRequest := platform.UpdateDedicatedClusterRequest{}

	// workspaceIds
	if !data.WorkspaceIds.IsNull() {
		workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
		updateDedicatedClusterRequest.WorkspaceIds = &workspaceIds
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	err := updateClusterRequest.FromUpdateDedicatedClusterRequest(updateDedicatedClusterRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to update cluster error: %v", err))
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster request body, got error: %s", err),
		)
		return
	}

	cluster, err := r.platformClient.UpdateClusterWithResponse(ctx, r.organizationId, data.ClusterId.ValueString(), updateClusterRequest)
	if err != nil {
		tflog.Error(ctx, "failed to update cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster, got error: %s", err),
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
		Refresh:    r.resourceRefreshFunc(ctx, cluster.JSON200.Id),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster update failed", err.Error())
		return
	}

	diags = data.ReadFromResponse(readyCluster.(*platform.Cluster))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Updated cluster workspace authorization for: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterWorkspaceAuthorizationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.ClusterWorkspaceAuthorizationResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	var updateClusterRequest platform.UpdateClusterRequest
	updateDedicatedClusterRequest := platform.UpdateDedicatedClusterRequest{
		WorkspaceIds: nil,
	}

	err := updateClusterRequest.FromUpdateDedicatedClusterRequest(updateDedicatedClusterRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to update cluster error: %v", err))
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster request body, got error: %s", err),
		)
		return
	}

	cluster, err := r.platformClient.UpdateClusterWithResponse(ctx, r.organizationId, data.ClusterId.ValueString(), updateClusterRequest)
	if err != nil {
		tflog.Error(ctx, "failed to update cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster, got error: %s", err),
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
		Refresh:    r.resourceRefreshFunc(ctx, cluster.JSON200.Id),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster update failed", err.Error())
		return
	}

	diags = data.ReadFromResponse(readyCluster.(*platform.Cluster))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted cluster workspace authorization for: %v", data.ClusterId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// resourceRefreshFunc returns a retry.StateRefreshFunc that polls the platform API for the cluster status
// If the cluster is not found, it returns "DELETED" status
// If the cluster is found, it returns the cluster status
// If there is an error, it returns the error
// WaitForStateContext will keep polling until the target status is reached, the timeout is reached or an err is returned
func (r *ClusterWorkspaceAuthorizationResource) resourceRefreshFunc(ctx context.Context, clusterId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		cluster, err := r.platformClient.GetClusterWithResponse(ctx, r.organizationId, clusterId)
		if err != nil {
			tflog.Error(ctx, "failed to get cluster while polling for cluster 'CREATED' status", map[string]interface{}{"error": err})
			return nil, "", err
		}
		statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
		if statusCode == http.StatusNotFound {
			return &platform.Cluster{}, "DELETED", nil
		}
		if diagnostic != nil {
			return nil, "", fmt.Errorf("error getting cluster %s", diagnostic.Detail())
		}
		if cluster != nil && cluster.JSON200 != nil {
			switch cluster.JSON200.Status {
			case platform.ClusterStatusCREATED:
				return cluster.JSON200, string(cluster.JSON200.Status), nil
			case platform.ClusterStatusUPDATEFAILED, platform.ClusterStatusCREATEFAILED:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("cluster mutation failed for cluster '%v'", cluster.JSON200.Id)
			case platform.ClusterStatusCREATING, platform.ClusterStatusUPDATING:
				return cluster.JSON200, string(cluster.JSON200.Status), nil
			default:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("unexpected cluster status '%v' for cluster '%v'", cluster.JSON200.Status, cluster.JSON200.Id)
			}
		}
		return nil, "", fmt.Errorf("error getting cluster %s", clusterId)
	}
}
