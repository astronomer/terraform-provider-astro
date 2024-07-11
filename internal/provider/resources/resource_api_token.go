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

func NewApiTokenResource() *ApiTokenResource {
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

	roles, _ := RequestApiTokenRoles(ctx, data.Roles)

	// Create the API token request
	createApiTokenRequest := iam.CreateApiTokenRequest{
		Name:                    data.Name.ValueString(),
		Description:             data.Description.ValueStringPointer(),
		Role:                    roles[0].Role,
		TokenExpiryPeriodInDays: lo.ToPtr(int(data.ExpiryPeriodInDays.ValueInt64())),
	}

	// If the entity type is WORKSPACE or DEPLOYMENT, set the entity id
	if data.Type.ValueString() == string(iam.WORKSPACE) || data.Type.ValueString() == string(iam.DEPLOYMENT) {
		createApiTokenRequest.EntityId = lo.ToPtr(roles[0].EntityId)
	}

	if data.Type.ValueString() == string(iam.ORGANIZATION) {
		createApiTokenRequest.Type = iam.ORGANIZATION
	} else if data.Type.ValueString() == string(iam.WORKSPACE) {
		createApiTokenRequest.Type = iam.WORKSPACE
	} else if data.Type.ValueString() == string(iam.DEPLOYMENT) {
		createApiTokenRequest.Type = iam.DEPLOYMENT
	} else {
		tflog.Error(ctx, "failed to create api_token", map[string]interface{}{"error": "Invalid entity type"})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to create api_token, got error: Invalid entity type"),
		)
		return
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

	diags = data.ReadFromResponse(ctx, apiToken.JSON200)
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

	diags := data.ReadFromResponse(ctx, apiToken.JSON200)
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

	// update request
	updateApiTokenRequest := iam.UpdateApiTokenJSONRequestBody{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueStringPointer(),
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
	_, diagnostic := clients.NormalizeAPIError(ctx, apiToken.HTTPResponse, apiToken.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, apiToken.JSON200)
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
