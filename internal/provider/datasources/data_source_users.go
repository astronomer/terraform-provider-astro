package datasources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &usersDataSource{}
var _ datasource.DataSourceWithConfigure = &usersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// usersDataSource defines the data source implementation.
type usersDataSource struct {
	IamClient      iam.ClientWithResponsesInterface
	OrganizationId string
}

func (d *usersDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Users data source",
		Attributes:          schemas.UsersDataSourceSchemaAttributes(),
	}
}

func (d *usersDataSource) Configure(
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

	d.IamClient = apiClients.IamClient
	d.OrganizationId = apiClients.OrganizationId
}

func (d *usersDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.Users

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &iam.ListUsersParams{
		Limit: lo.ToPtr(1000),
	}
	var diags diag.Diagnostics
	workspaceId := data.WorkspaceId.ValueString()
	if workspaceId != "" {
		params.WorkspaceId = &workspaceId
	}
	deploymentId := data.DeploymentId.ValueString()
	if deploymentId != "" {
		params.DeploymentId = &deploymentId
	}

	var users []iam.User
	offset := 0
	for {
		params.Offset = &offset
		usersResp, err := d.IamClient.ListUsersWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list users", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read users, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, usersResp.HTTPResponse, usersResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if usersResp.JSON200 == nil {
			tflog.Error(ctx, "failed to list users", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read users, got nil response")
			return
		}

		users = append(users, usersResp.JSON200.Users...)

		if usersResp.JSON200.TotalCount <= offset {
			break
		}

		offset += 1000
	}

	// Populate the model with the response data
	diags = data.ReadFromResponse(ctx, users)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
