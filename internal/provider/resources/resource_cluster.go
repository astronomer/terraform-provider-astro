package resources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
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
		MarkdownDescription: "Cluster resource",
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

	var diags diag.Diagnostics
	var createClusterRequest platform.CreateClusterRequest

	switch platform.ClusterCloudProvider(data.CloudProvider.ValueString()) {
	case platform.ClusterCloudProviderAWS:
		createAwsDedicatedClusterRequest := platform.CreateAwsClusterRequest{
			CloudProvider:   platform.CreateAwsClusterRequestCloudProvider(data.CloudProvider.ValueString()),
			DbInstanceType:  data.DbInstanceType.ValueStringPointer(),
			Name:            data.Name.ValueString(),
			NodePools:       nil,
			ProviderAccount: data.ProviderAccount.ValueStringPointer(),
			Region:          data.Region.ValueString(),
			Type:            platform.CreateAwsClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:  data.VpcSubnetRange.ValueString(),
		}

		// workspaceIds
		createAwsDedicatedClusterRequest.WorkspaceIds, diags = utils.TypesListToStringSlicePtr(ctx, data.WorkspaceIds)
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
			DbInstanceType:  data.DbInstanceType.ValueStringPointer(),
			Name:            data.Name.ValueString(),
			NodePools:       nil,
			ProviderAccount: data.ProviderAccount.ValueStringPointer(),
			Region:          data.Region.ValueString(),
			TenantId:        data.TenantId.ValueStringPointer(),
			Type:            platform.CreateAzureClusterRequestType(data.Type.ValueString()),
			VpcSubnetRange:  data.VpcSubnetRange.ValueString(),
		}

		// workspaceIds
		createAzureDedicatedClusterRequest.WorkspaceIds, diags = utils.TypesListToStringSlicePtr(ctx, data.WorkspaceIds)
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
			DbInstanceType:      data.DbInstanceType.ValueStringPointer(),
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
		createGcpDedicatedClusterRequest.WorkspaceIds, diags = utils.TypesListToStringSlicePtr(ctx, data.WorkspaceIds)
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
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING)},
		Target:     []string{string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED)},
		Refresh:    r.resourceRefreshFunc(ctx, cluster.JSON200.Id),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
		Delay:      5 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster creation failed", err.Error())
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
		ClusterType:    (*platform.UpdateDedicatedClusterRequestClusterType)(data.Type.ValueStringPointer()),
		DbInstanceType: data.DbInstanceType.ValueStringPointer(),
		K8sTags:        []platform.ClusterK8sTag{},
		Name:           data.Name.ValueString(),
		NodePools:      nil,
		WorkspaceIds:   nil,
	}

	// workspaceIds
	updateDedicatedClusterRequest.WorkspaceIds, diags = utils.TypesListToStringSlicePtr(ctx, data.WorkspaceIds)
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

	cluster, err := r.platformClient.UpdateClusterWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
		updateClusterRequest,
	)
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
		Delay:      5 * time.Minute,
	}

	// readyCluster is the final state of the cluster after it has reached a target status
	readyCluster, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster update failed", err.Error())
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
	updateTimeout, diags := data.Timeouts.Update(ctx, 1*time.Hour)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// delete request
	cluster, err := r.platformClient.DeleteClusterWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete cluster", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete cluster, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Wait for the cluster to be deleted
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(platform.ClusterStatusCREATING), string(platform.ClusterStatusUPDATING), string(platform.ClusterStatusCREATED), string(platform.ClusterStatusUPDATEFAILED), string(platform.ClusterStatusCREATEFAILED)},
		Target:     []string{"DELETED"},
		Refresh:    r.resourceRefreshFunc(ctx, data.Id.ValueString()),
		Timeout:    1 * time.Hour,
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster deletion failed", err.Error())
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
	var diags diag.Diagnostics

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
	var diags diag.Diagnostics

	// Unallowed values
	if !data.TenantId.IsNull() {
		diags.AddError(
			"tenant_id is not allowed for 'AWS' cluster",
			"Please remove tenant_id",
		)
	}
	return diags
}

func validateGcpConfig(ctx context.Context, data *models.ClusterResource) diag.Diagnostics {
	var diags diag.Diagnostics

	// Unallowed values
	if !data.TenantId.IsNull() {
		diags.AddError(
			"tenant_id is not allowed for 'AWS' cluster",
			"Please remove tenant_id",
		)
	}
	return diags
}

// RequestClusterTags converts a Terraform list to a list of platform.ClusterTagRequest to be used in create and update requests
func RequestClusterTags(ctx context.Context, tagsObjList types.List) ([]platform.ClusterK8sTag, diag.Diagnostics) {
	if len(tagsObjList.Elements()) == 0 {
		return []platform.ClusterK8sTag{}, nil
	}

	var clusterTags []models.ClusterTag
	diags := tagsObjList.ElementsAs(ctx, &clusterTags, false)
	if diags.HasError() {
		return nil, diags
	}
	platformClusterTags := lo.Map(clusterTags, func(envVar models.ClusterTag, _ int) platform.ClusterK8sTag {
		return platform.ClusterK8sTag{
			Key:   envVar.Key.ValueStringPointer(),
			Value: envVar.Value.ValueStringPointer(),
		}
	})
	return platformClusterTags, nil
}

// resourceRefreshFunc returns a retry.StateRefreshFunc that polls the platform API for the cluster status
// If the cluster is not found, it returns "DELETED" status
// If the cluster is found, it returns the cluster status
// If there is an error, it returns the error
// WaitForStateContext will keep polling until the target status is reached, the timeout is reached or an err is returned
func (r *ClusterResource) resourceRefreshFunc(ctx context.Context, clusterId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		cluster, err := r.platformClient.GetClusterWithResponse(ctx, r.organizationId, clusterId)
		if err != nil {
			tflog.Error(ctx, "failed to get cluster while polling for cluster 'CREATED' status", map[string]interface{}{"error": err})
			return nil, "", err
		}
		statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
		if statusCode == http.StatusNotFound {
			return nil, "DELETED", nil
		}
		if diagnostic != nil {
			return nil, "", fmt.Errorf("error getting cluster %s", diagnostic.Detail())
		}
		return cluster.JSON200, string(cluster.JSON200.Status), nil
	}
}
