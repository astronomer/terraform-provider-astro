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
		"workspace":                              handleWorkspaces,
		"deployment":                             handleDeployments,
		"cluster":                                handleClusters,
		"hybrid_cluster_workspace_authorization": handleHybridClusterWorkspaceAuthorizations,
		"api_token":                              handleApiTokens,
		"team":                                   handleTeams,
		"team_roles":                             handleTeamRoles,
		"user_roles":                             handleUserRoles,
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

	return ""
}

func handleClusters(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {

	return ""
}

func handleHybridClusterWorkspaceAuthorizations(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {

	return ""
}

func handleApiTokens(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	return ""
}

func handleTeams(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	return ""
}

func handleTeamRoles(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	return ""
}

func handleUserRoles(ctx context.Context, platformClient *platform.ClientWithResponses, iamClient *iam.ClientWithResponses, organizationId string) string {
	return ""
}
