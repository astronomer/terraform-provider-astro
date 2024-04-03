package models

import (
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/iam"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
)

type ApiClientsModel struct {
	OrganizationId string
	PlatformClient *platform.ClientWithResponses
	IamClient      *iam.ClientWithResponses
}
