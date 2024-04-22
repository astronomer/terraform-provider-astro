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
var _ datasource.DataSource = &clusterOptionsDataSource{}
var _ datasource.DataSourceWithConfigure = &clusterOptionsDataSource{}

func NewClusterOptionsDataSource() datasource.DataSource {
	return &clusterOptionsDataSource{}
}

// clusterOptionsDataSource defines the data source implementation.
type clusterOptionsDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *clusterOptionsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_cluster_options"
}

func (d *clusterOptionsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterOptions data source",
		Attributes:          schemas.ClusterOptionsDataSourceSchemaAttributes(),
	}
}

func (d *clusterOptionsDataSource) Configure(
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

func (d *clusterOptionsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.ClusterOptionsDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	provider := platform.GetClusterOptionsParamsProvider(data.CloudProvider.ValueString())
	params := &platform.GetClusterOptionsParams{
		Type:     platform.GetClusterOptionsParamsType(data.Type.ValueString()),
		Provider: &provider,
	}

	var clusterOptions []platform.ClusterOptions
	clusterOptionsResp, err := d.PlatformClient.GetClusterOptionsWithResponse(
		ctx,
		d.OrganizationId,
		params,
	)

	if err != nil {
		tflog.Error(ctx, "failed to list clusterOptions", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read clusterOptions, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, clusterOptionsResp.HTTPResponse, clusterOptionsResp.Body)

	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if clusterOptionsResp.JSON200 == nil {
		tflog.Error(ctx, "failed to list clusterOptions", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to read clusterOptions, got nil response")
		return
	}
	clusterOptions = append(clusterOptions, *clusterOptionsResp.JSON200...)

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, clusterOptions)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
