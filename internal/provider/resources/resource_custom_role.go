package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &customRoleResource{}
var _ resource.ResourceWithImportState = &customRoleResource{}
var _ resource.ResourceWithConfigure = &customRoleResource{}
var _ resource.ResourceWithModifyPlan = &customRoleResource{}

const (
	maxListedTokensInDiagnostic        = 10
	directAccessTokenLookupTimeout     = 10 * time.Second
	directAccessTokenListPageSize      = 1000
	directAccessTokenLookupConcurrency = 8
)

func NewCustomRoleResource() resource.Resource {
	return &customRoleResource{}
}

// customRoleResource defines the resource implementation.
type customRoleResource struct {
	IamClient      *iam.ClientWithResponses
	OrganizationId string
}

func (r *customRoleResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_custom_role"
}

func (r *customRoleResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Custom role resource",
		Attributes:          schemas.CustomRoleResourceSchemaAttributes(),
	}
}

func (r *customRoleResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.ResourceApiClientConfigureError(ctx, req, resp)
		return
	}

	r.IamClient = apiClients.IamClient
	r.OrganizationId = apiClients.OrganizationId
}

// ModifyPlan catches role/Direct Access Token permission conflicts at plan time.
func (r *customRoleResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state, plan models.CustomRoleResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Permissions.Equal(plan.Permissions) {
		return
	}

	if r.IamClient == nil {
		return
	}

	roleId := state.Id.ValueString()
	roleName := state.Name.ValueString()

	lookupCtx, cancel := context.WithTimeout(ctx, directAccessTokenLookupTimeout)
	defer cancel()

	conflictingTokens, err := findDirectAccessTokensUsingRole(lookupCtx, r.IamClient, r.OrganizationId, roleName)
	if err != nil {
		tflog.Warn(ctx, "failed to check for Direct Access Tokens attached to custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddWarning(
			"Unable to verify Direct Access Token compatibility",
			fmt.Sprintf(
				"Could not check whether custom role %q is referenced by any Direct Access Tokens before applying this change: %s. "+
					"If it is, this apply will fail with an error from the API instead of being caught here.",
				roleName, err,
			),
		)
		return
	}
	if len(conflictingTokens) == 0 {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("permissions"),
		"Custom role has Direct Access Tokens attached",
		fmt.Sprintf(
			"Custom role %q (id: %s) is referenced by %d Direct Access Token(s): %s. "+
				"Direct Access Tokens capture a role's permissions at creation time, so the API rejects permission "+
				"changes to a role while any Direct Access Token references it, and this apply would fail. "+
				"Delete and recreate the affected token(s) (outside of Terraform, e.g. with `astro deployment token`) "+
				"so they pick up the new permissions, then re-apply this change.",
			roleName, roleId, len(conflictingTokens), formatTokenListForDiagnostic(conflictingTokens),
		),
	)
}

func findDirectAccessTokensUsingRole(
	ctx context.Context,
	client *iam.ClientWithResponses,
	organizationId string,
	roleName string,
) ([]iam.ApiToken, error) {
	return findDirectAccessTokensUsingRoleWithLimits(
		ctx, client, organizationId, roleName,
		directAccessTokenListPageSize, directAccessTokenLookupConcurrency,
	)
}

// The list endpoint can't filter by role and omits `roles`, so tokens are listed, then
// re-fetched individually (up to concurrency at a time) to check role assignments.
func findDirectAccessTokensUsingRoleWithLimits(
	ctx context.Context,
	client *iam.ClientWithResponses,
	organizationId string,
	roleName string,
	pageSize int,
	concurrency int,
) ([]iam.ApiToken, error) {
	var candidateIds []string

	params := &iam.ListApiTokensParams{
		Kind:  lo.ToPtr(iam.DIRECTACCESS),
		Limit: lo.ToPtr(pageSize),
	}
	offset := 0
	for {
		params.Offset = &offset
		tokensResp, err := client.ListApiTokensWithResponse(ctx, organizationId, params)
		if err != nil {
			return nil, err
		}
		if tokensResp.JSON200 == nil {
			return nil, fmt.Errorf("unable to list Direct Access Tokens, got status %v", tokensResp.StatusCode())
		}

		for _, token := range tokensResp.JSON200.Tokens {
			candidateIds = append(candidateIds, token.Id)
		}

		if tokensResp.JSON200.TotalCount <= offset {
			break
		}
		offset += pageSize
	}

	var (
		mu      sync.Mutex
		matches []iam.ApiToken
	)
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)
	for _, tokenId := range candidateIds {
		g.Go(func() error {
			tokenResp, err := client.GetApiTokenWithResponse(gctx, organizationId, tokenId)
			if err != nil {
				return err
			}
			if tokenResp.JSON200 == nil {
				return fmt.Errorf("unable to get Direct Access Token %s, got status %v", tokenId, tokenResp.StatusCode())
			}

			token := tokenResp.JSON200
			if token.Roles == nil {
				return nil
			}
			for _, tokenRole := range *token.Roles {
				if tokenRole.Role == roleName {
					mu.Lock()
					matches = append(matches, *token)
					mu.Unlock()
					break
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return matches, nil
}

// formatTokenListForDiagnostic renders a length-capped, human-readable list of tokens.
func formatTokenListForDiagnostic(tokens []iam.ApiToken) string {
	shown := tokens
	truncated := 0
	if len(shown) > maxListedTokensInDiagnostic {
		shown = tokens[:maxListedTokensInDiagnostic]
		truncated = len(tokens) - maxListedTokensInDiagnostic
	}

	names := make([]string, 0, len(shown))
	for _, token := range shown {
		names = append(names, fmt.Sprintf("%s (id: %s)", token.Name, token.Id))
	}

	list := strings.Join(names, ", ")
	if truncated > 0 {
		list = fmt.Sprintf("%s, and %d more", list, truncated)
	}
	return list
}

func (r *customRoleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert permissions Set to slice
	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert restricted workspace IDs Set to slice (optional)
	var restrictedWorkspaceIds []string
	if !data.RestrictedWorkspaceIds.IsNull() && !data.RestrictedWorkspaceIds.IsUnknown() {
		resp.Diagnostics.Append(data.RestrictedWorkspaceIds.ElementsAs(ctx, &restrictedWorkspaceIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create request
	createCustomRoleRequest := iam.CreateCustomRoleRequest{
		Name:        data.Name.ValueString(),
		Permissions: permissions,
		ScopeType:   iam.CreateCustomRoleRequestScopeType(data.ScopeType.ValueString()),
	}

	// Set optional fields
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createCustomRoleRequest.Description = lo.ToPtr(data.Description.ValueString())
	}

	createCustomRoleRequest.RestrictedWorkspaceIds = &restrictedWorkspaceIds

	customRole, err := r.IamClient.CreateCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		createCustomRoleRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create custom role, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, customRole.HTTPResponse, customRole.Body, customRole.JSON200, "create custom role")
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, customRole.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a custom role resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customRoleResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get request
	customRole, err := r.IamClient.GetCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get custom role, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, customRole.HTTPResponse, customRole.Body, customRole.JSON200, "read custom role")
	// If the resource no longer exists, it is recommended to ignore the errors
	// and call RemoveResource to remove the resource from the state. The next Terraform plan will recreate the resource.
	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, customRole.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a custom role resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customRoleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert permissions Set to slice
	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert restricted workspace IDs Set to slice (optional)
	var restrictedWorkspaceIds []string
	if !data.RestrictedWorkspaceIds.IsNull() && !data.RestrictedWorkspaceIds.IsUnknown() {
		resp.Diagnostics.Append(data.RestrictedWorkspaceIds.ElementsAs(ctx, &restrictedWorkspaceIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update request
	updateCustomRoleRequest := iam.UpdateCustomRoleRequest{
		Name:        data.Name.ValueString(),
		Permissions: permissions,
	}

	// Set optional fields
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		updateCustomRoleRequest.Description = lo.ToPtr(data.Description.ValueString())
	}

	updateCustomRoleRequest.RestrictedWorkspaceIds = &restrictedWorkspaceIds

	customRole, err := r.IamClient.UpdateCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		updateCustomRoleRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update custom role, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, customRole.HTTPResponse, customRole.Body, customRole.JSON200, "update custom role")
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, customRole.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a custom role resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customRoleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete request
	customRole, err := r.IamClient.DeleteCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete custom role, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, customRole.HTTPResponse, customRole.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a custom role resource: %v", data.Id.ValueString()))
}

func (r *customRoleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
