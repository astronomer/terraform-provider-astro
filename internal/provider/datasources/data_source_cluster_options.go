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
var _ datasource.DataSource = &clusterOptionsDataSource{}
var _ datasource.DataSourceWithConfigure = &clusterOptionsDataSource{}

func NewClusterOptionsDataSource() datasource.DataSource {
	return &clusterOptionsDataSource{}
}

// clusterOptionsDataSource defines the data source implementation.
type clusterOptionsDataSource struct {
	PlatformV1Client platform_v1.ClientWithResponsesInterface
	OrganizationId   string
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

	d.PlatformV1Client = apiClients.PlatformV1Client
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
	provider := platform_v1.GetClusterOptionsParamsProvider(data.CloudProvider.ValueString())
	params := &platform_v1.GetClusterOptionsParams{
		Type:     platform_v1.GetClusterOptionsParamsType(data.Type.ValueString()),
		Provider: &provider,
	}

	var clusterOptions []platform_v1.ClusterOptions
	clusterOptionsResp, err := d.PlatformV1Client.GetClusterOptionsWithResponse(
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
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, clusterOptionsResp.HTTPResponse, clusterOptionsResp.Body, clusterOptionsResp.JSON200, "read cluster options")

	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
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
