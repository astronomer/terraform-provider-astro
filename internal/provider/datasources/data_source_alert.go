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
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &alertDataSource{}
var _ datasource.DataSourceWithConfigure = &alertDataSource{}

func NewAlertDataSource() datasource.DataSource {
	return &alertDataSource{}
}

// alertDataSource defines the data source implementation.
type alertDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *alertDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (d *alertDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Alert data source",
		Attributes:          schemas.AlertDataSourceSchemaAttributes(),
	}
}

func (d *alertDataSource) Configure(
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

func (d *alertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.AlertDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := d.PlatformClient.GetAlertWithResponse(ctx, d.OrganizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "Failed to get alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read alert, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, alert.HTTPResponse, alert.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if alert.JSON200 == nil {
		tflog.Error(ctx, "failed to get alert", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to read alert, got nil response")
		return
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, alert.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
