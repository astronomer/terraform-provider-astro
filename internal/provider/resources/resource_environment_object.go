package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &EnvironmentObjectResource{}
var _ resource.ResourceWithImportState = &EnvironmentObjectResource{}
var _ resource.ResourceWithConfigure = &EnvironmentObjectResource{}

func NewEnvironmentObjectResource() resource.Resource {
	return &EnvironmentObjectResource{}
}

// EnvironmentObjectResource defines the resource implementation.
type EnvironmentObjectResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *EnvironmentObjectResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment_object"
}

func (r *EnvironmentObjectResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Environment Object resource",
		Attributes:          schemas.EnvironmentObjectSchemaAttributes(),
	}
}

func (r *EnvironmentObjectResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
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

func (r *EnvironmentObjectResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.EnvironmentObject

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read Terraform plan data into the model
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := platform.CreateEnvironmentObjectRequest{
		ObjectKey:     data.ObjectKey.ValueString(),
		ObjectType:    data.ObjectType.ValueString(),
		Scope:         data.Scope.ValueString(),
		ScopeEntityId: data.ScopeEntityId.ValueString(),
	}

	// Add additional fields based on ObjectType and Scope if necessary

	response, err := r.platformClient.CreateEnvironmentObjectWithResponse(ctx, r.organizationId, createRequest)
	if err != nil {
		tflog.Error(ctx, "failed to create environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create environment object, got error: %s", err),
		)
		return
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, response.HTTPResponse, response.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Read response into data
	// data.ReadFromResponse(ctx, response.JSON200)

	tflog.Trace(ctx, fmt.Sprintf("created an environment object resource: %v", data.Id.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save data into Terraform state
}

func (r *EnvironmentObjectResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Implement the Read method
}

func (r *EnvironmentObjectResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Implement the Update method
}

func (r *EnvironmentObjectResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Implement the Delete method
}

func (r *EnvironmentObjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Implement the ImportState method
}

func (r *EnvironmentObjectResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	// Implement the ValidateConfig method
}
