package main

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

func HandleAlerts(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing alerts for organization %s", organizationId)

	alertsResp, err := platformClient.ListAlertsWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list alerts: %v", err)
	}

	if alertsResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", alertsResp.StatusCode(), string(alertsResp.Body))
	}

	if alertsResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list alerts, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	alerts := alertsResp.JSON200.Alerts
	if alerts == nil {
		return "", fmt.Errorf("alerts list is nil")
	}

	// Define supported alert types that match the Terraform provider schema
	supportedAlertTypes := map[string]bool{
		"DAG_DURATION":   true,
		"DAG_FAILURE":    true,
		"DAG_SUCCESS":    true,
		"DAG_TIMELINESS": true,
		"TASK_FAILURE":   true,
		"TASK_DURATION":  true,
	}

	// Filter alerts to only include supported types
	var supportedAlerts []platform.Alert
	var skippedAlerts []string

	for _, alert := range alerts {
		alertType := string(alert.Type)
		if supportedAlertTypes[alertType] {
			supportedAlerts = append(supportedAlerts, alert)
		} else {
			skippedAlerts = append(skippedAlerts, fmt.Sprintf("%s (type: %s)", alert.Id, alertType))
		}
	}

	// Log information about skipped alerts
	if len(skippedAlerts) > 0 {
		log.Printf("Skipping %d alerts with unsupported types:", len(skippedAlerts))
		for _, skipped := range skippedAlerts {
			log.Printf("  - %s", skipped)
		}
	}

	alertIds := lo.Map(supportedAlerts, func(alert platform.Alert, _ int) string {
		return alert.Id
	})

	log.Printf("Importing %d supported alerts: %v", len(alertIds), alertIds)

	var importString string
	for _, alertId := range alertIds {
		alertImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_alert.alert_%v
}`, alertId, alertId)

		importString += alertImportString + "\n"
	}

	return importString, nil
}

func HandleNotificationChannels(ctx context.Context, platformClient *mocksPlatform.ClientWithResponsesInterface, iamClient *mocksIam.ClientWithResponsesInterface, organizationId string) (string, error) {
	log.Printf("Importing notification channels for organization %s", organizationId)

	notificationChannelsResp, err := platformClient.ListNotificationChannelsWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list notification channels: %v", err)
	}

	if notificationChannelsResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", notificationChannelsResp.StatusCode(), string(notificationChannelsResp.Body))
	}

	if notificationChannelsResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list notification channels, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	notificationChannels := notificationChannelsResp.JSON200.NotificationChannels
	if notificationChannels == nil {
		return "", fmt.Errorf("notification channels list is nil")
	}

	channelIds := lo.Map(notificationChannels, func(channel platform.NotificationChannel, _ int) string {
		return channel.Id
	})

	log.Printf("Importing Notification Channels: %v", channelIds)

	var importString string
	for _, channelId := range channelIds {
		channelImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_notification_channel.notification_channel_%v
}`, channelId, channelId)

		importString += channelImportString + "\n"
	}

	return importString, nil
}
