package datasources

import (
	"context"
	"fmt"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &workspaceDataSource{}
var _ datasource.DataSourceWithConfigure = &workspaceDataSource{}

func NewWorkspaceDataSource() datasource.DataSource {
	return &workspaceDataSource{}
}

// workspaceDataSource defines the data source implementation.
type workspaceDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *workspaceDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *workspaceDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Workspace data source",
		Attributes:          schemas.WorkspaceDataSourceSchemaAttributes(),
	}
}

func (d *workspaceDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.DataSourceApiClientConfigureError(ctx, req, resp)
		return
	}

	d.PlatformClient = apiClients.PlatformClient
	d.OrganizationId = apiClients.OrganizationId
}

func (d *workspaceDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.WorkspaceDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workspace, err := d.PlatformClient.GetWorkspaceWithResponse(
		ctx,
		d.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get workspace", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read workspace, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, workspace.HTTPResponse, workspace.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if workspace.JSON200 == nil {
		tflog.Error(ctx, "failed to get workspace", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to read workspace, got nil response")
		return
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, workspace.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
