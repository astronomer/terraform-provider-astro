package models

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
)

type ApiClientsModel struct {
	OrganizationId   string
	PlatformClient   *platform.ClientWithResponses
	PlatformV1Client *platform_v1.ClientWithResponses
	IamClient        *iam.ClientWithResponses
	LabsClient       *labs.ClientWithResponses
}
