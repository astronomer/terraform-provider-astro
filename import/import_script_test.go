package import_script_test

import (
	"context"
	"fmt"
	"net/http"

	import_script "github.com/astronomer/terraform-provider-astro/import"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	mocks_iam "github.com/astronomer/terraform-provider-astro/internal/mocks/iam"
	mocks_platform "github.com/astronomer/terraform-provider-astro/internal/mocks/platform"
	"github.com/lucsky/cuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Import Script", func() {
	var ctx context.Context
	var mockPlatformClient *mocks_platform.ClientWithResponsesInterface
	var mockIAMClient *mocks_iam.ClientWithResponsesInterface
	var organizationId string

	BeforeEach(func() {
		ctx = context.Background()
		mockPlatformClient = new(mocks_platform.ClientWithResponsesInterface)
		mockIAMClient = new(mocks_iam.ClientWithResponsesInterface)
		organizationId = cuid.New()
	})

	Describe("HandleWorkspaces", func() {
		It("should return a list of workspace resources", func() {
			workspaceId1 := cuid.New()
			workspaceId2 := cuid.New()

			workspaces := []platform.Workspace{
				{Id: workspaceId1},
				{Id: workspaceId2},
			}

			mockResponse := &platform.ListWorkspacesResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.WorkspacesPaginated{
					Workspaces: workspaces,
				},
			}

			mockPlatformClient.On("ListWorkspacesWithResponse", ctx, organizationId, (*platform.ListWorkspacesParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleWorkspaces(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_workspace.workspace_%s", workspaceId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_workspace.workspace_%s", workspaceId2)))
		})
	})

	Describe("HandleDeployments", func() {
		It("should return a list of deployment resources", func() {
			deploymentId1 := cuid.New()
			deploymentId2 := cuid.New()

			deployments := []platform.Deployment{
				{Id: deploymentId1},
				{Id: deploymentId2},
			}

			mockResponse := &platform.ListDeploymentsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.DeploymentsPaginated{
					Deployments: deployments,
				},
			}

			mockPlatformClient.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform.ListDeploymentsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", deploymentId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", deploymentId2)))
		})
	})

	Describe("HandleClusters", func() {
		It("should return a list of cluster resources", func() {
			clusterId1 := cuid.New()
			clusterId2 := cuid.New()
			workspaceId := cuid.New()

			clusters := []platform.Cluster{
				{Id: clusterId1},
				{Id: clusterId2,
					Type:         platform.ClusterTypeHYBRID,
					WorkspaceIds: &[]string{workspaceId}},
			}

			mockResponse := &platform.ListClustersResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.ClustersPaginated{
					Clusters: clusters,
				},
			}

			mockPlatformClient.On("ListClustersWithResponse", ctx, organizationId, (*platform.ListClustersParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleClusters(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_cluster.cluster_%s", clusterId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_cluster.cluster_%s", clusterId2)))

			// Test for hybrid cluster workspace authorization
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_hybrid_cluster_workspace_authorization.cluster_%s", clusterId2)))
		})
	})

	Describe("HandleApiTokens", func() {
		It("should return a list of api token resources", func() {
			apiTokenId1 := cuid.New()
			apiTokenId2 := cuid.New()

			apiTokens := []iam.ApiToken{
				{Id: apiTokenId1},
				{Id: apiTokenId2},
			}

			mockResponse := &iam.ListApiTokensResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &iam.ApiTokensPaginated{
					Tokens: apiTokens,
				},
			}

			mockIAMClient.On("ListApiTokensWithResponse", ctx, organizationId, (*iam.ListApiTokensParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleApiTokens(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_api_token.api_token_%s", apiTokenId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_api_token.api_token_%s", apiTokenId2)))
		})
	})

	Describe("HandleTeams", func() {
		It("should return a list of team resources", func() {
			teamId1 := cuid.New()
			teamId2 := cuid.New()

			teams := []iam.Team{
				{Id: teamId1},
				{Id: teamId2},
			}

			mockResponse := &iam.ListTeamsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &iam.TeamsPaginated{
					Teams: teams,
				},
			}

			mockIAMClient.On("ListTeamsWithResponse", ctx, organizationId, (*iam.ListTeamsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleTeams(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team.team_%s", teamId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team.team_%s", teamId2)))
		})
	})

	Describe("HandleTeamRoles", func() {
		It("should return a list of team role resources", func() {
			teamId1 := cuid.New()
			teamId2 := cuid.New()

			teams := []iam.Team{
				{Id: teamId1},
				{Id: teamId2},
			}

			mockResponse := &iam.ListTeamsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &iam.TeamsPaginated{
					Teams: teams,
				},
			}

			mockIAMClient.On("ListTeamsWithResponse", ctx, organizationId, (*iam.ListTeamsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleTeamRoles(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team_roles.team_%s", teamId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team_roles.team_%s", teamId2)))
		})
	})

	Describe("HandleUsers", func() {
		It("should return a list of user resources", func() {
			userId1 := cuid.New()
			userId2 := cuid.New()

			users := []iam.User{
				{Id: userId1},
				{Id: userId2},
			}

			mockResponse := &iam.ListUsersResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &iam.UsersPaginated{
					Users: users,
				},
			}

			mockIAMClient.On("ListUsersWithResponse", ctx, organizationId, (*iam.ListUsersParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleUserRoles(ctx, mockPlatformClient, mockIAMClient, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_user_roles.user_%s", userId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_user_roles.user_%s", userId2)))
		})
	})
})
