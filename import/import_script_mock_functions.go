package import_script

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	mocksIam "github.com/astronomer/terraform-provider-astro/internal/mocks/iam"
	mocksPlatform "github.com/astronomer/terraform-provider-astro/internal/mocks/platform"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

func HandleWorkspaces(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing workspaces for organization %s", organizationId)

	workspacesResp, err := platformClient.ListWorkspacesWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list workspaces: %v", err)
	}

	if workspacesResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", workspacesResp.StatusCode(), string(workspacesResp.Body))
	}

	if workspacesResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list workspaces, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, workspacesResp.HTTPResponse, workspacesResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	workspaces := workspacesResp.JSON200.Workspaces
	if workspaces == nil {
		return "", fmt.Errorf("workspaces list is nil")
	}

	workspaceIds := lo.Map(workspaces, func(workspace platform.Workspace, _ int) string {
		return workspace.Id
	})

	log.Printf("Importing Workspaces: %v", workspaceIds)

	var importString string
	for _, workspaceId := range workspaceIds {
		workspaceImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_workspace.workspace_%v
}`, workspaceId, workspaceId)

		importString += workspaceImportString + "\n"
	}

	return importString, nil
}

func HandleDeployments(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing deployments for organization %s", organizationId)

	deploymentsResp, err := platformClient.ListDeploymentsWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list deployments: %v", err)
	}

	if deploymentsResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", deploymentsResp.StatusCode(), string(deploymentsResp.Body))
	}

	if deploymentsResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list deployments, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, deploymentsResp.HTTPResponse, deploymentsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	deployments := deploymentsResp.JSON200.Deployments
	if deployments == nil {
		return "", fmt.Errorf("deployments list is nil")
	}

	deploymentIds := lo.Map(deployments, func(deployment platform.Deployment, _ int) string {
		return deployment.Id
	})
	log.Printf("Importing Deployments: %v", deploymentIds)

	var importString string
	for _, deploymentId := range deploymentIds {
		deploymentImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_deployment.deployment_%v
}`, deploymentId, deploymentId)

		importString += deploymentImportString + "\n"
	}

	return importString, nil
}

func HandleClusters(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing clusters for organization %s", organizationId)

	clustersResp, err := platformClient.ListClustersWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list clusters: %v", err)
	}

	if clustersResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", clustersResp.StatusCode(), string(clustersResp.Body))
	}

	if clustersResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list clusters, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, clustersResp.HTTPResponse, clustersResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	clusters := clustersResp.JSON200.Clusters
	if clusters == nil {
		return "", fmt.Errorf("clusters list is nil")
	}

	clusterMap := make(map[string]platform.ClusterType)
	for _, cluster := range clusters {
		if cluster.Id != "" {
			clusterMap[cluster.Id] = cluster.Type
		}
	}

	log.Printf("Importing Clusters: %v", maps.Keys(clusterMap))

	var importString string
	for clusterId, clusterType := range clusterMap {
		clusterImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_cluster.cluster_%v
}`, clusterId, clusterId)

		if clusterType == platform.ClusterTypeHYBRID {
			log.Printf("Importing hybrid cluster workspace authorization for cluster %s", clusterId)
			clusterImportString += fmt.Sprintf(`
import {
	id = "%v"
	to = astro_hybrid_cluster_workspace_authorization.cluster_%v
}`, clusterId, clusterId)
		}

		importString += clusterImportString + "\n"
	}

	return importString, nil
}

func HandleApiTokens(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing API tokens for organization %s", organizationId)

	apiTokensResp, err := iamClient.ListApiTokensWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list API tokens: %v", err)
	}

	if apiTokensResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", apiTokensResp.StatusCode(), string(apiTokensResp.Body))
	}

	if apiTokensResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list API tokens, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, apiTokensResp.HTTPResponse, apiTokensResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	apiTokens := apiTokensResp.JSON200.Tokens
	if apiTokens == nil {
		return "", fmt.Errorf("API tokens list is nil")
	}

	apiTokenIds := lo.Map(apiTokens, func(apiToken iam.ApiToken, _ int) string {
		return apiToken.Id
	})

	log.Printf("Importing API Tokens: %v", apiTokenIds)

	var importString string
	for _, apiTokenId := range apiTokenIds {
		apiTokenImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_api_token.api_token_%v
}`, apiTokenId, apiTokenId)

		importString += apiTokenImportString + "\n"
	}

	return importString, nil
}

func HandleTeams(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing teams for organization %s", organizationId)

	teamsResp, err := iamClient.ListTeamsWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list teams: %v", err)
	}

	if teamsResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", teamsResp.StatusCode(), string(teamsResp.Body))
	}

	if teamsResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list teams, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	teams := teamsResp.JSON200.Teams
	if teams == nil {
		return "", fmt.Errorf("teams list is nil")
	}

	teamIds := lo.Map(teams, func(team iam.Team, _ int) string {
		return team.Id
	})

	log.Printf("Importing Teams: %v", teamIds)

	var importString string
	for _, teamId := range teamIds {
		teamImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_team.team_%v
}`, teamId, teamId)

		importString += teamImportString + "\n"
	}

	return importString, nil
}

func HandleTeamRoles(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing team roles for organization %s", organizationId)

	teamsResp, err := iamClient.ListTeamsWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list teams: %v", err)
	}

	if teamsResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", teamsResp.StatusCode(), string(teamsResp.Body))
	}

	if teamsResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list teams, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	teams := teamsResp.JSON200.Teams
	if teams == nil {
		return "", fmt.Errorf("teams list is nil")
	}

	teamIds := lo.Map(teams, func(team iam.Team, _ int) string {
		return team.Id
	})

	log.Printf("Importing Team Roles: %v", teamIds)

	var importString string
	for _, teamId := range teamIds {
		teamImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_team_roles.team_%v
}`, teamId, teamId)

		importString += teamImportString + "\n"
	}

	return importString, nil
}

func HandleUserRoles(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing user roles for organization %s", organizationId)

	usersResp, err := iamClient.ListUsersWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list users: %v", err)
	}

	if usersResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", usersResp.StatusCode(), string(usersResp.Body))
	}

	if usersResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list users, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, usersResp.HTTPResponse, usersResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	users := usersResp.JSON200.Users
	if users == nil {
		return "", fmt.Errorf("users list is nil")
	}

	userIds := lo.Map(users, func(user iam.User, _ int) string {
		return user.Id
	})

	log.Printf("Importing User Roles: %v", userIds)

	var importString string
	for _, userId := range userIds {
		userImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_user_roles.user_%v
}`, userId, userId)

		importString += userImportString + "\n"
	}

	return importString, nil
}
