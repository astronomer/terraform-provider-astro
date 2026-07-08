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
var _ datasource.DataSource = &usersListDataSource{}
var _ datasource.DataSourceWithConfigure = &usersListDataSource{}

func NewUsersListDataSource() datasource.DataSource {
	return &usersListDataSource{}
}

// usersListDataSource is a List-based variant of the astro_users data source.
// It returns identical data but represents the users collection as an ordered
// List instead of a Set, which avoids the expensive nested-object hashing the
// framework performs for Sets and dramatically reduces plan time for large
// organizations (~30s -> ~1s at ~1000 users). The list is sorted by id so
// pagination and plan output are deterministic across reads.
type usersListDataSource struct {
	IamClient      iam.ClientWithResponsesInterface
	OrganizationId string
}

func (d *usersListDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_users_list"
}

func (d *usersListDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Users data source (list variant). Identical to `astro_users` but returns the `users` collection as an ordered list instead of a set, for significantly better `terraform plan` performance on large organizations.",
		Attributes:          schemas.UsersListDataSourceSchemaAttributes(),
	}
}

func (d *usersListDataSource) Configure(
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

func (d *usersListDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.UsersList

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &iam.ListUsersParams{
		Limit: lo.ToPtr(1000),
		// Sort by id (immutable, unique) so pagination is stable across calls and
		// the resulting list ordering is deterministic between plans. Without this,
		// the API's default ordering can shift between reads, producing spurious
		// plan diffs for an ordered List.
		Sorts: &[]iam.ListUsersParamsSorts{iam.ListUsersParamsSortsIdAsc},
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
