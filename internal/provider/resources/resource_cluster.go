package resources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}
var _ resource.ResourceWithConfigure = &ClusterResource{}
var _ resource.ResourceWithValidateConfig = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *ClusterResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster resource. If creating multiple clusters, add a delay between each cluster creation to avoid cluster creation limiting errors.",
		Attributes:          schemas.ClusterResourceSchemaAttributes(ctx),
	}
}

func (r *ClusterResource) Configure(
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

func (r *ClusterResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.ClusterResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var createClusterRequest platform.CreateClusterRequest

	switch platform.ClusterCloudProvider(data.CloudProvider.ValueString()) {
	case platform.ClusterCloudProviderAWS:
		createAwsDedicatedClusterRequest := platform.CreateAwsClusterRequest{
			CloudProvider:   platform.CreateAwsClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			Name:            data.Name.ValueString(),
			NodePools:       nil,
			ProviderAccount: data.ProviderAccount.ValueStringPointer(),
			Region:          data.Region.ValueString(),
			Type:            platform.CreateAwsClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:  data.VpcSubnetRange.ValueString(),
		}

		// workspaceIds
		workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
		createAwsDedicatedClusterRequest.WorkspaceIds = &workspaceIds
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := createClusterRequest.FromCreateAwsClusterRequest(createAwsDedicatedClusterRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create cluster error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create cluster request body, got error: %s", err),
			)
			return
		}
	case platform.ClusterCloudProviderAZURE:
		createAzureDedicatedClusterRequest := platform.CreateAzureClusterRequest{
			CloudProvider:   platform.CreateAzureClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			Name:            data.Name.ValueString(),
			NodePools:       nil,
			ProviderAccount: data.ProviderAccount.ValueStringPointer(),
			Region:          data.Region.ValueString(),
			TenantId:        data.TenantId.ValueStringPointer(),
			Type:            platform.CreateAzureClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:  data.VpcSubnetRange.ValueString(),
		}

		// workspaceIds
		workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
		createAzureDedicatedClusterRequest.WorkspaceIds = &workspaceIds
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := createClusterRequest.FromCreateAzureClusterRequest(createAzureDedicatedClusterRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create cluster error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create cluster request body, got error: %s", err),
			)
			return
		}
	case platform.ClusterCloudProviderGCP:
		createGcpDedicatedClusterRequest := platform.CreateGcpClusterRequest{
			CloudProvider:       platform.CreateGcpClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			Name:                data.Name.ValueString(),
			NodePools:           nil,
			PodSubnetRange:      data.PodSubnetRange.ValueString(),
			ProviderAccount:     data.ProviderAccount.ValueStringPointer(),
			Region:              data.Region.ValueString(),
			ServicePeeringRange: data.ServicePeeringRange.ValueString(),
			ServiceSubnetRange:  data.ServiceSubnetRange.ValueString(),
			Type:                platform.CreateGcpClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:      data.VpcSubnetRange.ValueString(),
		}

		// workspaceIds
		workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
		createGcpDedicatedClusterRequest.WorkspaceIds = &workspaceIds
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := createClusterRequest.FromCreateGcpClusterRequest(createGcpDedicatedClusterRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create cluster error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create cluster request body, got error: %s", err),
			)
			return
		}
	}

	// Create the timeout context for the cluster creation
	createTimeout, diags := data.Timeouts.Create(ctx, 3*time.Hour)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cluster, err := r.platformClient.CreateClusterWithResponse(
		ctx,
		r.organizationId,
		createClusterRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create cluster, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Wait for the cluster to be created (or fail)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusUPGRADEPENDING)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED), string(platform.ClusterStatusACCESSDENIED)},
		Refresh:    ClusterResourceRefreshFunc(ctx, r.platformClient, r.organizationId, cluster.JSON200.Id),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster creation failed", err.Error())
		return
	}

	diags = data.ReadFromResponse(ctx, readyCluster.(*platform.Cluster))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a cluster resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.ClusterResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	cluster, err := r.platformClient.GetClusterWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get cluster, got error: %s", err),
		)
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

	diags := data.ReadFromResponse(ctx, cluster.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a cluster resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.ClusterResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update request
	var diags diag.Diagnostics
	var updateClusterRequest platform.UpdateClusterRequest

	updateDedicatedClusterRequest := platform.UpdateDedicatedClusterRequest{
		ClusterType:  (*platform.UpdateDedicatedClusterRequestClusterType)(data.Type.ValueStringPointer()),
		K8sTags:      []platform.ClusterK8sTag{},
		Name:         data.Name.ValueString(),
		NodePools:    nil,
		WorkspaceIds: nil,
	}

	// workspaceIds
	workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
	updateDedicatedClusterRequest.WorkspaceIds = &workspaceIds
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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

	// Create the timeout context for the cluster update
	updateTimeout, diags := data.Timeouts.Update(ctx, 3*time.Hour)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Retry update cluster request if there is a 409 conflict (workflow already running)
	var cluster *platform.UpdateClusterResponse
	err = retry.RetryContext(ctx, updateTimeout, func() *retry.RetryError {
		var apiErr error
		cluster, apiErr = r.platformClient.UpdateClusterWithResponse(
			ctx,
			r.organizationId,
			data.Id.ValueString(),
			updateClusterRequest,
		)
		if apiErr != nil {
			tflog.Error(ctx, "failed to update cluster", map[string]interface{}{"error": apiErr})
			return retry.NonRetryableError(fmt.Errorf("unable to update cluster, got error: %s", apiErr))
		}
		statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
		if statusCode == http.StatusConflict {
			// Workflow is already running, retry after a delay
			tflog.Info(ctx, "cluster workflow in progress, retrying update", map[string]interface{}{"clusterId": data.Id.ValueString()})
			return retry.RetryableError(fmt.Errorf("workflow is already running for cluster, retrying"))
		}
		if diagnostic != nil {
			return retry.NonRetryableError(fmt.Errorf("%s", diagnostic.Detail()))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update cluster after retries, got error: %s", err),
		)
		return
	}

	// Wait for the cluster to be updated (or fail)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusUPGRADEPENDING)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED), string(platform.ClusterStatusACCESSDENIED)},
		Refresh:    ClusterResourceRefreshFunc(ctx, r.platformClient, r.organizationId, cluster.JSON200.Id),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster update failed", err.Error())
		return
	}

	diags = data.ReadFromResponse(ctx, readyCluster.(*platform.Cluster))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a cluster resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.ClusterResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the timeout context for the cluster delete
	deleteTimeout, diags := data.Timeouts.Delete(ctx, 1*time.Hour)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Retry delete cluster request if there is a 409 conflict (workflow already running)
	var cluster *platform.DeleteClusterResponse
	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		var apiErr error
		cluster, apiErr = r.platformClient.DeleteClusterWithResponse(
			ctx,
			r.organizationId,
			data.Id.ValueString(),
		)
		if apiErr != nil {
			tflog.Error(ctx, "failed to delete cluster", map[string]interface{}{"error": apiErr})
			return retry.NonRetryableError(fmt.Errorf("unable to delete cluster, got error: %s", apiErr))
		}
		statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
		// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
		if statusCode == http.StatusNotFound {
			return nil
		}
		if statusCode == http.StatusConflict {
			// Workflow is already running, retry after a delay
			tflog.Info(ctx, "cluster workflow in progress, retrying delete", map[string]interface{}{"clusterId": data.Id.ValueString()})
			return retry.RetryableError(fmt.Errorf("workflow is already running for cluster, retrying"))
		}
		if diagnostic != nil {
			return retry.NonRetryableError(fmt.Errorf("%s", diagnostic.Detail()))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete cluster after retries, got error: %s", err),
		)
		return
	}

	// Wait for the cluster to be deleted
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED), string(platform.ClusterStatusUPGRADEPENDING)},
		Target:     []string{"DELETED"},
		Refresh:    ClusterResourceRefreshFunc(ctx, r.platformClient, r.organizationId, data.Id.ValueString()),
		Timeout:    1 * time.Hour,
		MinTimeout: 30 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster deletion failed", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a cluster resource: %v", data.Id.ValueString()))
}

func (r *ClusterResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig validates the configuration of the resource as a whole before any operations are performed.
// This is a good place to check for any conflicting settings.
func (r *ClusterResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.ClusterResource

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Cloud provider specific validation
	switch platform.ClusterCloudProvider(data.Type.ValueString()) {
	case platform.ClusterCloudProviderAWS:
		resp.Diagnostics.Append(validateAwsConfig(ctx, &data)...)
	case platform.ClusterCloudProviderAZURE:
		resp.Diagnostics.Append(validateAzureConfig(ctx, &data)...)
	case platform.ClusterCloudProviderGCP:
		resp.Diagnostics.Append(validateGcpConfig(ctx, &data)...)
	}
}

func validateAwsConfig(ctx context.Context, data *models.ClusterResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)

	// Unallowed values
	if !data.TenantId.IsNull() {
		diags.AddError(
			"tenant_id is not allowed for 'AWS' cluster",
			"Please remove tenant_id",
		)
	}
	if !data.ServicePeeringRange.IsNull() {
		diags.AddError(
			"service_peering_range is not allowed for 'AWS' cluster",
			"Please remove service_peering_range",
		)
	}
	if !data.PodSubnetRange.IsNull() {
		diags.AddError(
			"pod_subnet_range is not allowed for 'AWS' cluster",
			"Please remove pod_subnet_range",
		)
	}
	if !data.ServiceSubnetRange.IsNull() {
		diags.AddError(
			"service_subnet_range is not allowed for 'AWS' cluster",
			"Please remove service_subnet_range",
		)
	}
	return diags
}

func validateAzureConfig(ctx context.Context, data *models.ClusterResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)

	// Unallowed values
	if !data.ServicePeeringRange.IsNull() {
		diags.AddError(
			"service_peering_range is not allowed for 'AZURE' cluster",
			"Please remove service_peering_range",
		)
	}
	if !data.PodSubnetRange.IsNull() {
		diags.AddError(
			"pod_subnet_range is not allowed for 'AZURE' cluster",
			"Please remove pod_subnet_range",
		)
	}
	if !data.ServiceSubnetRange.IsNull() {
		diags.AddError(
			"service_subnet_range is not allowed for 'AZURE' cluster",
			"Please remove service_subnet_range",
		)
	}
	return diags
}

func validateGcpConfig(ctx context.Context, data *models.ClusterResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)

	// required values
	if data.ServicePeeringRange.IsNull() {
		diags.AddError(
			"service_peering_range is required for 'GCP' cluster",
			"Please add service_peering_range",
		)
	}
	if data.PodSubnetRange.IsNull() {
		diags.AddError(
			"pod_subnet_range is required for 'GCP' cluster",
			"Please add pod_subnet_range",
		)
	}
	if data.ServiceSubnetRange.IsNull() {
		diags.AddError(
			"service_subnet_range is required for 'GCP' cluster",
			"Please add service_subnet_range",
		)
	}

	// Unallowed values
	if !data.TenantId.IsNull() {
		diags.AddError(
			"tenant_id is not allowed for 'AWS' cluster",
			"Please remove tenant_id",
		)
	}
	return diags
}
