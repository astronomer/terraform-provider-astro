package resources

import (
	"context"
	"fmt"
	"net/http"

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

	role := data.Role.ValueString()
	if len(role) == 0 {
		role, _, diags = RequestApiTokenRole(ctx, roles, data.Type.ValueString())
		if diags != nil {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// Create the API token request
	createApiTokenRequest := iam.CreateApiTokenRequest{
		Name: data.Name.ValueString(),
		Role: role,
		Type: iam.CreateApiTokenRequestType(data.Type.ValueString()),
	}

	// If the entity type is WORKSPACE or DEPLOYMENT, set the entity id
	if createApiTokenRequest.Type == iam.WORKSPACE || createApiTokenRequest.Type == iam.DEPLOYMENT {
		var entityId string
		_, entityId, diags = RequestApiTokenRole(ctx, roles, data.Type.ValueString())
		if diags != nil {
			resp.Diagnostics.Append(diags...)
			return
		}

		createApiTokenRequest.EntityId = lo.ToPtr(entityId)
	}

	if data.Description.IsNull() {
		createApiTokenRequest.Description = lo.ToPtr("")
	} else {
		createApiTokenRequest.Description = data.Description.ValueStringPointer()
	}

	if !data.ExpiryPeriodInDays.IsNull() {
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

	// Get api token
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

	diags = data.ReadFromResponse(ctx, apiTokenResp.JSON200, data.Role.ValueStringPointer())
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

	diags := data.ReadFromResponse(ctx, apiToken.JSON200, data.Role.ValueStringPointer())
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
	var data models.ApiTokenResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform set of roles to API token roles
	roles, diags := RequestApiTokenRoles(ctx, data.Roles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
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

	// Get api token
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

	diags = data.ReadFromResponse(ctx, apiTokenResp.JSON200, data.Role.ValueStringPointer())
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

	tokenRole := data.Role.ValueString()
	if len(tokenRole) == 0 {
		tokenRole, _, diags = RequestApiTokenRole(ctx, roles, data.Type.ValueString())
		if diags != nil {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	entityType := data.Type.ValueString()

	// Check if the role is valid for the entity type
	if !utils.CheckRole(tokenRole, entityType) {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Role %s is not valid for entity type %s", tokenRole, entityType),
		)
		return
	}

	// Validate the roles based on the entity type
	switch entityType {
	case string(iam.ApiTokenRoleEntityTypeORGANIZATION):
		resp.Diagnostics.Append(validateOrganizationApiToken(roles)...)
	case string(iam.ApiTokenRoleEntityTypeWORKSPACE):
		resp.Diagnostics.Append(validateWorkspaceApiToken(roles)...)
	case string(iam.ApiTokenRoleEntityTypeDEPLOYMENT):
		resp.Diagnostics.Append(validateDeploymentApiToken(roles)...)
	}
}

func validateOrganizationApiToken(roles []iam.ApiTokenRole) diag.Diagnostics {
	for _, role := range roles {
		if utils.CheckRole(role.Role, string(role.EntityType)) {
			return nil
		}
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Client Error",
			"Unable to find the role for the entity type",
		),
	}
}

func validateWorkspaceApiToken(roles []iam.ApiTokenRole) diag.Diagnostics {
	for _, role := range roles {
		if role.EntityType == iam.ApiTokenRoleEntityTypeORGANIZATION {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Client Error",
					"API Token of type WORKSPACE cannot have an ORGANIZATION role",
				),
			}
		}

		if utils.CheckRole(role.Role, string(role.EntityType)) {
			return nil
		}
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Client Error",
			"Unable to find the role for the entity type",
		),
	}
}

func validateDeploymentApiToken(roles []iam.ApiTokenRole) diag.Diagnostics {
	for _, role := range roles {
		if role.EntityType != iam.ApiTokenRoleEntityTypeDEPLOYMENT {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Client Error",
					"API Token of type DEPLOYMENT cannot have an ORGANIZATION or WORKSPACE role",
				),
			}
		}

		if utils.CheckRole(role.Role, string(role.EntityType)) {
			return nil
		}
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Client Error",
			"Unable to find the role for the entity type",
		),
	}
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
		var entityType iam.ApiTokenRoleEntityType

		if v.EntityType.ValueString() == string(iam.ApiTokenRoleEntityTypeORGANIZATION) {
			entityType = iam.ApiTokenRoleEntityTypeORGANIZATION
		} else if v.EntityType.ValueString() == string(iam.ApiTokenRoleEntityTypeWORKSPACE) {
			entityType = iam.ApiTokenRoleEntityTypeWORKSPACE
		} else if v.EntityType.ValueString() == string(iam.ApiTokenRoleEntityTypeDEPLOYMENT) {
			entityType = iam.ApiTokenRoleEntityTypeDEPLOYMENT
		}

		return iam.ApiTokenRole{
			Role:       v.Role.ValueString(),
			EntityId:   v.EntityId.ValueString(),
			EntityType: entityType,
		}
	})

	return apiTokenRoles, nil
}

func RequestApiTokenRole(ctx context.Context, roles []iam.ApiTokenRole, entityType string) (string, string, diag.Diagnostics) {
	for _, role := range roles {
		if role.EntityType == iam.ApiTokenRoleEntityType(entityType) {
			return role.Role, role.EntityId, nil
		}
	}
	return "", "", diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Client Error",
			"Unable to find the role for the entity type",
		),
	}
}
