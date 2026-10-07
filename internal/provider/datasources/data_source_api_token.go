package datasources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &apiTokenDataSource{}
var _ datasource.DataSourceWithConfigure = &apiTokenDataSource{}

func NewApiTokenDataSource() datasource.DataSource {
	return &apiTokenDataSource{}
}

// apiTokenDataSource defines the data source implementation.
type apiTokenDataSource struct {
	PlatformV1Client platform_v1.ClientWithResponsesInterface
	OrganizationId   string
}

func (d *apiTokenDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (d *apiTokenDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "API Token data source",
		Attributes:          schemas.ApiTokenDataSourceSchemaAttributes(),
	}
}

func (d *apiTokenDataSource) Configure(
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

	d.PlatformV1Client = apiClients.PlatformV1Client
	d.OrganizationId = apiClients.OrganizationId
}

func (d *apiTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.ApiTokenDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiToken, err := d.PlatformV1Client.GetApiTokenWithResponse(ctx, d.OrganizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "Failed to get api token", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read api token, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, apiToken.HTTPResponse, apiToken.Body, apiToken.JSON200, "read API token")
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, apiToken.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
