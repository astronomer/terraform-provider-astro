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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &teamDataSource{}
var _ datasource.DataSourceWithConfigure = &teamDataSource{}

func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

// teamDataSource defines the data source implementation.
type teamDataSource struct {
	IamClient      iam.ClientWithResponsesInterface
	OrganizationId string
}

func (d *teamDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Team data source",
		Attributes:          schemas.TeamDataSourceSchemaAttributes(),
	}
}

func (d *teamDataSource) Configure(
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

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.TeamDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := d.IamClient.GetTeamWithResponse(ctx, d.OrganizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "Failed to get team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read team, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, team.HTTPResponse, team.Body, team.JSON200, "read team")
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	teamMembers, err := d.IamClient.ListTeamMembersWithResponse(ctx, d.OrganizationId, data.Id.ValueString(), nil)
	if err != nil {
		tflog.Error(ctx, "Failed to get team members", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read team members, got error: %s", err),
		)
		return
	}
	_, diagnostic = clients.NormalizeAPIResponseWithBody(ctx, teamMembers.HTTPResponse, teamMembers.Body, teamMembers.JSON200, "read team members")
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, team.JSON200, &teamMembers.JSON200.TeamMembers)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
