package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
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
var _ resource.ResourceWithValidateConfig = &EnvironmentObjectResource{}

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
		Attributes:          schemas.EnvironmentObjectResourceSchemaAttributes(),
	}
}

func (r *EnvironmentObjectResource) Configure(
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
}

func (r *EnvironmentObjectResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.EnvironmentObjectResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create environment object request
	envObject, err := r.platformClient.CreateEnvironmentObjectWithResponse(
		ctx,
		data.ObjectKey.ValueString(),
		platform.CreateEnvironmentObjectRequest{
			ObjectType:    platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()),
			Scope:         platform.CreateEnvironmentObjectRequestScope(data.Scope.ValueString()),
			ScopeEntityId: data.ScopeEntityId.ValueString(),
		},
	)
	if err != nil {
		tflog.Error(ctx, "failed to create environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create environment object, got error: %s", err),
		)
		return
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, envObject.HTTPResponse, envObject.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// diags := data.ReadFromResponse(ctx, envObject.JSON200)
	// if diags.HasError() {
	// 	resp.Diagnostics.Append(diags...)
	// 	return
	// }

	tflog.Trace(ctx, fmt.Sprintf("created an environment object resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentObjectResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.EnvironmentObjectResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environment object request
	envObject, err := r.platformClient.GetEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get environment object, got error: %s", err),
		)
		return
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, envObject.HTTPResponse, envObject.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, envObject.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read an environment object resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentObjectResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.EnvironmentObjectResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update environment object request
	envObject, err := r.platformClient.UpdateEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
		platform.UpdateEnvironmentObjectRequest{},
	)
	if err != nil {
		tflog.Error(ctx, "failed to update environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update environment object, got error: %s", err),
		)
		return
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, envObject.HTTPResponse, envObject.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, envObject.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated an environment object resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentObjectResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.EnvironmentObjectResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete environment object request
	envObject, err := r.platformClient.DeleteEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete environment object, got error: %s", err),
		)
		return
	}

	statusCode, diagnostic := clients.NormalizeAPIError(ctx, envObject.HTTPResponse, envObject.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted an environment object resource: %v", data.Id.ValueString()))
}

func (r *EnvironmentObjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig validates the configuration of the resource as a whole before any operations are performed.
// This is a good place to check for any conflicting settings.
func (r *EnvironmentObjectResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.EnvironmentObjectResource

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add any specific validation logic here
}
