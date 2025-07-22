package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"

	"golang.org/x/exp/maps"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"

	"github.com/astronomer/terraform-provider-astro/internal/clients"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/samber/lo"
)

type HandlerResult struct {
	Resource     string
	ImportString string
	Error        error
}

func main() {
	log.SetFlags(0)
	log.Println("Terraform Import Script Starting")

	// collect all arguments from the user, indicating all the resources that need to be imported
	resourcesPtr := flag.String("resources", "workspace,deployment,cluster,api_token,team,team_roles,user_roles,alert,notification_channel", "Comma separated list of resources to import. The only accepted values are workspace, deployment, cluster, api_token, team, team_roles, user_roles, alert, notification_channel")
	tokenPtr := flag.String("token", "", "API token to authenticate with the platform")
	hostPtr := flag.String("host", "https://api.astronomer.io", "API host to connect to")
	organizationIdPtr := flag.String("organizationId", "", "Organization ID to import resources into")
	runTerraformInitPtr := flag.Bool("runTerraformInit", false, "Run terraform init after generating the import configuration")
	helpFlag := flag.Bool("help", false, "Display help information")

	flag.Parse()

	// display help information
	if *helpFlag {
		printHelp()
		return
	}

	err := checkRequiredArguments(*resourcesPtr, *tokenPtr, *organizationIdPtr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// validate the resources argument
	resources := strings.Split(strings.ToLower(*resourcesPtr), ",")
	acceptedResources := []string{"workspace", "deployment", "cluster", "api_token", "team", "team_roles", "user_roles", "alert", "notification_channel"}
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

	err = os.Setenv("ASTRO_API_TOKEN", token)
	if err != nil {
		log.Fatalf("Failed to set ASTRO_API_TOKEN environment variable: %v", err)
		return
	}

	// set the host
	var host string
	if *hostPtr == "dev" {
		host = "https://api.astronomer-dev.io"
	} else if *hostPtr == "stage" {
		host = "https://api.astronomer-stage.io"
	} else {
		host = "https://api.astronomer.io"
	}

	// set the organization ID
	organizationId := *organizationIdPtr
	if organizationId == "" {
		log.Fatalf("Organization ID not provided")
	}

	log.Printf("Using organization ID: %s", organizationId)

	// Check if Terraform is installed and the version is supported
	err = checkTerraformVersion()
	if err != nil {
		log.Fatalf("Error: %v", err)
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
			source = "astronomer/astro"
		}
	}
}

provider "astro" {
	organization_id = "%s"
	host = "%s"
}
`, organizationId, host)

	//	for each resource, we get the list of entities and generate the terraform import command

	resourceHandlers := map[string]func(context.Context, *platform.ClientWithResponses, *iam.ClientWithResponses, string) (string, error){
		"workspace":            handleWorkspaces,
		"deployment":           handleDeployments,
		"cluster":              handleClusters,
		"api_token":            handleApiTokens,
		"team":                 handleTeams,
		"team_roles":           handleTeamRoles,
		"user_roles":           handleUserRoles,
		"alert":                handleAlert,
		"notification_channel": handleNotificationChannel,
	}

	results := make(chan HandlerResult, len(resources))
	var wg sync.WaitGroup

	for _, resource := range resources {
		wg.Add(1)

		go func(resource string) {
			defer wg.Done()
			handler, exists := resourceHandlers[resource]
			if !exists {
				log.Printf("Resource not supported: %s", resource)
				results <- HandlerResult{Resource: resource, Error: fmt.Errorf("resource not supported")}
				return
			}
			result, err := handler(ctx, platformClient, iamClient, organizationId)
			if err != nil {
				log.Printf("Error handling resource %s: %v", resource, err)
				results <- HandlerResult{Resource: resource, Error: err}
			} else {
				results <- HandlerResult{Resource: resource, ImportString: result}
			}
		}(resource)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []HandlerResult
	var deploymentImportString string
	var notificationChannelImportString string
	for result := range results {
		allResults = append(allResults, result)
		if result.Error != nil {
			log.Printf("Error handling resource %s: %v", result.Resource, result.Error)
		} else {
			if result.Resource == "deployment" {
				deploymentImportString += result.ImportString
			} else if result.Resource == "notification_channel" {
				notificationChannelImportString += result.ImportString
			} else {
				importString += result.ImportString
			}
			log.Printf("Successfully handled resource %s", result.Resource)
		}
	}

	// write the terraform configuration to a file
	err = os.WriteFile("import.tf", []byte(importString), 0644)
	if err != nil {
		log.Fatalf("Failed to write import configuration to file: %v", err)
		return
	}

	log.Println("Successfully wrote import configuration to import.tf")

	// Trigger terraform init if the flag is set - used to download the provider in CI integration tests
	if *runTerraformInitPtr {
		log.Println("Running terraform init")
		rootDir, err := os.Getwd()

		// Find the import_script.go file
		importScriptPath := filepath.Join(rootDir, "import_script.go")
		_, err = os.Stat(importScriptPath)
		if err != nil {
			// If not found, try going up one directory
			rootDir = filepath.Dir(rootDir)
			importScriptPath = filepath.Join(rootDir, "import_script.go")
			_, err = os.Stat(importScriptPath)
		}

		cmd := exec.Command("terraform", "init")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to run terraform init: %v", err)
			return
		}
	}

	// Generate the corresponding terraform HCL configuration for each import block
	err = generateTerraformConfig()
	if err != nil {
		log.Fatalf("Failed to run Terraform command: %v", err)
		return
	}

	// Add deployment import blocks and HCL to the generated file
	if deploymentImportString != "" {
		err = addDeploymentsToGeneratedFile(deploymentImportString, organizationId, platformClient, ctx)
		if err != nil {
			log.Fatalf("Failed to add deployments to generated file: %v", err)
			return
		}

		log.Println("Import process completed successfully. The 'generated.tf' file now includes all resources, including deployments.")
	}

	// Add notification channel import blocks and HCL to the generated file
	if notificationChannelImportString != "" {
		err = addNotificationChannelsToGeneratedFile(notificationChannelImportString, organizationId, platformClient, ctx)
		if err != nil {
			log.Fatalf("Failed to add notification channels to generated file: %v", err)
			return
		}

		log.Println("Import process completed successfully. The 'generated.tf' file now includes all resources, including notification channels.")
	}

	// Print summary of results
	log.Println("Import process completed. Summary:")
	for _, result := range allResults {
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
	log.Println("        workspace, deployment, cluster, api_token, team, team_roles, user_roles, alert, notification_channel")
	log.Println("  -token string")
	log.Println("        API token to authenticate with the platform")
	log.Println("  -organizationId string")
	log.Println("        Organization ID to import resources into")
	log.Println("  -runTerraformInit")
	log.Println("        Run terraform init after generating the import configuration")
	log.Println("  -help")
	log.Println("        Display this help information")
	log.Println("\nExample:")
	log.Println("  go run script.go -resources=workspace,deployment -token=your_api_token -organizationId=your_org_id")
	log.Println("\nNote: If the -token flag is not provided, the script will attempt to use the ASTRO_API_TOKEN environment variable.")
}

// checkRequiredArguments checks if the required arguments are provided
func checkRequiredArguments(resourcesPtr string, tokenPtr string, organizationIdPtr string) error {
	var missingArgs []string

	if resourcesPtr == "" {
		missingArgs = append(missingArgs, "-resources (comma-separated list: workspace, deployment, cluster, api_token, team, team_roles, user_roles, alert, notification_channel)")
	}

	if tokenPtr == "" && len(os.Getenv("ASTRO_API_TOKEN")) == 0 {
		missingArgs = append(missingArgs, "-token (or ASTRO_API_TOKEN environment variable)")
	}

	if organizationIdPtr == "" {
		missingArgs = append(missingArgs, "-organizationId")
	}

	if len(missingArgs) > 0 {
		return fmt.Errorf("Missing required argument(s):\n%s", strings.Join(missingArgs, "\n"))
	}

	return nil
}

// checkTerraformVersion checks if Terraform is installed and the version is supported
func checkTerraformVersion() error {
	// Check if Terraform is installed
	_, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("Terraform is not installed or not in PATH. Please install Terraform and make sure it's in your system PATH")
	}

	// Get Terraform version
	cmd := exec.Command("terraform", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to get Terraform version: %v", err)
	}

	// Parse the version string
	versionStr := strings.TrimSpace(strings.Split(string(output), "\n")[0])
	versionStr = strings.TrimPrefix(versionStr, "Terraform v")

	// Parse the version
	currentVersion, err := version.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("Failed to parse Terraform version: %v", err)
	}

	// Define the minimum required version
	minVersion, _ := version.NewVersion("1.7.0")

	// Compare versions
	if currentVersion.LessThan(minVersion) {
		return fmt.Errorf("Terraform version %s is required. Your version (%s) is too old. Please upgrade Terraform", minVersion, currentVersion)
	}

	fmt.Printf("Terraform version %s is installed and meets the minimum required version.\n", currentVersion)
	return nil
}

// generateTerraformConfig runs terraform plan to generate the configuration
func generateTerraformConfig() error {
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

	workspacesResp, err := platformClient.ListWorkspacesWithResponse(ctx, organizationId, &platform.ListWorkspacesParams{Limit: lo.ToPtr(1000)})
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

	deploymentsResp, err := platformClient.ListDeploymentsWithResponse(ctx, organizationId, &platform.ListDeploymentsParams{Limit: lo.ToPtr(1000)})
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

	clustersResp, err := platformClient.ListClustersWithResponse(ctx, organizationId, &platform.ListClustersParams{Limit: lo.ToPtr(1000)})
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
	var clusterImportString string
	for clusterId, clusterType := range clusterMap {
		if clusterType != platform.ClusterTypeHYBRID {
			clusterImportString = fmt.Sprintf(`
import {
	id = "%v"
	to = astro_cluster.cluster_%v
}`, clusterId, clusterId)
		} else {
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

	// Check if SCIM is enabled for the organization, if so, exit as teams cannot be imported
	organizationResp, err := platformClient.GetOrganizationWithResponse(ctx, organizationId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get organization: %v", err)
	}

	if organizationResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", organizationResp.StatusCode(), string(organizationResp.Body))
	}

	if organizationResp.JSON200 == nil {
		return "", fmt.Errorf("failed to get organization, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, organizationResp.HTTPResponse, organizationResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	organization := organizationResp.JSON200
	if organization.IsScimEnabled == true {
		return "", fmt.Errorf("SCIM is enabled for the organization, teams cannot be imported")
	}

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

	_, diagnostic = clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
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

func handleAlert(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
	log.Printf("Importing alerts for organization %s", organizationId)

	alertsResp, err := platformClient.ListAlertsWithResponse(ctx, organizationId, &platform.ListAlertsParams{Limit: lo.ToPtr(1000)})
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

func handleNotificationChannel(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) (string, error) {
	log.Printf("Importing notification channels for organization %s", organizationId)

	notificationChannelsResp, err := platformClient.ListNotificationChannelsWithResponse(ctx, organizationId, &platform.ListNotificationChannelsParams{Limit: lo.ToPtr(1000)})
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

// addDeploymentsToGeneratedFile adds the deployment import blocks and HCL to the generated.tf file
func addDeploymentsToGeneratedFile(deploymentImportString string, organizationId string, platformClient *platform.ClientWithResponses, ctx context.Context) error {
	var contentBytes []byte
	var err error

	// Try to read the existing generated.tf file
	contentBytes, err = os.ReadFile("generated.tf")
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, we'll create our own content
			log.Println("generated.tf does not exist. Creating new file with deployment information.")
			contentBytes = []byte{}
		} else {
			// Some other error occurred
			return fmt.Errorf("error reading generated.tf: %v", err)
		}
	}

	// Generate deployment HCL
	deploymentHCL, err := generateDeploymentHCL(ctx, platformClient, organizationId)
	if err != nil {
		return fmt.Errorf("failed to generate deployment HCL: %v", err)
	}

	// Combine existing content (if any), deployment import blocks, and deployment HCL
	existingContent := strings.TrimSpace(string(contentBytes))
	newContent := existingContent
	if newContent != "" {
		newContent += "\n\n"
	}
	newContent += "// generated Deployment HCL \n" + strings.TrimSpace(deploymentImportString) + "\n\n" + strings.TrimSpace(deploymentHCL)

	// Write the updated content to generated.tf
	err = os.WriteFile("generated.tf", []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated generated.tf: %v", err)
	}

	log.Println("Successfully updated generated.tf with deployment information.")
	return nil
}

// addNotificationChannelsToGeneratedFile adds the notification channel import blocks and HCL to the generated.tf file
func addNotificationChannelsToGeneratedFile(notificationChannelImportString string, organizationId string, platformClient *platform.ClientWithResponses, ctx context.Context) error {
	var contentBytes []byte
	var err error

	// Try to read the existing generated.tf file
	contentBytes, err = os.ReadFile("generated.tf")
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, we'll create our own content
			log.Println("generated.tf does not exist. Creating new file with notification channel information.")
			contentBytes = []byte{}
		} else {
			// Some other error occurred
			return fmt.Errorf("error reading generated.tf: %v", err)
		}
	}

	// Generate notification channel HCL
	notificationChannelHCL, channelsWithPlaceholders, err := generateNotificationChannelHCL(ctx, platformClient, organizationId)
	if err != nil {
		return fmt.Errorf("failed to generate notification channel HCL: %v", err)
	}

	// Combine existing content (if any), notification channel import blocks, and notification channel HCL
	existingContent := strings.TrimSpace(string(contentBytes))
	newContent := existingContent
	if newContent != "" {
		newContent += "\n\n"
	}
	newContent += "// generated Notification Channel HCL \n" + strings.TrimSpace(notificationChannelImportString) + "\n\n" + strings.TrimSpace(notificationChannelHCL)

	// Write the updated content to generated.tf
	err = os.WriteFile("generated.tf", []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated generated.tf: %v", err)
	}

	log.Println("Successfully updated generated.tf with notification channel information.")

	// Print message about notification channels that need to be updated
	if len(channelsWithPlaceholders) > 0 {
		log.Println("\n" + strings.Repeat("=", 80))
		log.Println("⚠️  IMPORTANT: The following notification channels contain placeholder values")
		log.Println("   that need to be updated with actual sensitive information:")
		log.Println(strings.Repeat("=", 80))
		for _, channel := range channelsWithPlaceholders {
			log.Printf("   • %s", channel)
		}
		log.Println(strings.Repeat("=", 80))
		log.Println("   Please update these notification channels in your generated.tf file")
		log.Println("   with the actual sensitive values (webhook URLs, API keys, tokens, etc.)")
		log.Println("   before running 'terraform apply'.")
		log.Println(strings.Repeat("=", 80) + "\n")
	}

	return nil
}

// generateDeploymentHCL generates the HCL for all deployments in the organization
// generateTerraformConfig has trouble with deployments, so we generate the HCL manually
func generateDeploymentHCL(ctx context.Context, platformClient *platform.ClientWithResponses, organizationId string) (string, error) {
	deploymentsResp, err := platformClient.ListDeploymentsWithResponse(ctx, organizationId, &platform.ListDeploymentsParams{Limit: lo.ToPtr(1000)})
	if err != nil {
		return "", fmt.Errorf("failed to list deployments: %v", err)
	}

	if deploymentsResp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", deploymentsResp.StatusCode(), string(deploymentsResp.Body))
	}

	if deploymentsResp.JSON200 == nil {
		return "", fmt.Errorf("failed to list deployments, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	deployments := deploymentsResp.JSON200.Deployments
	if deployments == nil {
		return "", fmt.Errorf("deployments list is nil")
	}

	deploymentIds := lo.Map(deployments, func(deployment platform.Deployment, _ int) string {
		return deployment.Id
	})

	var hclString string
	for _, deploymentId := range deploymentIds {
		var deploymentHCL string

		// get deployment details
		deploymentResp, err := platformClient.GetDeploymentWithResponse(ctx, organizationId, deploymentId)
		if err != nil {
			return "", fmt.Errorf("failed to list deployments: %v", err)
		}

		if deploymentsResp.StatusCode() != http.StatusOK {
			return "", fmt.Errorf("unexpected status code: %d, body: %s", deploymentsResp.StatusCode(), string(deploymentsResp.Body))
		}

		if deploymentsResp.JSON200 == nil {
			return "", fmt.Errorf("failed to list deployments, JSON200 resp is nil, organizationId: %v", organizationId)
		}

		deployment := deploymentResp.JSON200
		if deployment == nil {
			return "", fmt.Errorf("deployment is nil")
		}

		contactEmailsString := formatContactEmails(deployment.ContactEmails)
		environmentVariablesString := formatEnvironmentVariables(deployment.EnvironmentVariables)
		workerQueuesString := formatWorkerQueues(deployment.WorkerQueues, (*string)(deployment.Executor))

		deploymentType := deployment.Type

		workloadIdentity := deployment.WorkloadIdentity
		workloadIdentityString := ""
		if workloadIdentity != nil {
			workloadIdentityString = fmt.Sprintf(`desired_workload_identity = "%s"`, *workloadIdentity)
		}

		if *deploymentType == platform.DeploymentTypeDEDICATED {
			deploymentHCL = fmt.Sprintf(`
resource "astro_deployment" "deployment_%s" {
	cluster_id = "%s"
	%s
	default_task_pod_cpu = "%s"
	default_task_pod_memory = "%s"
	description = "%s"
	%s
	executor = "%s"
	is_cicd_enforced = %t
	is_dag_deploy_enabled = %t
	is_development_mode = %t
	is_high_availability = %t
	name = "%s"
	resource_quota_cpu = "%s"
	resource_quota_memory = "%s"
	scheduler_size = "%s"
	type = "%s"
	workspace_id = "%s"
	%s
    %s
}
`,
				deployment.Id,
				stringValue(deployment.ClusterId),
				contactEmailsString,
				stringValue(deployment.DefaultTaskPodCpu),
				stringValue(deployment.DefaultTaskPodMemory),
				stringValue(deployment.Description),
				environmentVariablesString,
				stringValue((*string)(deployment.Executor)),
				deployment.IsCicdEnforced,
				deployment.IsDagDeployEnabled,
				boolValue(deployment.IsDevelopmentMode),
				boolValue(deployment.IsHighAvailability),
				deployment.Name,
				stringValue(deployment.ResourceQuotaCpu),
				stringValue(deployment.ResourceQuotaMemory),
				stringValue((*string)(deployment.SchedulerSize)),
				stringValue((*string)(deploymentType)),
				deployment.WorkspaceId,
				workerQueuesString,
				workloadIdentityString,
			)
		} else if *deploymentType == platform.DeploymentTypeSTANDARD {
			deploymentHCL = fmt.Sprintf(`
resource "astro_deployment" "deployment_%s" {
	cloud_provider = "%s"
	%s
	default_task_pod_cpu = "%s"
	default_task_pod_memory = "%s"
	description = "%s"
	%s
	executor = "%s"
	is_cicd_enforced = %t
	is_dag_deploy_enabled = %t
	is_development_mode = %t
	is_high_availability = %t
	name = "%s"
	region = "%s"
	resource_quota_cpu = "%s"
	resource_quota_memory = "%s"
	scheduler_size = "%s"
	type = "%s"
	workspace_id = "%s"
	%s
    %s
}
`,
				deployment.Id,
				stringValue((*string)(deployment.CloudProvider)),
				contactEmailsString,
				stringValue(deployment.DefaultTaskPodCpu),
				stringValue(deployment.DefaultTaskPodMemory),
				stringValue(deployment.Description),
				environmentVariablesString,
				stringValue((*string)(deployment.Executor)),
				deployment.IsCicdEnforced,
				deployment.IsDagDeployEnabled,
				boolValue(deployment.IsDevelopmentMode),
				boolValue(deployment.IsHighAvailability),
				deployment.Name,
				stringValue(deployment.Region),
				stringValue(deployment.ResourceQuotaCpu),
				stringValue(deployment.ResourceQuotaMemory),
				stringValue((*string)(deployment.SchedulerSize)),
				stringValue((*string)(deploymentType)),
				deployment.WorkspaceId,
				workerQueuesString,
				workloadIdentityString,
			)
		} else {
			log.Printf("Skipping deployment %s: unsupported deployment type %s", deployment.Id, stringValue((*string)(deploymentType)))
		}
		log.Printf("Generated import for astro_deployment.deployment_%s", deployment.Id)

		hclString += deploymentHCL
	}

	return hclString, nil
}

// generateNotificationChannelHCL generates the HCL for all notification channels in the organization
func generateNotificationChannelHCL(ctx context.Context, platformClient *platform.ClientWithResponses, organizationId string) (string, []string, error) {
	notificationChannelsResp, err := platformClient.ListNotificationChannelsWithResponse(ctx, organizationId, &platform.ListNotificationChannelsParams{Limit: lo.ToPtr(1000)})
	if err != nil {
		return "", nil, fmt.Errorf("failed to list notification channels: %v", err)
	}

	if notificationChannelsResp.StatusCode() != http.StatusOK {
		return "", nil, fmt.Errorf("unexpected status code: %d, body: %s", notificationChannelsResp.StatusCode(), string(notificationChannelsResp.Body))
	}

	if notificationChannelsResp.JSON200 == nil {
		return "", nil, fmt.Errorf("failed to list notification channels, JSON200 resp is nil, organizationId: %v", organizationId)
	}

	_, diagnostic := clients.NormalizeAPIError(ctx, notificationChannelsResp.HTTPResponse, notificationChannelsResp.Body)
	if diagnostic != nil {
		log.Printf("API Error diagnostic: %+v", diagnostic)
	}

	notificationChannels := notificationChannelsResp.JSON200.NotificationChannels
	if notificationChannels == nil {
		return "", nil, fmt.Errorf("notification channels list is nil")
	}

	var hclString string
	var channelsWithPlaceholders []string

	for _, channel := range notificationChannels {
		var channelHCL string
		hasPlaceholder := false

		channelType := channel.Type
		channelName := channel.Name
		entityId := channel.EntityId
		entityType := channel.EntityType
		isShared := channel.IsShared

		// Generate appropriate definition based on channel type
		var definition string
		switch channelType {
		case string(platform.AlertNotificationChannelTypeEMAIL):
			if defMap, ok := channel.Definition.(map[string]interface{}); ok {
				if recipientsArray, ok := defMap["recipients"].([]interface{}); ok {
					var recipients []string
					for _, r := range recipientsArray {
						if email, ok := r.(string); ok {
							recipients = append(recipients, fmt.Sprintf(`"%s"`, email))
						}
					}
					recipientsString := strings.Join(recipients, ", ")
					definition = fmt.Sprintf(`definition = {
		recipients = [%s]
	}`, recipientsString)
				} else {
					definition = `definition = {
		recipients = ["PLACEHOLDER_EMAIL"] # Replace with actual email addresses
	}`
					hasPlaceholder = true
				}
			} else {
				definition = `definition = {
		recipients = ["PLACEHOLDER_EMAIL"] # Replace with actual email addresses
	}`
				hasPlaceholder = true
			}
		case string(platform.AlertNotificationChannelTypeSLACK):
			definition = `definition = {
		webhook_url = "PLACEHOLDER_WEBHOOK_URL" # Replace with actual webhook URL
	}`
			hasPlaceholder = true
		case string(platform.AlertNotificationChannelTypePAGERDUTY):
			definition = `definition = {
		integration_key = "PLACEHOLDER_INTEGRATION_KEY" # Replace with actual integration key
	}`
			hasPlaceholder = true
		case string(platform.AlertNotificationChannelTypeOPSGENIE):
			definition = `definition = {
		api_key = "PLACEHOLDER_API_KEY" # Replace with actual API key
	}`
			hasPlaceholder = true
		case string(platform.AlertNotificationChannelTypeDAGTRIGGER):
			// Type assert the definition to access its fields
			if defMap, ok := channel.Definition.(map[string]interface{}); ok {
				dagId, _ := defMap["dagId"].(string)
				deploymentId, _ := defMap["deploymentId"].(string)

				definition = fmt.Sprintf(`definition = {
		dag_id = "%s"
		deployment_api_token = "PLACEHOLDER_API_TOKEN" # Replace with actual deployment API token
		deployment_id = "%s"
	}`, dagId, deploymentId)
				hasPlaceholder = true // DAG_TRIGGER always has placeholder for deployment_api_token
			} else {
				definition = `definition = {
		dag_id = "PLACEHOLDER_DAG_ID"
		deployment_api_token = "PLACEHOLDER_API_TOKEN"
		deployment_id = "PLACEHOLDER_DEPLOYMENT_ID"
	}`
				hasPlaceholder = true
			}

		default:
			log.Printf("Skipping notification channel %s: unsupported type %s", channel.Id, channelType)
			continue
		}

		// Track channels that have placeholder values
		if hasPlaceholder {
			channelsWithPlaceholders = append(channelsWithPlaceholders, fmt.Sprintf("%s (%s - %s)", channel.Id, channelName, channelType))
		}

		channelHCL = fmt.Sprintf(`
resource "astro_notification_channel" "notification_channel_%s" {
	name = "%s"
	type = "%s"
	entity_id = "%s"
	entity_type = "%s"
	is_shared = %t
	%s
}
`,
			channel.Id,
			channelName,
			channelType,
			entityId,
			entityType,
			isShared,
			definition,
		)

		log.Printf("Generated import for astro_notification_channel.notification_channel_%s", channel.Id)
		hclString += channelHCL
	}

	return hclString, channelsWithPlaceholders, nil
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func formatContactEmails(emails *[]string) string {
	if emails == nil || len(*emails) == 0 {
		return fmt.Sprintf(`contact_emails = []`)
	}
	quotedEmails := make([]string, len(*emails))
	for i, email := range *emails {
		quotedEmails[i] = fmt.Sprintf(`"%s"`, email)
	}
	return fmt.Sprintf(`contact_emails = [%s]`, strings.Join(quotedEmails, ", "))
}

func formatEnvironmentVariables(envVars *[]platform.DeploymentEnvironmentVariable) string {
	if envVars == nil || len(*envVars) == 0 {
		return fmt.Sprintf(`environment_variables = []`)
	}
	variables := lo.Map(*envVars, func(envVar platform.DeploymentEnvironmentVariable, _ int) string {
		value := fmt.Sprintf(`"%s"`, stringValue(envVar.Value))

		if envVar.IsSecret {
			value = "null"
		}

		return fmt.Sprintf(`{
		key = "%s"
		value = %s
		is_secret = %t
	}`, envVar.Key, value, envVar.IsSecret)
	})
	return fmt.Sprintf(`environment_variables = [%s]`, strings.Join(variables, ", "))
}

func formatWorkerQueues(queues *[]platform.WorkerQueue, executor *string) string {
	// If queues is nil and executor is not CELERY, return an empty string
	if queues == nil && (executor == nil || *executor != "CELERY") {
		return ""
	}

	// If queues is empty but executor is CELERY, return an empty worker_queues array
	if (queues == nil || len(*queues) == 0) && executor != nil && *executor == "CELERY" {
		return `worker_queues = []`
	}

	// If we have queues, format them
	if queues != nil && len(*queues) > 0 {
		workerQueues := lo.Map(*queues, func(queue platform.WorkerQueue, _ int) string {
			return fmt.Sprintf(`{
		astro_machine = "%s"
		name = "%s"
		is_default = %t
		max_worker_count = %d
		min_worker_count = %d
		worker_concurrency = %d
	}`, stringValue(queue.AstroMachine), queue.Name, queue.IsDefault, queue.MaxWorkerCount, queue.MinWorkerCount, queue.WorkerConcurrency)
		})
		return fmt.Sprintf(`worker_queues = [%s]`, strings.Join(workerQueues, ", "))
	}

	// If we've reached here, it means queues is nil or empty, and executor is not CELERY
	return ""
}
