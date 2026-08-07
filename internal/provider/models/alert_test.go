package models_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
)

// A typed nil used to segfault the provider on `*v`.
func TestUnit_AlertNotificationChannelsTypesSet_TypedNil(t *testing.T) {
	set, diags := models.AlertNotificationChannelsTypesSet(context.Background(), (*[]platform.AlertNotificationChannel)(nil))
	assert.False(t, diags.HasError())
	assert.Equal(t, types.SetNull(types.ObjectType{AttrTypes: schemas.NotificationChannelsElementAttributeTypes()}), set)
}

func TestUnit_AlertNotificationChannelsTypesSet_UnexpectedType(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"untyped nil", nil},
		{"string", "not-a-slice"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := models.AlertNotificationChannelsTypesSet(context.Background(), tc.input)
			assert.True(t, diags.HasError())
		})
	}
}

func TestUnit_AlertRulesTypesObject_TypedNil(t *testing.T) {
	obj, diags := models.AlertRulesTypesObject(context.Background(), (*platform.AlertRules)(nil))
	assert.False(t, diags.HasError())
	assert.Equal(t, types.ObjectNull(schemas.AlertRulesAttributeTypes()), obj)
}

// rules is Required on the resource, so nil must error rather than produce a null.
func TestUnit_AlertRulesResourceTypesObject_Nil(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"untyped nil", nil},
		{"typed nil", (*platform.AlertRules)(nil)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := models.AlertRulesResourceTypesObject(context.Background(), tc.input)
			assert.True(t, diags.HasError())
		})
	}
}

// An alert with notificationChannels omitted must still be readable.
func TestUnit_AlertDataSourceReadFromResponse_OmittedNotificationChannels(t *testing.T) {
	var data models.AlertDataSource
	diags := data.ReadFromResponse(context.Background(), &platform.Alert{Id: "alert-id"})
	assert.False(t, diags.HasError())
	assert.Equal(t, types.StringValue("alert-id"), data.Id)
	assert.True(t, data.NotificationChannels.IsNull())
}
