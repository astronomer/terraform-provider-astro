package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var platformClient *platform.ClientWithResponses

const TestResourceDescription = "Created by Terraform Acceptance Test - will self-cleanup"

func GenerateTestResourceName(numRandomChars int) string {
	return fmt.Sprintf("TFAcceptanceTest_%v", strings.ToUpper(acctest.RandStringFromCharSet(numRandomChars, acctest.CharSetAlpha)))
}

func GetTestPlatformClient() (*platform.ClientWithResponses, error) {
	if platformClient != nil {
		return platformClient, nil
	}
	return platform.NewPlatformClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("ASTRO_API_TOKEN"), "acceptancetests")
}
