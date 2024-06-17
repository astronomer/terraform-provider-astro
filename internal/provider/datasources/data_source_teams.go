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
var _ datasource.DataSource = &teamsDataSource{}
var _ datasource.DataSourceWithConfigure = &teamsDataSource{}

func NewTeamsDataSource() datasource.DataSource {
	return &teamsDataSource{}
}

// teamsDataSource defines the data source implementation.
type teamsDataSource struct {
	IamClient      iam.ClientWithResponsesInterface
	OrganizationId string
}

func (d *teamsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *teamsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Teams data source",
		Attributes:          schemas.TeamsDataSourceSchemaAttributes(),
	}
}

func (d *teamsDataSource) Configure(
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

func (d *teamsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.Teams

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &iam.ListTeamsParams{
		Limit: lo.ToPtr(1000),
	}
	var diags diag.Diagnostics
	names, diags := utils.TypesSetToStringSlice(ctx, data.Names)
	if len(names) > 0 {
		params.Names = &names
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var teams []iam.Team
	offset := 0
	for {
		params.Offset = &offset
		teamsResp, err := d.IamClient.ListTeamsWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list teams", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read teams, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if teamsResp.JSON200 == nil {
			tflog.Error(ctx, "failed to list teams", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read teams, got nil response")
			return
		}

		teams = append(teams, teamsResp.JSON200.Teams...)

		if teamsResp.JSON200.TotalCount <= offset {
			break
		}

		offset += 1000
	}

	// Populate the model with the response data
	diags = data.ReadFromResponse(ctx, teams)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
