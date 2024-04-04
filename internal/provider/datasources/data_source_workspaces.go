package datasources

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/samber/lo"
	"strings"

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
var _ datasource.DataSource = &workspacesDataSource{}
var _ datasource.DataSourceWithConfigure = &workspacesDataSource{}

func NewWorkspacesDataSource() datasource.DataSource {
	return &workspacesDataSource{}
}

// workspacesDataSource defines the data source implementation.
type workspacesDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *workspacesDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_workspaces"
}

func (d *workspacesDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Workspaces data source",
		Attributes:          schemas.WorkspacesDataSourceSchemaAttributes(),
	}
}

func (d *workspacesDataSource) Configure(
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

func (d *workspacesDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.WorkspacesDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	params := &platform.ListWorkspacesParams{
		Limit: lo.ToPtr(1000),
	}
	workspaceIds := data.WorkspaceIds.Elements()
	if len(workspaceIds) > 0 {
		workspaceIdsParam := lo.Map(workspaceIds, func(id attr.Value, _ int) string {
			// Terraform includes quotes around the string, so we need to remove them
			return strings.ReplaceAll(id.String(), `"`, "")
		})
		params.WorkspaceIds = &workspaceIdsParam
	}
	names := data.Names.Elements()
	if len(names) > 0 {
		namesParam := lo.Map(names, func(name attr.Value, _ int) string {
			// Terraform includes quotes around the string, so we need to remove them
			return strings.ReplaceAll(name.String(), `"`, "")
		})
		params.Names = &namesParam
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var workspaces []platform.Workspace
	offset := 0
	for {
		params.Offset = &offset
		workspacesResp, err := d.PlatformClient.ListWorkspacesWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to get workspace", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read workspace, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, workspacesResp.HTTPResponse, workspacesResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if workspacesResp.JSON200 == nil {
			tflog.Error(ctx, "failed to get workspace", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read workspace, got nil response")
			return
		}

		workspaces = append(workspaces, workspacesResp.JSON200.Workspaces...)

		if workspacesResp.JSON200.TotalCount <= offset {
			break
		}

		offset += 1000
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, workspaces)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
