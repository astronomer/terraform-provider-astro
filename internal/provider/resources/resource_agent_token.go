package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AgentTokenResource{}
var _ resource.ResourceWithImportState = &AgentTokenResource{}
var _ resource.ResourceWithConfigure = &AgentTokenResource{}

func NewAgentTokenResource() resource.Resource {
	return &AgentTokenResource{}
}

// AgentTokenResource defines the resource implementation.
type AgentTokenResource struct {
	IamClient      *iam.ClientWithResponses
	OrganizationId string
}

func (r *AgentTokenResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_agent_token"
}

func (r *AgentTokenResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Agent Token resource",
		Attributes:          schemas.AgentTokenResourceSchemaAttributes(),
	}
}

func (r *AgentTokenResource) Configure(
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

	r.IamClient = apiClients.IamClient
	r.OrganizationId = apiClients.OrganizationId
}

func (r *AgentTokenResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.AgentTokenResource

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := iam.CreateAgentTokenJSONRequestBody{
		Name: data.Name.ValueString(),
	}
	if !data.Description.IsNull() {
		createRequest.Description = data.Description.ValueStringPointer()
	} else {
		createRequest.Description = lo.ToPtr("")
	}
	if !data.ExpiryPeriodInDays.IsNull() {
		createRequest.TokenExpiryPeriodInDays = lo.ToPtr(int(data.ExpiryPeriodInDays.ValueInt64()))
	}

	createResp, err := r.IamClient.CreateAgentTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.DeploymentId.ValueString(),
		createRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create agent token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create agent token, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, createResp.HTTPResponse, createResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if createResp.JSON200 == nil {
		tflog.Error(ctx, "failed to create agent token", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to create agent token, got nil response")
		return
	}
	if createResp.JSON200.Token == nil {
		tflog.Error(ctx, "failed to create agent token", map[string]interface{}{"error": "nil token value"})
		resp.Diagnostics.AddError("Client Error", "Unable to create agent token, got nil token value in response")
		return
	}

	data.ReadFromResponse(createResp.JSON200, *createResp.JSON200.Token)

	tflog.Trace(ctx, fmt.Sprintf("created an agent token resource: %v", data.Id.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentTokenResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.AgentTokenResource

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.IamClient.GetAgentTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.DeploymentId.ValueString(),
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get agent token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get agent token, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, getResp.HTTPResponse, getResp.Body)
	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	data.ReadFromResponse(getResp.JSON200, data.Token.ValueString())

	tflog.Trace(ctx, fmt.Sprintf("read an agent token resource: %v", data.Id.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op: all user-configurable fields use RequiresReplace, so Terraform
// will always destroy and recreate rather than calling Update.
func (r *AgentTokenResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError(
		"Agent Token does not support in-place updates",
		"All fields require replacement. This is a provider bug if Update was called.",
	)
}

func (r *AgentTokenResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.AgentTokenResource

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.IamClient.DeleteAgentTokenWithResponse(
		ctx,
		r.OrganizationId,
		data.DeploymentId.ValueString(),
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete agent token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete agent token, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, deleteResp.HTTPResponse, deleteResp.Body)
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted an agent token resource: %v", data.Id.ValueString()))
}

// ImportState expects a composite ID in the format "<deployment_id>/<token_id>".
func (r *AgentTokenResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: <deployment_id>/<token_id>, got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deployment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
