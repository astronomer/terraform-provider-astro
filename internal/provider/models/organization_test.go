package models_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
)

// updated_by was populated from CreatedBy.
func TestUnit_OrganizationReadFromResponse_UpdatedBy(t *testing.T) {
	var data models.Organization
	diags := data.ReadFromResponse(context.Background(), &platform.Organization{
		Id:        "org-id",
		Name:      "org",
		CreatedBy: platform.BasicSubjectProfile{Id: "creator-id"},
		UpdatedBy: platform.BasicSubjectProfile{Id: "updater-id"},
	})

	assert.False(t, diags.HasError())
	assert.Contains(t, data.CreatedBy.String(), "creator-id")
	assert.Contains(t, data.UpdatedBy.String(), "updater-id")
	assert.NotContains(t, data.UpdatedBy.String(), "creator-id")
}
