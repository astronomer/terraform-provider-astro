package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"

	"github.com/astronomer/terraform-provider-astro/internal/clients"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/samber/lo"
)

func main() {
	log.Println("Terraform Import Script Starting")

	// collect all arguments from the user, indicating all the resources that need to be imported
	resourcesPtr := flag.String("resources", "", "Comma separated list of resources to import")
	tokenPtr := flag.String("token", "", "API token to authenticate with the platform")
	hostPtr := flag.String("host", "https://api.astronomer.io", "API host to connect to")
	organizationIdPtr := flag.String("organizationId", "", "Organization ID to import resources into")

	flag.Parse()

	*resourcesPtr = strings.ToLower(*resourcesPtr)
	resources := strings.Split(*resourcesPtr, ",")
	log.Println("Resources to import: ", resources)

	// set the API token
	token := *tokenPtr
	if token == "" {
		token = os.Getenv("ASTRO_API_TOKEN")
	}

	if len(token) == 0 {
		log.Fatal("API token not provided")
		return
	}

	// set the host
	var host string
	if *hostPtr == "dev" {
		host = "https://api.astronomer-dev.io"
	} else if *hostPtr == "stage" {
		host = "https://api.astronomer-stage.io"
	} else {
		host = *hostPtr
	}

	// set the organization ID
	organizationId := *organizationIdPtr
	if organizationId == "" {
		log.Fatalf("Organization ID not provided")
	}

	log.Printf("Using API host: %s", host)
	log.Printf("Using API token: %s", token)
	log.Printf("Using organization ID: %s", organizationId)

	// connect to v1beta1 client
	ctx := context.Background()
	platformClient, err := platform.NewPlatformClient(host, token, "import")
	if err != nil {
		log.Fatalf("Failed to create platform client: %v", err)
	}

	iamClient, err := iam.NewIamClient(host, token, "import")
	if err != nil {
		log.Fatalf("Failed to create iam client: %v", err)
		return
	}

	// set terraform provider configuration
	var importString string
	importString += fmt.Sprintf(`
	terraform {
		required_providers {
			astro = {
				source = "registry.terraform.io/astronomer/astro"
			}
		}
	}
	
	provider "astro" {
		organization_id = "%s"
		host = "%s"
		token = "%s"
	}

	`, organizationId, host, token)

	//	for each resource, we get the list of entities and generate the terraform import command

	resourceHandlers := map[string]func(context.Context, *platform.ClientWithResponses, *iam.ClientWithResponses, string) string{
		"workspace":  handleWorkspaces,
		"deployment": handleDeployments,
		"cluster":    handleClusters,
		"api_token":  handleApiTokens,
		"team":       handleTeams,
		"team_roles": handleTeamRoles,
		"user_roles": handleUserRoles,
	}

	for _, resource := range resources {
		handler, exists := resourceHandlers[resource]
		if !exists {
			log.Println("Resource not supported: ", resource)
			continue
		}
		importString += handler(ctx, platformClient, iamClient, organizationId)
	}

	// write the terraform configuration to a file
	err = os.WriteFile("import.tf", []byte(importString), 0644)
	if err != nil {
		log.Fatalf("Failed to write import configuration to file: %v", err)
		return
	}

	log.Println("Successfully wrote import configuration to import.tf")

	// Generate the corresponding terraform HCL configuration for each import block
	err = runTerraformCommand()
	if err != nil {
		log.Fatalf("Failed to run Terraform command: %v", err)
		return
	}

	log.Println("Terraform plan executed successfully")
}

func runTerraformCommand() error {
	// delete the generated.tf file if it exists
	filename := "generated.tf"
	// Check if the file exists
	if _, err := os.Stat(filename); err == nil {
		// File exists, so delete it
		err = os.Remove(filename)
		if err != nil {
			return err
		}
		log.Printf("Successfully deleted %s", filename)
	} else if os.IsNotExist(err) {
		// File does not exist, nothing to do
		log.Printf("%s does not exist, no need to delete", filename)
	} else {
		// Some other error occurred
		return err
	}

	// terraform plan to generate the configuration
	cmd := exec.Command("terraform", "plan", "-generate-config-out=generated.tf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// terraform apply to import the resources
	cmd = exec.Command("terraform", "apply", "-auto-approve")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func handleWorkspaces(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing workspaces for organization %s", organizationId)

	workspacesResp, err := platformClient.ListWorkspacesWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list workspaces: %v", err)
		return ""
	}

	// Check HTTP status code
	if workspacesResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", workspacesResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(workspacesResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if workspacesResp.JSON200 == nil {
		log.Printf("Failed to list workspaces, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(workspacesResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, workspacesResp.HTTPResponse, workspacesResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	workspaces := workspacesResp.JSON200.Workspaces
	if workspaces == nil {
		log.Printf("Workspaces list is nil")
		return ""
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

	return importString
}

func handleDeployments(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing deployments for organization %s", organizationId)

	deploymentsResp, err := platformClient.ListDeploymentsWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list deployments: %v", err)
		return ""
	}

	// Check HTTP status code
	if deploymentsResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", deploymentsResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(deploymentsResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if deploymentsResp.JSON200 == nil {
		log.Printf("Failed to list deployments, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(deploymentsResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, deploymentsResp.HTTPResponse, deploymentsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	deployments := deploymentsResp.JSON200.Deployments
	if deployments == nil {
		log.Printf("Deployments list is nil")
		return ""
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

	return importString
}

func handleClusters(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing clusters for organization %s", organizationId)

	clustersResp, err := platformClient.ListClustersWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list clusters: %v", err)
		return ""
	}

	// Check HTTP status code
	if clustersResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", clustersResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(clustersResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if clustersResp.JSON200 == nil {
		log.Printf("Failed to list clusters, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(clustersResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, clustersResp.HTTPResponse, clustersResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	clusters := clustersResp.JSON200.Clusters
	if clusters == nil {
		log.Printf("Clusters list is nil")
		return ""
	}

	var clusterMap map[string]platform.ClusterType
	for _, cluster := range clusters {
		clusterMap[cluster.Id] = cluster.Type
	}

	log.Printf("Importing Clusters: %v", maps.Keys(clusterMap))

	var importString string
	for clusterId, clusterType := range clusterMap {
		clusterImportString := fmt.Sprintf(`
import {
	id = "%v"
	to = astro_cluster.cluster_%v
}`, clusterId, clusterId)

		// if clusterType is hybrid, we need to import the hybrid cluster workspace authorization
		if clusterType == platform.ClusterTypeHYBRID {
			clusterImportString += fmt.Sprintf(`
import {
	id = "%v"
	to = astro_hybrid_cluster_workspace_authorization.cluster_%v
}`, clusterId, clusterId)
		}

		importString += clusterImportString + "\n"
	}

	return importString
}

func handleApiTokens(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing API tokens for organization %s", organizationId)

	apiTokensResp, err := iamClient.ListApiTokensWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list API tokens: %v", err)
		return ""
	}

	// Check HTTP status code
	if apiTokensResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", apiTokensResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(apiTokensResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if apiTokensResp.JSON200 == nil {
		log.Printf("Failed to list API tokens, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(apiTokensResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, apiTokensResp.HTTPResponse, apiTokensResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	apiTokens := apiTokensResp.JSON200.Tokens
	if apiTokens == nil {
		log.Printf("API tokens list is nil")
		return ""
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

	return importString
}

func handleTeams(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing teams for organization %s", organizationId)

	teamsResp, err := iamClient.ListTeamsWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list teams: %v", err)
		return ""
	}

	// Check HTTP status code
	if teamsResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", teamsResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(teamsResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if teamsResp.JSON200 == nil {
		log.Printf("Failed to list teams, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(teamsResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	teams := teamsResp.JSON200.Teams
	if teams == nil {
		log.Printf("Teams list is nil")
		return ""
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

	return importString
}

func handleTeamRoles(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing team roles for organization %s", organizationId)

	teamsResp, err := iamClient.ListTeamsWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list teams: %v", err)
		return ""
	}

	// Check HTTP status code
	if teamsResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", teamsResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(teamsResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if teamsResp.JSON200 == nil {
		log.Printf("Failed to list teams, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(teamsResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	teams := teamsResp.JSON200.Teams
	if teams == nil {
		log.Printf("Teams list is nil")
		return ""
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

	return importString
}

func handleUserRoles(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	log.Printf("Importing user roles for organization %s", organizationId)

	usersResp, err := iamClient.ListUsersWithResponse(ctx, organizationId, nil)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return ""
	}

	// Check HTTP status code
	if usersResp.StatusCode() != http.StatusOK {
		log.Printf("Unexpected status code: %d", usersResp.StatusCode())

		// Try to read and log the response body
		bodyString := string(usersResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	if usersResp.JSON200 == nil {
		log.Printf("Failed to list users, JSON200 resp is nil, organizationId: %v", organizationId)
		// Try to read and log the response body
		bodyString := string(usersResp.Body)
		log.Printf("Response body: %s", bodyString)
		return ""
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, usersResp.HTTPResponse, usersResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	users := usersResp.JSON200.Users
	if users == nil {
		log.Printf("Users list is nil")
		return ""
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

	return importString
}
