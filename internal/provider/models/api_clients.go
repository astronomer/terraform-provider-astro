package models

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
)

type ApiClientsModel struct {
	OrganizationId string
	PlatformClient *platform.ClientWithResponses
	IamClient      *iam.ClientWithResponses
}
