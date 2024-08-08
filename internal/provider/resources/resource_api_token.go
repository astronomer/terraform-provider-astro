package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"

	"github.com/astronomer/terraform-provider-astro/internal/provider/common"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApiTokenResource{}
var _ resource.ResourceWithImportState = &ApiTokenResource{}
var _ resource.ResourceWithConfigure = &ApiTokenResource{}
var _ resource.ResourceWithValidateConfig = &ApiTokenResource{}

func NewApiTokenResource() resource.Resource {
	return &ApiTokenResource{}
}

// ApiTokenResource defines the resource implementation.
type ApiTokenResource struct {
	IamClient      *iam.ClientWithResponses
	PlatformClient *platform.ClientWithResponses
	OrganizationId string
}

func (r *ApiTokenResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (r *ApiTokenResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "API Token resource",
		Attributes:          schemas.ApiTokenResourceSchemaAttributes(),
	}
}
func (r *ApiTokenResource) Configure(
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

	r.IamClient = apiClients.IamClient
	r.PlatformClient = apiClients.PlatformClient
	r.OrganizationId = apiClients.OrganizationId
}

func (r *ApiTokenResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.ApiTokenResource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics

	// Convert Terraform set of roles to API token roles
	roles, diags := RequestApiTokenRoles(ctx, data.Roles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Get the role for the entity type
	role, diags := RequestApiTokenPrimaryRole(roles, data.Type.ValueString())
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Validate organization id
	if string(role.EntityType) == string(iam.ORGANIZATION) {
		if role.EntityId != r.OrganizationId {
			resp.Diagnostics.AddError(
				"API Token of type 'ORGANIZATION' cannot have an 'ORGANIZATION' role with a different organization id",
				"Please provide a valid role for the entity type 'ORGANIZATION' with the correct organization id",
			)
			return
		}
	}

	// Validate workspaces
	workspaceRoles := FilterApiTokenRolesByType(roles, string(iam.WORKSPACE))
	diags = r.HasValidWorkspaces(ctx, workspaceRoles)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Validate deployments
	deploymentRoles := FilterApiTokenRolesByType(roles, string(iam.DEPLOYMENT))
	diags = r.HasValidDeployments(ctx, deploymentRoles)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create the API token request
	createApiTokenRequest := iam.CreateApiTokenRequest{
		Name: data.Name.ValueString(),
		Role: role.Role,
		Type: iam.CreateApiTokenRequestType(data.Type.ValueString()),
	}

	// If the entity type is WORKSPACE or DEPLOYMENT, set the entity id
	if createApiTokenRequest.Type == iam.WORKSPACE || createApiTokenRequest.Type == iam.DEPLOYMENT {
		createApiTokenRequest.EntityId = lo.ToPtr(role.EntityId)
	}

	if data.Description.IsNull() {
		createApiTokenRequest.Description = lo.ToPtr("")
	} else {
		createApiTokenRequest.Description = data.Description.ValueStringPointer()
	}

	if !data.ExpiryPeriodInDays.IsNull() && data.ExpiryPeriodInDays.ValueInt64() > 0 {
		createApiTokenRequest.TokenExpiryPeriodInDays = lo.ToPtr(int(data.ExpiryPeriodInDays.ValueInt64()))
	}

	apiToken, err := r.IamClient.CreateApiTokenWithResponse(
		ctx,
		r.OrganizationId,
		createApiTokenRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create API token, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, apiToken.HTTPResponse, apiToken.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	tokenId := apiToken.JSON200.Id

	// Update api token with additional roles
	if len(roles) > 1 {
		updateApiTokenRolesRequest := iam.UpdateApiTokenRolesRequest{
			Roles: roles,
		}
		updatedApiToken, err := r.IamClient.UpdateApiTokenRolesWithResponse(
			ctx,
			r.OrganizationId,
			tokenId,
			updateApiTokenRolesRequest,
		)
		if err != nil {
			tflog.Error(ctx, "failed to create API token", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create API token and add additional roles, got error: %s", err),
			)
			return
		}
		_, diagnostic = clients.NormalizeAPIError(ctx, updatedApiToken.HTTPResponse, updatedApiToken.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
	}

	// Get api token and use this as data since it will have the correct roles
	apiTokenResp, err := r.IamClient.GetApiTokenWithResponse(
		ctx,
		r.OrganizationId,
		tokenId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create API token and get API token, got error: %s", err),
		)
		return
	}

	diags = data.ReadFromResponse(ctx, apiTokenResp.JSON200, *apiToken.JSON200.Token)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created an API token resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiTokenResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.ApiTokenResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	apiToken, err := r.IamClient.GetApiTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)

	if err != nil {
		tflog.Error(ctx, "failed to get API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get API token, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, apiToken.HTTPResponse, apiToken.Body)
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

	diags := data.ReadFromResponse(ctx, apiToken.JSON200, data.Token.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read an API token resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiTokenResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data, currentState models.ApiTokenResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	req.State.Get(ctx, &currentState)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform set of roles to API token roles
	roles, diags := RequestApiTokenRoles(ctx, data.Roles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Get the role for the entity type
	role, diags := RequestApiTokenPrimaryRole(roles, data.Type.ValueString())
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Validate organization id
	if string(role.EntityType) == string(iam.ORGANIZATION) {
		if role.EntityId != r.OrganizationId {
			resp.Diagnostics.AddError(
				"API Token of type 'ORGANIZATION' cannot have an 'ORGANIZATION' role with a different organization id",
				"Please provide a valid role for the entity type 'ORGANIZATION' with the correct organization id",
			)
			return
		}
	}

	// Validate workspaces
	workspaceRoles := FilterApiTokenRolesByType(roles, string(iam.WORKSPACE))
	diags = r.HasValidWorkspaces(ctx, workspaceRoles)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Validate deployments
	deploymentRoles := FilterApiTokenRolesByType(roles, string(iam.DEPLOYMENT))
	diags = r.HasValidDeployments(ctx, deploymentRoles)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Validate token expiry
	if data.ExpiryPeriodInDays.ValueInt64() != currentState.ExpiryPeriodInDays.ValueInt64() {
		resp.Diagnostics.AddError(
			"API Token expiry period cannot be updated",
			"Please provide the same expiry period as the existing API token",
		)
		return
	}

	// Update API token roles
	updateApiTokenRolesRequest := iam.UpdateApiTokenRolesRequest{
		Roles: roles,
	}
	updatedApiToken, err := r.IamClient.UpdateApiTokenRolesWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		updateApiTokenRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update API token, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, updatedApiToken.HTTPResponse, updatedApiToken.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// update request
	updateApiTokenRequest := iam.UpdateApiTokenJSONRequestBody{
		Name: data.Name.ValueString(),
	}

	// description
	if !data.Description.IsNull() {
		updateApiTokenRequest.Description = data.Description.ValueStringPointer()
	} else {
		updateApiTokenRequest.Description = lo.ToPtr("")
	}

	apiToken, err := r.IamClient.UpdateApiTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		updateApiTokenRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update API token, got error: %s", err),
		)
		return
	}
	_, diagnostic = clients.NormalizeAPIError(ctx, apiToken.HTTPResponse, apiToken.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Get api token and use this as data since it will have the correct roles
	apiTokenResp, err := r.IamClient.GetApiTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to update API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update API token and get API token, got error: %s", err),
		)
		return
	}

	diags = data.ReadFromResponse(ctx, apiTokenResp.JSON200, data.Token.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated an API token resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiTokenResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.ApiTokenResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	apiToken, err := r.IamClient.DeleteApiTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete API token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete API token, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, apiToken.HTTPResponse, apiToken.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted an API token resource: %v", data.Id.ValueString()))
}

func (r *ApiTokenResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ApiTokenResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.ApiTokenResource

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform set of roles to API token roles
	roles, diags := RequestApiTokenRoles(ctx, data.Roles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tokenRole, diags := RequestApiTokenPrimaryRole(roles, data.Type.ValueString())
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	entityType := data.Type.ValueString()

	// Check if the role is valid for the token entity type
	if !common.ValidateRoleMatchesEntityType(tokenRole.Role, entityType) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Role '%s' is not valid for token type '%s'", tokenRole.Role, entityType),
			fmt.Sprintf("Please provide a valid role for the entity type '%s'", entityType),
		)
		return
	}

	diags = r.ValidateApiTokenRoles(entityType, roles)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (r *ApiTokenResource) ValidateApiTokenRoles(entityType string, roles []iam.ApiTokenRole) diag.Diagnostics {
	var numRolesMatchingEntityType int
	var invalidRoleError string

	for _, role := range roles {
		if entityType == string(iam.ApiTokenRoleEntityTypeWORKSPACE) && role.EntityType == iam.ApiTokenRoleEntityTypeORGANIZATION {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"API Token of type 'WORKSPACE' cannot have an 'ORGANIZATION' role",
					"Please remove the 'ORGANIZATION' role from the 'roles' list",
				),
			}
		}

		if entityType == string(iam.ApiTokenRoleEntityTypeDEPLOYMENT) && role.EntityType != iam.ApiTokenRoleEntityTypeDEPLOYMENT {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"API Token of type 'DEPLOYMENT' cannot have an 'ORGANIZATION' or 'WORKSPACE' role",
					"Please remove the 'ORGANIZATION' or 'WORKSPACE' role from the 'roles' list",
				),
			}
		}

		if !common.ValidateRoleMatchesEntityType(role.Role, string(role.EntityType)) {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					fmt.Sprintf("Role '%s' is not valid for entity type '%s'", role.Role, role.EntityType),
					fmt.Sprintf("Please provide a valid role for the entity type '%s'", role.EntityType),
				),
			}
		}

		if common.ValidateRoleMatchesEntityType(role.Role, entityType) {
			numRolesMatchingEntityType++
		}
	}

	switch entityType {
	case string(iam.ApiTokenRoleEntityTypeORGANIZATION):
		invalidRoleError = "There is no 'ORGANIZATION' role in 'roles'"
	case string(iam.ApiTokenRoleEntityTypeWORKSPACE):
		invalidRoleError = "There is no 'WORKSPACE' role in 'roles'"
	case string(iam.ApiTokenRoleEntityTypeDEPLOYMENT):
		invalidRoleError = "There is no 'DEPLOYMENT' role in 'roles'"
	}

	if numRolesMatchingEntityType > 1 {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("API Token of type '%s' cannot have more than one role of the same type", entityType),
				"Please provide only one role for the entity type",
			),
		}
	} else if numRolesMatchingEntityType < 1 {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				invalidRoleError,
				fmt.Sprintf("Please provide a valid role for the entity type '%s'", entityType),
			),
		}
	}

	return nil
}

// RequestApiTokenRoles converts a Terraform set to a list of iam.ApiTokenRole to be used in create and update requests
func RequestApiTokenRoles(ctx context.Context, apiTokenRolesObjSet types.Set) ([]iam.ApiTokenRole, diag.Diagnostics) {
	if len(apiTokenRolesObjSet.Elements()) == 0 {
		return []iam.ApiTokenRole{}, nil
	}

	var roles []models.ApiTokenRole
	diags := apiTokenRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	apiTokenRoles := lo.Map(roles, func(v models.ApiTokenRole, _ int) iam.ApiTokenRole {
		return iam.ApiTokenRole{
			Role:       v.Role.ValueString(),
			EntityId:   v.EntityId.ValueString(),
			EntityType: iam.ApiTokenRoleEntityType(v.EntityType.ValueString()),
		}
	})

	return apiTokenRoles, nil
}

func RequestApiTokenPrimaryRole(roles []iam.ApiTokenRole, entityType string) (iam.ApiTokenRole, diag.Diagnostics) {
	for _, role := range roles {
		if role.EntityType == iam.ApiTokenRoleEntityType(entityType) {
			return role, nil
		}
	}
	return iam.ApiTokenRole{}, diag.Diagnostics{
		diag.NewErrorDiagnostic(
			fmt.Sprintf("No matching role found for the specified entity type '%s'", entityType),
			fmt.Sprintf("Please provide a valid role for the API token entity type '%s'", entityType),
		),
	}
}

func FilterApiTokenRolesByType(roles []iam.ApiTokenRole, entityType string) []iam.ApiTokenRole {
	var filteredRoles []iam.ApiTokenRole
	for _, role := range roles {
		if role.EntityType == iam.ApiTokenRoleEntityType(entityType) {
			filteredRoles = append(filteredRoles, role)
		}
	}
	return filteredRoles
}

func (r *ApiTokenResource) HasValidWorkspaces(ctx context.Context, workspaceRoles []iam.ApiTokenRole) diag.Diagnostics {
	if len(workspaceRoles) == 0 {
		return nil
	}
	// Get workspace ids
	var workspaceIds []string
	for _, workspaceRole := range workspaceRoles {
		workspaceIds = append(workspaceIds, workspaceRole.EntityId)
	}

	listWorkspacesRequest := platform.ListWorkspacesParams{
		WorkspaceIds: lo.ToPtr(workspaceIds),
	}

	// List organization workspaces
	workspaces, err := r.PlatformClient.ListWorkspacesWithResponse(
		ctx,
		r.OrganizationId,
		&listWorkspacesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to list workspaces", map[string]interface{}{"error": err})
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to list workspaces, got error: %s", err),
			),
		}
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, workspaces.HTTPResponse, workspaces.Body)
	if diagnostic != nil {
		return diag.Diagnostics{diagnostic}
	}
	organizationWorkspaceIds := lo.Map(workspaces.JSON200.Workspaces, func(workspace platform.Workspace, _ int) string {
		return workspace.Id
	})

	invalidWorkspaceIds, _ := lo.Difference(workspaceIds, organizationWorkspaceIds)
	if len(invalidWorkspaceIds) > 0 {
		tflog.Error(ctx, "invalid workspace ids")
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"One or more workspaces is not in the organization, cannot set roles for workspaces that do not exist",
				fmt.Sprintf("The following workspace ids are invalid: %v", invalidWorkspaceIds),
			),
		}
	}

	return nil
}

func (r *ApiTokenResource) HasValidDeployments(ctx context.Context, deploymentRoles []iam.ApiTokenRole) diag.Diagnostics {
	if len(deploymentRoles) == 0 {
		return nil
	}
	// Get deployment ids
	var deploymentIds []string
	for _, deploymentRole := range deploymentRoles {
		deploymentIds = append(deploymentIds, deploymentRole.EntityId)
	}

	listDeploymentsRequest := platform.ListDeploymentsParams{
		DeploymentIds: lo.ToPtr(deploymentIds),
	}

	// List organization deployments
	deployments, err := r.PlatformClient.ListDeploymentsWithResponse(
		ctx,
		r.OrganizationId,
		&listDeploymentsRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to list deployments", map[string]interface{}{"error": err})
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to list deployments, got error: %s", err),
			),
		}
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, deployments.HTTPResponse, deployments.Body)
	if diagnostic != nil {
		return diag.Diagnostics{diagnostic}
	}
	organizationDeploymentIds := lo.Map(deployments.JSON200.Deployments, func(deployment platform.Deployment, _ int) string {
		return deployment.Id
	})

	invalidDeploymentIds, _ := lo.Difference(deploymentIds, organizationDeploymentIds)
	if len(invalidDeploymentIds) > 0 {
		tflog.Error(ctx, "invalid deployment ids")
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"One or more deployments is not in the organization, cannot set roles for deployments that do not exist",
				fmt.Sprintf("The following deployment ids are invalid: %v", invalidDeploymentIds),
			),
		}
	}

	return nil
}

func ApiTokenRolesToWorkspaceRoles(roles []iam.ApiTokenRole) []iam.WorkspaceRole {
	var workspaceRoles []iam.WorkspaceRole
	for _, role := range roles {
		if string(role.EntityType) == string(iam.WORKSPACE) {
			workspaceRoles = append(workspaceRoles, iam.WorkspaceRole{
				WorkspaceId: role.EntityId,
				Role:        iam.WorkspaceRoleRole(role.Role),
			})
		}
	}
	return workspaceRoles
}

func ApiTokenRolesToDeploymentRoles(roles []iam.ApiTokenRole) []iam.DeploymentRole {
	var deploymentRoles []iam.DeploymentRole
	for _, role := range roles {
		if string(role.EntityType) == string(iam.DEPLOYMENT) {
			deploymentRoles = append(deploymentRoles, iam.DeploymentRole{
				DeploymentId: role.EntityId,
				Role:         role.Role,
			})
		}
	}
	return deploymentRoles
}
