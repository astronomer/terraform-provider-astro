package common_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	mocks_platform "github.com/astronomer/terraform-provider-astro/internal/mocks/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateRoleMatchesEntityType(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		scopeType string
		want      bool
	}{
		{
			name:      "empty role",
			role:      "",
			scopeType: "organization",
			want:      false,
		},
		{
			name:      "empty scope type",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOWNER),
			scopeType: "",
			want:      false,
		},
		{
			name:      "organization owner for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOWNER),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization member for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONMEMBER),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization billing admin for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONBILLINGADMIN),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization observe admin for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization observe member for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "workspace role invalid for organization scope",
			role:      string(iam.WORKSPACEOWNER),
			scopeType: "organization",
			want:      false,
		},
		{
			name:      "deployment role invalid for organization scope",
			role:      "DEPLOYMENT_ADMIN",
			scopeType: "organization",
			want:      false,
		},
		{
			name:      "observe member invalid for workspace scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "workspace",
			want:      false,
		},
		{
			name:      "observe admin invalid for workspace scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
			scopeType: "workspace",
			want:      false,
		},
		{
			name:      "organization owner invalid for workspace scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOWNER),
			scopeType: "workspace",
			want:      false,
		},
		{
			name:      "workspace owner valid for workspace scope",
			role:      string(iam.WORKSPACEOWNER),
			scopeType: "workspace",
			want:      true,
		},
		{
			name:      "observe member invalid for deployment scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "deployment",
			want:      false,
		},
		{
			name:      "deployment admin valid for deployment scope",
			role:      "DEPLOYMENT_ADMIN",
			scopeType: "deployment",
			want:      true,
		},
		{
			name:      "scope type is case insensitive",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "ORGANIZATION",
			want:      true,
		},
		{
			name:      "unlisted role not treated as workspace or deployment role for organization scope",
			role:      "ORGANIZATION_UNKNOWN",
			scopeType: "organization",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.ValidateRoleMatchesEntityType(tt.role, tt.scopeType)
			if got != tt.want {
				t.Fatalf("ValidateRoleMatchesEntityType(%q, %q) = %v, want %v",
					tt.role, tt.scopeType, got, tt.want)
			}
		})
	}
}

func makeDeploymentsPaginatedResponse(deployments []platform.Deployment, totalCount int) *platform.ListDeploymentsResponse {
	return &platform.ListDeploymentsResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		JSON200: &platform.DeploymentsPaginated{
			Deployments: deployments,
			TotalCount:  totalCount,
		},
	}
}

func TestValidateWorkspaceDeploymentRoles(t *testing.T) {
	ctx := context.Background()
	orgId := "org-1"
	workspaceId := "ws-1"

	t.Run("no deployment roles returns nil diagnostics", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{},
			WorkspaceRoles:  []iam.WorkspaceRole{},
		})
		assert.Nil(t, diags)
		mockClient.AssertNotCalled(t, "ListDeploymentsWithResponse")
	})

	t.Run("client error is returned as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(nil, fmt.Errorf("connection refused"))

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "connection refused")
	})

	t.Run("non-200 API response is returned as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		errResp := &platform.ListDeploymentsResponse{
			HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			Body:         []byte(`{"message":"internal error"}`),
		}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(errResp, nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.True(t, diags.HasError())
	})

	t.Run("invalid deployment ID is reported as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		// API returns no deployments matching the requested ID
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(makeDeploymentsPaginatedResponse([]platform.Deployment{}, 0), nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-nonexistent", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "dep-nonexistent")
	})

	t.Run("missing workspace role for deployment's workspace is reported as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		dep := platform.Deployment{Id: "dep-1", WorkspaceId: "ws-other"}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(makeDeploymentsPaginatedResponse([]platform.Deployment{dep}, 1), nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"}},
			// workspace role is for a different workspace than the deployment belongs to
			WorkspaceRoles: []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.True(t, diags.HasError())
	})

	t.Run("valid single-page result returns no diagnostics", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		dep := platform.Deployment{Id: "dep-1", WorkspaceId: workspaceId}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(makeDeploymentsPaginatedResponse([]platform.Deployment{dep}, 1), nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.False(t, diags.HasError())
	})

	t.Run("paginated results are accumulated across multiple pages", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)

		// Page 1: returns 100 deployments, totalCount=150 signals a second page is needed
		page1Deps := make([]platform.Deployment, 100)
		for i := range page1Deps {
			page1Deps[i] = platform.Deployment{Id: fmt.Sprintf("dep-%d", i), WorkspaceId: workspaceId}
		}
		// Page 2: the remaining 50 deployments, including the one referenced in DeploymentRoles
		page2Deps := make([]platform.Deployment, 50)
		for i := range page2Deps {
			page2Deps[i] = platform.Deployment{Id: fmt.Sprintf("dep-%d", i+100), WorkspaceId: workspaceId}
		}

		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.MatchedBy(func(p *platform.ListDeploymentsParams) bool {
			return p.Offset != nil && *p.Offset == 0
		})).Return(makeDeploymentsPaginatedResponse(page1Deps, 150), nil).Once()

		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.MatchedBy(func(p *platform.ListDeploymentsParams) bool {
			return p.Offset != nil && *p.Offset == 100
		})).Return(makeDeploymentsPaginatedResponse(page2Deps, 150), nil).Once()

		// dep-149 only appears on page 2 — this validates pagination is working
		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-149", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.False(t, diags.HasError())
		mockClient.AssertNumberOfCalls(t, "ListDeploymentsWithResponse", 2)
	})

	t.Run("client error on second page is returned as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)

		page1Deps := make([]platform.Deployment, 100)
		for i := range page1Deps {
			page1Deps[i] = platform.Deployment{Id: fmt.Sprintf("dep-%d", i), WorkspaceId: workspaceId}
		}

		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.MatchedBy(func(p *platform.ListDeploymentsParams) bool {
			return p.Offset != nil && *p.Offset == 0
		})).Return(makeDeploymentsPaginatedResponse(page1Deps, 150), nil).Once()

		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.MatchedBy(func(p *platform.ListDeploymentsParams) bool {
			return p.Offset != nil && *p.Offset == 100
		})).Return(nil, fmt.Errorf("timeout on page 2")).Once()

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-149", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "timeout on page 2")
	})

	t.Run("nil JSON200 in response is returned as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		nilBodyResp := &platform.ListDeploymentsResponse{
			HTTPResponse: &http.Response{StatusCode: http.StatusOK},
			JSON200:      nil,
		}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(nilBodyResp, nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.True(t, diags.HasError())
	})

	t.Run("deployments across multiple workspaces with all workspace roles covered returns no diagnostics", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		deps := []platform.Deployment{
			{Id: "dep-1", WorkspaceId: "ws-1"},
			{Id: "dep-2", WorkspaceId: "ws-2"},
		}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(makeDeploymentsPaginatedResponse(deps, 2), nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient: mockClient,
			OrganizationId: orgId,
			Limit:          100,
			DeploymentRoles: []iam.DeploymentRole{
				{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"},
				{DeploymentId: "dep-2", Role: "DEPLOYMENT_ADMIN"},
			},
			WorkspaceRoles: []iam.WorkspaceRole{
				{WorkspaceId: "ws-1", Role: iam.WORKSPACEOWNER},
				{WorkspaceId: "ws-2", Role: iam.WORKSPACEMEMBER},
			},
		})
		assert.False(t, diags.HasError())
	})

	t.Run("deployments across multiple workspaces with only partial workspace roles is reported as diagnostic", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		deps := []platform.Deployment{
			{Id: "dep-1", WorkspaceId: "ws-1"},
			{Id: "dep-2", WorkspaceId: "ws-2"},
		}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(makeDeploymentsPaginatedResponse(deps, 2), nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient: mockClient,
			OrganizationId: orgId,
			Limit:          100,
			DeploymentRoles: []iam.DeploymentRole{
				{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"},
				{DeploymentId: "dep-2", Role: "DEPLOYMENT_ADMIN"},
			},
			// ws-2 is missing a workspace role
			WorkspaceRoles: []iam.WorkspaceRole{
				{WorkspaceId: "ws-1", Role: iam.WORKSPACEOWNER},
			},
		})
		assert.True(t, diags.HasError())
	})

	t.Run("duplicate deployment IDs are deduplicated before validation", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)
		dep := platform.Deployment{Id: "dep-1", WorkspaceId: workspaceId}
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.Anything).
			Return(makeDeploymentsPaginatedResponse([]platform.Deployment{dep}, 1), nil)

		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient: mockClient,
			OrganizationId: orgId,
			Limit:          100,
			// same deployment ID provided twice
			DeploymentRoles: []iam.DeploymentRole{
				{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"},
				{DeploymentId: "dep-1", Role: "DEPLOYMENT_ADMIN"},
			},
			WorkspaceRoles: []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.False(t, diags.HasError())
	})

	t.Run("pagination offset advances by actual page size for partial last page", func(t *testing.T) {
		mockClient := new(mocks_platform.ClientWithResponsesInterface)

		// Page 1: only 75 results (less than the 100 limit), totalCount=125
		page1Deps := make([]platform.Deployment, 75)
		for i := range page1Deps {
			page1Deps[i] = platform.Deployment{Id: fmt.Sprintf("dep-%d", i), WorkspaceId: workspaceId}
		}
		// Page 2: 50 remaining results starting at offset 75
		page2Deps := make([]platform.Deployment, 50)
		for i := range page2Deps {
			page2Deps[i] = platform.Deployment{Id: fmt.Sprintf("dep-%d", i+75), WorkspaceId: workspaceId}
		}

		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.MatchedBy(func(p *platform.ListDeploymentsParams) bool {
			return p.Offset != nil && *p.Offset == 0
		})).Return(makeDeploymentsPaginatedResponse(page1Deps, 125), nil).Once()

		// Offset must be 75 (len of page 1), not 100 (fixed limit)
		mockClient.On("ListDeploymentsWithResponse", ctx, orgId, mock.MatchedBy(func(p *platform.ListDeploymentsParams) bool {
			return p.Offset != nil && *p.Offset == 75
		})).Return(makeDeploymentsPaginatedResponse(page2Deps, 125), nil).Once()

		// dep-124 only appears on page 2
		diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
			PlatformClient:  mockClient,
			OrganizationId:  orgId,
			Limit:           100,
			DeploymentRoles: []iam.DeploymentRole{{DeploymentId: "dep-124", Role: "DEPLOYMENT_ADMIN"}},
			WorkspaceRoles:  []iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: iam.WORKSPACEOWNER}},
		})
		assert.False(t, diags.HasError())
		mockClient.AssertNumberOfCalls(t, "ListDeploymentsWithResponse", 2)
	})
}
