package main_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	import_script "github.com/astronomer/terraform-provider-astro/import"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	mocks_platform "github.com/astronomer/terraform-provider-astro/internal/mocks/platform"
	mocks_platform_v1 "github.com/astronomer/terraform-provider-astro/internal/mocks/platform_v1"
	"github.com/lucsky/cuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Import Script", func() {
	var ctx context.Context
	var mockPlatformClient *mocks_platform.ClientWithResponsesInterface
	var mockPlatformV1Client *mocks_platform_v1.ClientWithResponsesInterface
	var organizationId string

	BeforeEach(func() {
		ctx = context.Background()
		mockPlatformClient = new(mocks_platform.ClientWithResponsesInterface)
		mockPlatformV1Client = new(mocks_platform_v1.ClientWithResponsesInterface)
		organizationId = cuid.New()
	})

	Describe("HandleWorkspaces", func() {
		It("should return an error if the platform client returns an error", func() {
			mockPlatformV1Client.On("ListWorkspacesWithResponse", ctx, organizationId, (*platform_v1.ListWorkspacesParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleWorkspaces(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the platform client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListWorkspacesResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListWorkspacesWithResponse", ctx, organizationId, (*platform_v1.ListWorkspacesParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleWorkspaces(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of workspace resources", func() {
			workspaceId1 := cuid.New()
			workspaceId2 := cuid.New()

			workspaces := []platform_v1.Workspace{
				{Id: workspaceId1},
				{Id: workspaceId2},
			}

			mockResponse := &platform_v1.ListWorkspacesResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.WorkspacesPaginated{
					Workspaces: workspaces,
				},
			}

			mockPlatformV1Client.On("ListWorkspacesWithResponse", ctx, organizationId, (*platform_v1.ListWorkspacesParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleWorkspaces(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_workspace.workspace_%s", workspaceId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_workspace.workspace_%s", workspaceId2)))
		})
	})

	Describe("HandleDeployments", func() {
		It("should return an error if the platform client returns an error", func() {
			mockPlatformV1Client.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform_v1.ListDeploymentsParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the platform client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListDeploymentsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform_v1.ListDeploymentsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of deployment resources", func() {
			deploymentId1 := cuid.New()
			deploymentId2 := cuid.New()

			deployments := []platform_v1.Deployment{
				{Id: deploymentId1},
				{Id: deploymentId2},
			}

			mockResponse := &platform_v1.ListDeploymentsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.DeploymentsPaginated{
					Deployments: deployments,
				},
			}

			mockPlatformV1Client.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform_v1.ListDeploymentsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", deploymentId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", deploymentId2)))
		})

		It("should generate import statements for hybrid deployments", func() {
			deploymentId := cuid.New()
			clusterId := cuid.New()
			workspaceId := cuid.New()
			taskPodNodePoolId := cuid.New()
			deploymentType := platform_v1.DeploymentTypeHYBRID

			deployments := []platform_v1.Deployment{
				{
					Id:                 deploymentId,
					Name:               "hybrid-deployment",
					Type:               &deploymentType,
					ClusterId:          &clusterId,
					WorkspaceId:        workspaceId,
					TaskPodNodePoolId:  &taskPodNodePoolId,
					Executor:           (*platform_v1.DeploymentExecutor)(&[]string{"KUBERNETES"}[0]),
					SchedulerAu:        &[]int{5}[0],
					SchedulerReplicas:  1,
					IsCicdEnforced:     false,
					IsDagDeployEnabled: true,
				},
			}

			mockResponse := &platform_v1.ListDeploymentsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.DeploymentsPaginated{
					Deployments: deployments,
				},
			}

			mockPlatformV1Client.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform_v1.ListDeploymentsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", deploymentId)))
		})

		It("should generate import statements for hybrid deployments with worker queues", func() {
			deploymentId := cuid.New()
			clusterId := cuid.New()
			workspaceId := cuid.New()
			nodePoolId := cuid.New()
			deploymentType := platform_v1.DeploymentTypeHYBRID

			workerQueues := []platform_v1.WorkerQueue{
				{
					Name:              "default",
					IsDefault:         true,
					MaxWorkerCount:    10,
					MinWorkerCount:    1,
					WorkerConcurrency: 16,
					NodePoolId:        &nodePoolId,
				},
			}

			deployments := []platform_v1.Deployment{
				{
					Id:                 deploymentId,
					Name:               "hybrid-deployment-with-queues",
					Type:               &deploymentType,
					ClusterId:          &clusterId,
					WorkspaceId:        workspaceId,
					Executor:           (*platform_v1.DeploymentExecutor)(&[]string{"CELERY"}[0]),
					SchedulerAu:        &[]int{5}[0],
					SchedulerReplicas:  1,
					IsCicdEnforced:     false,
					IsDagDeployEnabled: true,
					WorkerQueues:       &workerQueues,
				},
			}

			mockResponse := &platform_v1.ListDeploymentsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.DeploymentsPaginated{
					Deployments: deployments,
				},
			}

			mockPlatformV1Client.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform_v1.ListDeploymentsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", deploymentId)))
		})

		It("should generate import statements for mixed deployment types", func() {
			standardDeploymentId := cuid.New()
			hybridDeploymentId := cuid.New()
			standardType := platform_v1.DeploymentTypeSTANDARD
			hybridType := platform_v1.DeploymentTypeHYBRID
			taskPodNodePoolId := cuid.New()

			deployments := []platform_v1.Deployment{
				{
					Id:                 standardDeploymentId,
					Name:               "standard-deployment",
					Type:               &standardType,
					WorkspaceId:        cuid.New(),
					SchedulerSize:      (*platform_v1.DeploymentSchedulerSize)(&[]string{"SMALL"}[0]),
					IsCicdEnforced:     false,
					IsDagDeployEnabled: true,
				},
				{
					Id:                 hybridDeploymentId,
					Name:               "hybrid-deployment",
					Type:               &hybridType,
					WorkspaceId:        cuid.New(),
					TaskPodNodePoolId:  &taskPodNodePoolId,
					SchedulerAu:        &[]int{5}[0],
					SchedulerReplicas:  1,
					IsCicdEnforced:     false,
					IsDagDeployEnabled: true,
				},
			}

			mockResponse := &platform_v1.ListDeploymentsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.DeploymentsPaginated{
					Deployments: deployments,
				},
			}

			mockPlatformV1Client.On("ListDeploymentsWithResponse", ctx, organizationId, (*platform_v1.ListDeploymentsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleDeployments(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			// Should include both deployment import statements
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", standardDeploymentId)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_deployment.deployment_%s", hybridDeploymentId)))
		})
	})

	Describe("HandleClusters", func() {
		It("should return an error if the platform client returns an error", func() {
			mockPlatformV1Client.On("ListClustersWithResponse", ctx, organizationId, (*platform_v1.ListClustersParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleClusters(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the platform client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListClustersResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListClustersWithResponse", ctx, organizationId, (*platform_v1.ListClustersParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleClusters(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of cluster resources", func() {
			clusterId1 := cuid.New()
			clusterId2 := cuid.New()
			workspaceId := cuid.New()

			clusters := []platform_v1.Cluster{
				{Id: clusterId1},
				{Id: clusterId2,
					Type:         platform_v1.ClusterTypeHYBRID,
					WorkspaceIds: &[]string{workspaceId}},
			}

			mockResponse := &platform_v1.ListClustersResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.ClustersPaginated{
					Clusters: clusters,
				},
			}

			mockPlatformV1Client.On("ListClustersWithResponse", ctx, organizationId, (*platform_v1.ListClustersParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleClusters(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_cluster.cluster_%s", clusterId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_cluster.cluster_%s", clusterId2)))

			// Test for hybrid cluster workspace authorization
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_hybrid_cluster_workspace_authorization.cluster_%s", clusterId2)))
		})
	})

	Describe("HandleApiTokens", func() {
		It("should return an error if the iam client returns an error", func() {
			mockPlatformV1Client.On("ListApiTokensWithResponse", ctx, organizationId, (*platform_v1.ListApiTokensParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleApiTokens(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the iam client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListApiTokensResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListApiTokensWithResponse", ctx, organizationId, (*platform_v1.ListApiTokensParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleApiTokens(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of api token resources", func() {
			apiTokenId1 := cuid.New()
			apiTokenId2 := cuid.New()

			apiTokens := []platform_v1.ApiToken{
				{Id: apiTokenId1},
				{Id: apiTokenId2},
			}

			mockResponse := &platform_v1.ListApiTokensResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.ApiTokensPaginated{
					Tokens: apiTokens,
				},
			}

			mockPlatformV1Client.On("ListApiTokensWithResponse", ctx, organizationId, (*platform_v1.ListApiTokensParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleApiTokens(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_api_token.api_token_%s", apiTokenId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_api_token.api_token_%s", apiTokenId2)))
		})
	})

	Describe("HandleTeams", func() {
		It("should return an error if the iam client returns an error", func() {
			mockPlatformV1Client.On("ListTeamsWithResponse", ctx, organizationId, (*platform_v1.ListTeamsParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleTeams(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the iam client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListTeamsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListTeamsWithResponse", ctx, organizationId, (*platform_v1.ListTeamsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleTeams(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of team resources", func() {
			teamId1 := cuid.New()
			teamId2 := cuid.New()

			teams := []platform_v1.Team{
				{Id: teamId1},
				{Id: teamId2},
			}

			mockResponse := &platform_v1.ListTeamsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.TeamsPaginated{
					Teams: teams,
				},
			}

			mockPlatformV1Client.On("ListTeamsWithResponse", ctx, organizationId, (*platform_v1.ListTeamsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleTeams(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team.team_%s", teamId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team.team_%s", teamId2)))
		})
	})

	Describe("HandleTeamRoles", func() {
		It("should return an error if the iam client returns an error", func() {
			mockPlatformV1Client.On("ListTeamsWithResponse", ctx, organizationId, (*platform_v1.ListTeamsParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleTeamRoles(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the iam client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListTeamsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListTeamsWithResponse", ctx, organizationId, (*platform_v1.ListTeamsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleTeamRoles(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of team role resources", func() {
			teamId1 := cuid.New()
			teamId2 := cuid.New()

			teams := []platform_v1.Team{
				{Id: teamId1},
				{Id: teamId2},
			}

			mockResponse := &platform_v1.ListTeamsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.TeamsPaginated{
					Teams: teams,
				},
			}

			mockPlatformV1Client.On("ListTeamsWithResponse", ctx, organizationId, (*platform_v1.ListTeamsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleTeamRoles(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team_roles.team_%s", teamId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_team_roles.team_%s", teamId2)))
		})
	})

	Describe("HandleUserRoles", func() {
		It("should return an error if the iam client returns an error", func() {
			mockPlatformV1Client.On("ListUsersWithResponse", ctx, organizationId, (*platform_v1.ListUsersParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleUserRoles(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the iam client returns a non-200 status code", func() {
			mockResponse := &platform_v1.ListUsersResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformV1Client.On("ListUsersWithResponse", ctx, organizationId, (*platform_v1.ListUsersParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleUserRoles(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of user resources", func() {
			userId1 := cuid.New()
			userId2 := cuid.New()

			users := []platform_v1.User{
				{Id: userId1},
				{Id: userId2},
			}

			mockResponse := &platform_v1.ListUsersResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform_v1.UsersPaginated{
					Users: users,
				},
			}

			mockPlatformV1Client.On("ListUsersWithResponse", ctx, organizationId, (*platform_v1.ListUsersParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleUserRoles(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_user_roles.user_%s", userId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_user_roles.user_%s", userId2)))
		})
	})

	Describe("HandleAlerts", func() {
		It("should return an error if the platform client returns an error", func() {
			mockPlatformClient.On("ListAlertsWithResponse", ctx, organizationId, (*platform.ListAlertsParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleAlerts(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the platform client returns a non-200 status code", func() {
			mockResponse := &platform.ListAlertsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformClient.On("ListAlertsWithResponse", ctx, organizationId, (*platform.ListAlertsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleAlerts(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of supported alert resources and skip unsupported types", func() {
			alertId1 := cuid.New()
			alertId2 := cuid.New()
			alertId3 := cuid.New()
			alertId4 := cuid.New()

			alerts := []platform.Alert{
				{Id: alertId1, Type: platform.AlertTypeDAGFAILURE},
				{Id: alertId2, Type: platform.AlertTypeDAGSUCCESS},
				{Id: alertId3, Type: platform.AlertType("JOB_SCHEDULING_DISABLED")},  // unsupported type
				{Id: alertId4, Type: platform.AlertType("WORKER_QUEUE_AT_CAPACITY")}, // unsupported type
			}

			mockResponse := &platform.ListAlertsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.AlertsPaginated{
					Alerts: alerts,
				},
			}

			mockPlatformClient.On("ListAlertsWithResponse", ctx, organizationId, (*platform.ListAlertsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleAlerts(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			// Should only include supported alert types
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_alert.alert_%s", alertId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_alert.alert_%s", alertId2)))
			// Should not include unsupported alert types
			Expect(result).ToNot(ContainSubstring(fmt.Sprintf("astro_alert.alert_%s", alertId3)))
			Expect(result).ToNot(ContainSubstring(fmt.Sprintf("astro_alert.alert_%s", alertId4)))
		})

		It("should return empty string when no supported alerts exist", func() {
			alerts := []platform.Alert{
				{Id: cuid.New(), Type: platform.AlertType("DEPRECATED_RUNTIME_VERSION")},
				{Id: cuid.New(), Type: platform.AlertType("AIRFLOW_DB_STORAGE_UNUSUALLY_HIGH")},
			}

			mockResponse := &platform.ListAlertsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.AlertsPaginated{
					Alerts: alerts,
				},
			}

			mockPlatformClient.On("ListAlertsWithResponse", ctx, organizationId, (*platform.ListAlertsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleAlerts(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should handle all supported alert types", func() {
			alertIds := make([]string, 6)
			alerts := make([]platform.Alert, 6)
			supportedTypes := []platform.AlertType{
				platform.AlertTypeDAGDURATION,
				platform.AlertTypeDAGFAILURE,
				platform.AlertTypeDAGSUCCESS,
				platform.AlertTypeDAGTIMELINESS,
				platform.AlertTypeTASKFAILURE,
				platform.AlertTypeTASKDURATION,
			}

			for i, alertType := range supportedTypes {
				alertIds[i] = cuid.New()
				alerts[i] = platform.Alert{Id: alertIds[i], Type: alertType}
			}

			mockResponse := &platform.ListAlertsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.AlertsPaginated{
					Alerts: alerts,
				},
			}

			mockPlatformClient.On("ListAlertsWithResponse", ctx, organizationId, (*platform.ListAlertsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleAlerts(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			for _, alertId := range alertIds {
				Expect(result).To(ContainSubstring(fmt.Sprintf("astro_alert.alert_%s", alertId)))
			}
		})
	})

	Describe("HandleNotificationChannels", func() {
		It("should return an error if the platform client returns an error", func() {
			mockPlatformClient.On("ListNotificationChannelsWithResponse", ctx, organizationId, (*platform.ListNotificationChannelsParams)(nil)).Return(nil, fmt.Errorf("error"))

			result, err := import_script.HandleNotificationChannels(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return an error if the platform client returns a non-200 status code", func() {
			mockResponse := &platform.ListNotificationChannelsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
			}

			mockPlatformClient.On("ListNotificationChannelsWithResponse", ctx, organizationId, (*platform.ListNotificationChannelsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleNotificationChannels(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return a list of notification channel resources", func() {
			channelId1 := cuid.New()
			channelId2 := cuid.New()

			channels := []platform.NotificationChannel{
				{Id: channelId1, Type: string(platform.AlertNotificationChannelTypeEMAIL)},
				{Id: channelId2, Type: string(platform.AlertNotificationChannelTypeSLACK)},
			}

			mockResponse := &platform.ListNotificationChannelsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.NotificationChannelsPaginated{
					NotificationChannels: channels,
				},
			}

			mockPlatformClient.On("ListNotificationChannelsWithResponse", ctx, organizationId, (*platform.ListNotificationChannelsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleNotificationChannels(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_notification_channel.notification_channel_%s", channelId1)))
			Expect(result).To(ContainSubstring(fmt.Sprintf("astro_notification_channel.notification_channel_%s", channelId2)))
		})

		It("should handle empty notification channels list", func() {
			mockResponse := &platform.ListNotificationChannelsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.NotificationChannelsPaginated{
					NotificationChannels: []platform.NotificationChannel{},
				},
			}

			mockPlatformClient.On("ListNotificationChannelsWithResponse", ctx, organizationId, (*platform.ListNotificationChannelsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleNotificationChannels(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should handle different notification channel types", func() {
			channelIds := make([]string, 5)
			channels := make([]platform.NotificationChannel, 5)
			channelTypes := []string{
				string(platform.AlertNotificationChannelTypeEMAIL),
				string(platform.AlertNotificationChannelTypeSLACK),
				string(platform.AlertNotificationChannelTypePAGERDUTY),
				string(platform.AlertNotificationChannelTypeOPSGENIE),
				string(platform.AlertNotificationChannelTypeDAGTRIGGER),
			}

			for i, channelType := range channelTypes {
				channelIds[i] = cuid.New()
				channels[i] = platform.NotificationChannel{Id: channelIds[i], Type: channelType}
			}

			mockResponse := &platform.ListNotificationChannelsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &platform.NotificationChannelsPaginated{
					NotificationChannels: channels,
				},
			}

			mockPlatformClient.On("ListNotificationChannelsWithResponse", ctx, organizationId, (*platform.ListNotificationChannelsParams)(nil)).Return(mockResponse, nil)

			result, err := import_script.HandleNotificationChannels(ctx, mockPlatformClient, mockPlatformV1Client, organizationId)

			Expect(err).To(BeNil())
			for _, channelId := range channelIds {
				Expect(result).To(ContainSubstring(fmt.Sprintf("astro_notification_channel.notification_channel_%s", channelId)))
			}
		})
	})
})

// will only work locally if organizationId and token are set
var _ = Describe("Integration Test", func() {
	var organizationId, token, rootDir, importScriptPath string

	BeforeEach(func() {
		organizationId = os.Getenv("HOSTED_ORGANIZATION_ID")
		token = os.Getenv("HOSTED_ORGANIZATION_API_TOKEN")

		// Get the current working directory
		var err error
		rootDir, err = os.Getwd()
		Expect(err).To(BeNil(), "Failed to get current working directory")

		// Find the import_script.go file
		importScriptPath = filepath.Join(rootDir, "import_script.go")
		_, err = os.Stat(importScriptPath)
		if err != nil {
			// If not found, try going up one directory
			rootDir = filepath.Dir(rootDir)
			importScriptPath = filepath.Join(rootDir, "import_script.go")
			_, err = os.Stat(importScriptPath)
		}
		Expect(err).To(BeNil(), fmt.Sprintf("import_script.go not found at %s", importScriptPath))
	})

	It("should return a list of generated resources", func() {
		envValue := os.Getenv("RUN_IMPORT_SCRIPT_TEST")

		if envValue != "true" {
			Skip(fmt.Sprintf("Skipping integration test - RUN_IMPORT_SCRIPT_TEST is '%s'", envValue))
			return
		}

		// Build the local provider binary so terraform plan uses it instead of downloading from the registry.
		// This is necessary because the CI workflow uses pull_request_target, which runs the workflow
		// definition from main — meaning any workflow-level dev_overrides setup in the PR branch is ignored.
		// By building the provider here in the test, we are fully self-sufficient.
		repoRoot := filepath.Dir(rootDir) // rootDir is import/, repoRoot is one level up
		providerDir, err := os.MkdirTemp("", "tf-provider-astro")
		Expect(err).To(BeNil(), "Failed to create temp dir for provider binary")
		defer os.RemoveAll(providerDir)

		buildCmd := exec.Command("go", "build", "-o", filepath.Join(providerDir, "terraform-provider-astro"), ".")
		buildCmd.Dir = repoRoot
		buildOutput, buildErr := buildCmd.CombinedOutput()
		Expect(buildErr).To(BeNil(), fmt.Sprintf("Failed to build provider binary: %s", string(buildOutput)))

		// Write a .terraformrc pointing dev_overrides at the temp dir containing the local binary.
		terraformrcPath := filepath.Join(providerDir, ".terraformrc")
		terraformrcContent := fmt.Sprintf("provider_installation {\n  dev_overrides {\n    \"astronomer/astro\" = %q\n  }\n  direct {}\n}\n", providerDir)
		Expect(os.WriteFile(terraformrcPath, []byte(terraformrcContent), 0644)).To(BeNil())

		// Run the import_script.go file
		cmd := exec.Command("go", "run", importScriptPath,
			"-resources", "workspace,deployment,cluster,api_token,team,team_roles,user_roles,alert,notification_channel",
			"-token", token,
			"-organizationId", organizationId,
			"-host", "dev")

		// Propagate TF_CLI_CONFIG_FILE so all terraform subprocesses inside the import script
		// use dev_overrides and skip the registry download entirely.
		cmd.Env = append(os.Environ(), fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", terraformrcPath))

		// Set the working directory to the directory containing import_script.go
		cmd.Dir = filepath.Dir(importScriptPath)

		// Capture the output of the command
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			fmt.Printf("Command output: %s\n", string(output))
			Fail(fmt.Sprintf("Command failed with error: %v", err))
		}

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("astro_workspace"))
		Expect(outputStr).To(ContainSubstring("astro_deployment"))
		Expect(outputStr).To(ContainSubstring("astro_cluster"))
		Expect(outputStr).To(ContainSubstring("astro_api_token"))
		Expect(outputStr).To(ContainSubstring("astro_team"))
		Expect(outputStr).To(ContainSubstring("astro_team_roles"))
		Expect(outputStr).To(ContainSubstring("astro_user_roles"))
		Expect(outputStr).To(ContainSubstring("astro_alert"))
		Expect(outputStr).To(ContainSubstring("astro_notification_channel"))

		fmt.Printf("Integration test has successfully passed!\n")
	})
})
