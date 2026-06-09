package datasources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &environmentObjectsDataSource{}
var _ datasource.DataSourceWithConfigure = &environmentObjectsDataSource{}

func NewEnvironmentObjectsDataSource() datasource.DataSource {
	return &environmentObjectsDataSource{}
}

// environmentObjectsDataSource defines the data source implementation.
type environmentObjectsDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *environmentObjectsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment_objects"
}

func (d *environmentObjectsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environment Objects data source. Lists environment objects with optional filters.",
		Attributes:          schemas.EnvironmentObjectsDataSourceSchemaAttributes(),
	}
}

func (d *environmentObjectsDataSource) Configure(
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

func (d *environmentObjectsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.EnvironmentObjects

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &platform.ListEnvironmentObjectsParams{
		Limit: lo.ToPtr(1000),
	}

	if !data.WorkspaceId.IsNull() && !data.WorkspaceId.IsUnknown() {
		params.WorkspaceId = data.WorkspaceId.ValueStringPointer()
	}
	if !data.DeploymentId.IsNull() && !data.DeploymentId.IsUnknown() {
		params.DeploymentId = data.DeploymentId.ValueStringPointer()
	}
	if !data.ObjectType.IsNull() && !data.ObjectType.IsUnknown() {
		ot := platform.ListEnvironmentObjectsParamsObjectType(data.ObjectType.ValueString())
		params.ObjectType = &ot
	}
	if !data.ObjectKey.IsNull() && !data.ObjectKey.IsUnknown() {
		params.ObjectKey = data.ObjectKey.ValueStringPointer()
	}
	if !data.ShowSecrets.IsNull() && !data.ShowSecrets.IsUnknown() {
		params.ShowSecrets = data.ShowSecrets.ValueBoolPointer()
	}
	if !data.ResolveLinked.IsNull() && !data.ResolveLinked.IsUnknown() {
		params.ResolveLinked = data.ResolveLinked.ValueBoolPointer()
	}

	var allObjects []platform.EnvironmentObject
	offset := 0
	for {
		params.Offset = &offset
		listResp, err := d.PlatformClient.ListEnvironmentObjectsWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list environment objects", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list environment objects: %s", err))
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, listResp.HTTPResponse, listResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if listResp.JSON200 == nil {
			tflog.Error(ctx, "failed to list environment objects", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to list environment objects, got nil response")
			return
		}

		// Defensive: an empty page after the first call means the server has nothing
		// more to return; without this guard a server that under-reports TotalCount
		// would loop forever.
		if len(listResp.JSON200.EnvironmentObjects) == 0 {
			break
		}

		allObjects = append(allObjects, listResp.JSON200.EnvironmentObjects...)

		if listResp.JSON200.TotalCount <= offset+len(listResp.JSON200.EnvironmentObjects) {
			break
		}

		// Advance by the actual page length (not the requested limit) so partial
		// pages don't skip records. Matches the pattern used in common/role.go.
		offset += len(listResp.JSON200.EnvironmentObjects)
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, allObjects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
