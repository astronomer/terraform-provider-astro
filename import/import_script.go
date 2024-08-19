package import_script

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

type HandlerResult struct {
	Resource string
	Error    error
}

func main() {
	log.SetFlags(0)
	log.Println("Terraform Import Script Starting")

	// collect all arguments from the user, indicating all the resources that need to be imported
	resourcesPtr := flag.String("resources", "", "Comma separated list of resources to import. The only accepted values are workspace, deployment, cluster, api_token, team, team_roles, user_roles")
	tokenPtr := flag.String("token", "", "API token to authenticate with the platform")
	hostPtr := flag.String("host", "https://api.astronomer.io", "API host to connect to")
	organizationIdPtr := flag.String("organizationId", "", "Organization ID to import resources into")
	helpFlag := flag.Bool("help", false, "Display help information")

	flag.Parse()

	// display help information
	if *helpFlag {
		printHelp()
		return
	}

	// validate the resources argument
	resources := strings.Split(strings.ToLower(*resourcesPtr), ",")
	acceptedResources := []string{"workspace", "deployment", "cluster", "api_token", "team", "team_roles", "user_roles"}
	for _, resource := range resources {
		if !lo.Contains(acceptedResources, resource) {
			log.Fatalf("Invalid resource: %s is not accepted. The only accepted resources are %s", resource, acceptedResources)
			return
		}
	}

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

	log.Printf("Using organization ID: %s", organizationId)

	// Check if Terraform is installed
	_, err := exec.LookPath("terraform")
	if err != nil {
		log.Fatalf("Error: Terraform is not installed or not in PATH. Please install Terraform and make sure it's in your system PATH")
	}

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
	importString += fmt.Sprintf(`terraform {
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

	resourceHandlers := map[string]func(context.Context, *platform.ClientWithResponses, *iam.ClientWithResponses, string) (string, error){
		"workspace":  handleWorkspaces,
		"deployment": handleDeployments,
		"cluster":    handleClusters,
		"api_token":  handleApiTokens,
		"team":       handleTeams,
		"team_roles": handleTeamRoles,
		"user_roles": handleUserRoles,
	}

	var results []HandlerResult

	for _, resource := range resources {
		handler, exists := resourceHandlers[resource]
		if !exists {
			log.Printf("Resource not supported: %s", resource)
			results = append(results, HandlerResult{Resource: resource, Error: fmt.Errorf("resource not supported")})
			continue
		}
		result, err := handler(ctx, platformClient, iamClient, organizationId)
		if err != nil {
			log.Printf("Error handling resource %s: %v", resource, err)
			results = append(results, HandlerResult{Resource: resource, Error: err})
		} else {
			importString += result
		}
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

	// Print summary of results
	log.Println("Import process completed. Summary:")
	for _, result := range results {
		if result.Error != nil {
			log.Printf("Resource %s failed: %v", result.Resource, result.Error)
		} else {
			log.Printf("Resource %s processed successfully", result.Resource)
		}
	}
}

func printHelp() {
	log.Println("Terraform Import Script")
	log.Println("\nUsage: go run script.go [options]")
	log.Println("\nOptions:")
	log.Println("  -resources string")
	log.Println("        Comma separated list of resources to import. Accepted values:")
	log.Println("        workspace, deployment, cluster, api_token, team, team_roles, user_roles")
	log.Println("  -token string")
	log.Println("        API token to authenticate with the platform")
	log.Println("  -host string")
	log.Println("        API host to connect to (default: https://api.astronomer.io)")
	log.Println("        Use 'dev' for https://api.astronomer-dev.io")
	log.Println("        Use 'stage' for https://api.astronomer-stage.io")
	log.Println("  -organizationId string")
	log.Println("        Organization ID to import resources into")
	log.Println("  -help")
	log.Println("        Display this help information")
	log.Println("\nExample:")
	log.Println("  go run script.go -resources=workspace,deployment -token=your_api_token -organizationId=your_org_id")
	log.Println("\nNote: If the -token flag is not provided, the script will attempt to use the ASTRO_API_TOKEN environment variable.")
}

func runTerraformCommand() error {
	// delete the generated.tf file if it exists
	filenames := []string{"generated.tf", "terraform.tfstate"}
	for _, filename := range filenames {
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
	}

	// terraform plan to generate the configuration
	cmd := exec.Command("terraform", "plan", "-generate-config-out=generated.tf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func handleWorkspaces(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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

func handleDeployments(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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

func handleClusters(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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

func handleApiTokens(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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

func handleTeams(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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

func handleTeamRoles(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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

func handleUserRoles(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
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
