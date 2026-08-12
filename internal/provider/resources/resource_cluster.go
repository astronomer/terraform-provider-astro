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
			CloudProvider:                platform.CreateAwsClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			Name:                         data.Name.ValueString(),
			NodePools:                    nil,
			ProviderAccount:              data.ProviderAccount.ValueStringPointer(),
			Region:                       data.Region.ValueString(),
			Type:                         platform.CreateAwsClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:               data.VpcSubnetRange.ValueString(),
			SecondaryVpcCidr:             data.SecondaryVpcCidr.ValueStringPointer(),
			DrRegion:                     data.DrRegion.ValueStringPointer(),
			DrVpcSubnetRange:             data.DrVpcSubnetRange.ValueStringPointer(),
			DrSecondaryVpcCidr:           data.DrSecondaryVpcCidr.ValueStringPointer(),
			EnableReplicationTimeControl: data.EnableReplicationTimeControl.ValueBoolPointer(),
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
			CloudProvider:                platform.CreateAzureClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			Name:                         data.Name.ValueString(),
			NodePools:                    nil,
			ProviderAccount:              data.ProviderAccount.ValueStringPointer(),
			Region:                       data.Region.ValueString(),
			TenantId:                     data.TenantId.ValueStringPointer(),
			Type:                         platform.CreateAzureClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:               data.VpcSubnetRange.ValueString(),
			DrRegion:                     data.DrRegion.ValueStringPointer(),
			DrVpcSubnetRange:             data.DrVpcSubnetRange.ValueStringPointer(),
			EnableReplicationTimeControl: azureEffectiveEnableReplicationTimeControl(&data),
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
			CloudProvider:         platform.CreateGcpClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			Name:                  data.Name.ValueString(),
			NodePools:             nil,
			PodSubnetRange:        data.PodSubnetRange.ValueString(),
			ProviderAccount:       data.ProviderAccount.ValueStringPointer(),
			Region:                data.Region.ValueString(),
			ServicePeeringRange:   data.ServicePeeringRange.ValueString(),
			ServiceSubnetRange:    data.ServiceSubnetRange.ValueString(),
			Type:                  platform.CreateGcpClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:        data.VpcSubnetRange.ValueString(),
			DrRegion:              data.DrRegion.ValueStringPointer(),
			DrVpcSubnetRange:      data.DrVpcSubnetRange.ValueStringPointer(),
			DrPodSubnetRange:      data.DrPodSubnetRange.ValueStringPointer(),
			DrServicePeeringRange: data.DrServicePeeringRange.ValueStringPointer(),
			DrServiceSubnetRange:  data.DrServiceSubnetRange.ValueStringPointer(),
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
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, cluster.HTTPResponse, cluster.Body, cluster.JSON200, "create cluster")
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Wait for the cluster to be created (or fail)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusUPGRADEPENDING), string(platform.ClusterStatusFAILINGOVER)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED), string(platform.ClusterStatusACCESSDENIED), string(platform.ClusterStatusFAILOVERFAILED)},
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
	statusCode, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, cluster.HTTPResponse, cluster.Body, cluster.JSON200, "read cluster")
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
	// Set EnableDr to false if the user explicitly set is_dr_enabled to false
	if !data.IsDrEnabled.IsNull() && !data.IsDrEnabled.IsUnknown() && !data.IsDrEnabled.ValueBool() {
		enableDr := false
		updateDedicatedClusterRequest.EnableDr = &enableDr
	}
	// Unlike AWS and GCP, Azure has no replication-lag phase, so DR can be enabled
	// on an existing cluster in a single update request.
	if platform.ClusterCloudProvider(data.CloudProvider.ValueString()) == platform.ClusterCloudProviderAZURE &&
		!data.IsDrEnabled.IsNull() && !data.IsDrEnabled.IsUnknown() && data.IsDrEnabled.ValueBool() {
		enableDr := true
		updateDedicatedClusterRequest.EnableDr = &enableDr
		updateDedicatedClusterRequest.DrRegion = data.DrRegion.ValueStringPointer()
		updateDedicatedClusterRequest.DrVpcSubnetRange = data.DrVpcSubnetRange.ValueStringPointer()
		updateDedicatedClusterRequest.EnableReplicationTimeControl = azureEffectiveEnableReplicationTimeControl(&data)
	}
	// Set IsFailedOver if specified (must be an explicit value, not null or unknown)
	if !data.IsFailedOver.IsNull() && !data.IsFailedOver.IsUnknown() {
		updateDedicatedClusterRequest.IsFailedOver = data.IsFailedOver.ValueBoolPointer()
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
		statusCode, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, cluster.HTTPResponse, cluster.Body, cluster.JSON200, "update cluster")
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
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusUPGRADEPENDING), string(platform.ClusterStatusFAILINGOVER)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED), string(platform.ClusterStatusACCESSDENIED), string(platform.ClusterStatusFAILOVERFAILED)},
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
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED), string(platform.ClusterStatusUPGRADEPENDING), string(platform.ClusterStatusFAILINGOVER), string(platform.ClusterStatusFAILOVERFAILED)},
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
	switch platform.ClusterCloudProvider(data.CloudProvider.ValueString()) {
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
	if !data.DrPodSubnetRange.IsNull() {
		diags.AddError(
			"dr_pod_subnet_range is not allowed for 'AWS' cluster",
			"Please remove dr_pod_subnet_range",
		)
	}
	if !data.DrServicePeeringRange.IsNull() {
		diags.AddError(
			"dr_service_peering_range is not allowed for 'AWS' cluster",
			"Please remove dr_service_peering_range",
		)
	}
	if !data.DrServiceSubnetRange.IsNull() {
		diags.AddError(
			"dr_service_subnet_range is not allowed for 'AWS' cluster",
			"Please remove dr_service_subnet_range",
		)
	}

	// DR validation
	if !data.IsDrEnabled.IsNull() && !data.IsDrEnabled.IsUnknown() && data.IsDrEnabled.ValueBool() {
		if data.DrRegion.IsNull() {
			diags.AddError(
				"dr_region is required when is_dr_enabled is true",
				"Please set dr_region to the secondary region for Disaster Recovery",
			)
		}
	}
	if !data.IsDrEnabled.IsNull() && !data.IsDrEnabled.ValueBool() {
		if !data.DrVpcSubnetRange.IsNull() {
			diags.AddError(
				"dr_vpc_subnet_range is only valid when is_dr_enabled is true",
				"Please remove dr_vpc_subnet_range or set is_dr_enabled to true",
			)
		}
		if !data.DrSecondaryVpcCidr.IsNull() {
			diags.AddError(
				"dr_secondary_vpc_cidr is only valid when is_dr_enabled is true",
				"Please remove dr_secondary_vpc_cidr or set is_dr_enabled to true",
			)
		}
		if !data.EnableReplicationTimeControl.IsNull() {
			diags.AddError(
				"enable_replication_time_control is only valid when is_dr_enabled is true",
				"Please remove enable_replication_time_control or set is_dr_enabled to true",
			)
		}
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

	// secondary_vpc_cidr / dr_secondary_vpc_cidr are AWS only
	if !data.SecondaryVpcCidr.IsNull() {
		diags.AddError(
			"secondary_vpc_cidr is not allowed for 'AZURE' cluster",
			"Please remove secondary_vpc_cidr",
		)
	}
	if !data.DrSecondaryVpcCidr.IsNull() {
		diags.AddError(
			"dr_secondary_vpc_cidr is not allowed for 'AZURE' cluster",
			"Please remove dr_secondary_vpc_cidr",
		)
	}

	// dr_pod_subnet_range, dr_service_peering_range, dr_service_subnet_range are GCP only
	if !data.DrPodSubnetRange.IsNull() {
		diags.AddError(
			"dr_pod_subnet_range is not allowed for 'AZURE' cluster",
			"Please remove dr_pod_subnet_range",
		)
	}
	if !data.DrServicePeeringRange.IsNull() {
		diags.AddError(
			"dr_service_peering_range is not allowed for 'AZURE' cluster",
			"Please remove dr_service_peering_range",
		)
	}
	if !data.DrServiceSubnetRange.IsNull() {
		diags.AddError(
			"dr_service_subnet_range is not allowed for 'AZURE' cluster",
			"Please remove dr_service_subnet_range",
		)
	}

	// DR validation
	if !data.IsDrEnabled.IsNull() && !data.IsDrEnabled.IsUnknown() && data.IsDrEnabled.ValueBool() {
		if data.DrRegion.IsNull() {
			diags.AddError(
				"dr_region is required when is_dr_enabled is true",
				"Please set dr_region to the secondary region for Disaster Recovery",
			)
		}
		diags.Append(validateAzureReplicationTimeControlRegionPair(data)...)
	}
	if !data.IsDrEnabled.IsNull() && !data.IsDrEnabled.ValueBool() {
		if !data.DrVpcSubnetRange.IsNull() {
			diags.AddError(
				"dr_vpc_subnet_range is only valid when is_dr_enabled is true",
				"Please remove dr_vpc_subnet_range or set is_dr_enabled to true",
			)
		}
		if !data.EnableReplicationTimeControl.IsNull() {
			diags.AddError(
				"enable_replication_time_control is only valid when is_dr_enabled is true",
				"Please remove enable_replication_time_control or set is_dr_enabled to true",
			)
		}
	}

	return diags
}

// validateAzureReplicationTimeControlRegionPair mirrors apps/core's
// ValidateEnableReplicationTimeControl: Replication Time Control requires the
// cluster's primary region and dr_region to be on the same continent. This runs
// client-side (using the same region/continent data as the backend's provider
// spec) so an invalid region pair is rejected at plan time instead of only
// surfacing as an apply-time API error.
//
// The user may always explicitly set this to `false`, regardless of the region
// pair. A left-unset value is not validated here at all: the provider computes
// a safe value itself (see azureEffectiveEnableReplicationTimeControl) instead
// of requiring the user to pick a compatible region pair up front.
func validateAzureReplicationTimeControlRegionPair(data *models.ClusterResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)

	if data.EnableReplicationTimeControl.IsUnknown() || data.EnableReplicationTimeControl.IsNull() {
		return diags
	}
	if !data.EnableReplicationTimeControl.ValueBool() {
		// Explicitly opted out of Replication Time Control - always allowed.
		return diags
	}
	if data.Region.IsUnknown() || data.DrRegion.IsNull() || data.DrRegion.IsUnknown() {
		// Region values aren't known yet (e.g. derived from another resource); defer to apply-time validation.
		return diags
	}

	region := data.Region.ValueString()
	drRegion := data.DrRegion.ValueString()

	regionContinent, regionOk := azureRegionContinent(region)
	if !regionOk {
		diags.AddError(
			fmt.Sprintf("Region %s has no defined location", region),
			"Please use a supported Azure region, or set enable_replication_time_control to false.",
		)
		return diags
	}
	drRegionContinent, drRegionOk := azureRegionContinent(drRegion)
	if !drRegionOk {
		diags.AddError(
			fmt.Sprintf("Region %s has no defined location", drRegion),
			"Please use a supported Azure region, or set enable_replication_time_control to false.",
		)
		return diags
	}

	if regionContinent != drRegionContinent {
		diags.AddError(
			"Replication time control requires regions on the same continent",
			"enable_replication_time_control cannot be set to true for this region pair. Set it explicitly to false, or choose a dr_region on the same continent as region.",
		)
	}

	return diags
}

// azureEffectiveEnableReplicationTimeControl determines the value to send to the
// API for enable_replication_time_control on an Azure cluster.
//
// If the user set the value explicitly, it is sent as-is (validateAzureConfig
// already ensures an explicit `true` is only allowed when region and dr_region
// share a continent; `false` is always allowed). If the user left it unset, the
// API no longer defaults this to a region-aware value, so the provider computes
// one: `true` only when region and dr_region are on the same continent,
// otherwise nothing is sent (leaving the field, and Replication Time Control,
// off).
//
// data is populated from the plan (req.Plan.Get), not the config. This attribute
// is Optional+Computed with no plan modifiers, so when the user leaves it unset,
// the framework's default "proposed new state" logic plans it as Unknown, not
// Null — Null is really only ever seen here if it's read from config elsewhere.
// Both must be treated as "not explicitly set": bailing out only on Unknown (as
// this used to) skips the compute-from-continent logic below on every Create,
// since that IS the unset case at plan time.
func azureEffectiveEnableReplicationTimeControl(data *models.ClusterResource) *bool {
	if !data.EnableReplicationTimeControl.IsNull() && !data.EnableReplicationTimeControl.IsUnknown() {
		return data.EnableReplicationTimeControl.ValueBoolPointer()
	}

	if data.Region.IsUnknown() || data.DrRegion.IsNull() || data.DrRegion.IsUnknown() {
		return nil
	}

	regionContinent, regionOk := azureRegionContinent(data.Region.ValueString())
	if !regionOk {
		return nil
	}
	drRegionContinent, drRegionOk := azureRegionContinent(data.DrRegion.ValueString())
	if !drRegionOk || regionContinent != drRegionContinent {
		return nil
	}

	enable := true
	return &enable
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
			"tenant_id is not allowed for 'GCP' cluster",
			"Please remove tenant_id",
		)
	}
	if !data.SecondaryVpcCidr.IsNull() {
		diags.AddError(
			"secondary_vpc_cidr is not allowed for 'GCP' cluster",
			"Please remove secondary_vpc_cidr",
		)
	}
	if !data.DrSecondaryVpcCidr.IsNull() {
		diags.AddError(
			"dr_secondary_vpc_cidr is not allowed for 'GCP' cluster",
			"Please remove dr_secondary_vpc_cidr",
		)
	}

	// DR validation
	if !data.IsDrEnabled.IsNull() && data.IsDrEnabled.ValueBool() {
		if data.DrRegion.IsNull() || data.DrRegion.IsUnknown() {
			diags.AddError(
				"dr_region is required when is_dr_enabled is true",
				"Please set dr_region to the secondary region for Disaster Recovery",
			)
		}
		if data.DrVpcSubnetRange.IsNull() || data.DrVpcSubnetRange.IsUnknown() {
			diags.AddError(
				"dr_vpc_subnet_range is required for 'GCP' cluster when is_dr_enabled is true",
				"Please set dr_vpc_subnet_range",
			)
		}
		if data.DrPodSubnetRange.IsNull() || data.DrPodSubnetRange.IsUnknown() {
			diags.AddError(
				"dr_pod_subnet_range is required for 'GCP' cluster when is_dr_enabled is true",
				"Please set dr_pod_subnet_range",
			)
		}
		if data.DrServicePeeringRange.IsNull() || data.DrServicePeeringRange.IsUnknown() {
			diags.AddError(
				"dr_service_peering_range is required for 'GCP' cluster when is_dr_enabled is true",
				"Please set dr_service_peering_range",
			)
		}
		if data.DrServiceSubnetRange.IsNull() || data.DrServiceSubnetRange.IsUnknown() {
			diags.AddError(
				"dr_service_subnet_range is required for 'GCP' cluster when is_dr_enabled is true",
				"Please set dr_service_subnet_range",
			)
		}
	}

	if !data.IsDrEnabled.IsNull() && !data.IsDrEnabled.ValueBool() {
		if !data.DrVpcSubnetRange.IsNull() {
			diags.AddError(
				"dr_vpc_subnet_range is only valid when is_dr_enabled is true",
				"Please remove dr_vpc_subnet_range or set is_dr_enabled to true",
			)
		}
		if !data.EnableReplicationTimeControl.IsNull() {
			diags.AddError(
				"enable_replication_time_control is only valid when is_dr_enabled is true",
				"Please remove enable_replication_time_control or set is_dr_enabled to true",
			)
		}
		if !data.DrPodSubnetRange.IsNull() {
			diags.AddError(
				"dr_pod_subnet_range is only valid when is_dr_enabled is true",
				"Please remove dr_pod_subnet_range or set is_dr_enabled to true",
			)
		}
		if !data.DrServicePeeringRange.IsNull() {
			diags.AddError(
				"dr_service_peering_range is only valid when is_dr_enabled is true",
				"Please remove dr_service_peering_range or set is_dr_enabled to true",
			)
		}
		if !data.DrServiceSubnetRange.IsNull() {
			diags.AddError(
				"dr_service_subnet_range is only valid when is_dr_enabled is true",
				"Please remove dr_service_subnet_range or set is_dr_enabled to true",
			)
		}
	}

	return diags
}
