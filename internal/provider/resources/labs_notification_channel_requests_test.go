package resources

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// definitionObject builds a notification channel definition types.Object with the given overrides,
// defaulting every other attribute to null.
func definitionObject(t *testing.T, overrides map[string]attr.Value) types.Object {
	t.Helper()
	attrs := map[string]attr.Value{
		"dag_id":               types.StringNull(),
		"deployment_api_token": types.StringNull(),
		"deployment_id":        types.StringNull(),
		"recipients":           types.SetNull(types.StringType),
		"api_key":              types.StringNull(),
		"integration_key":      types.StringNull(),
		"webhook_url":          types.StringNull(),
	}
	for k, v := range overrides {
		attrs[k] = v
	}
	obj, diags := types.ObjectValue(schemas.NotificationChannelDefinitionAttributeTypes(), attrs)
	require.False(t, diags.HasError())
	return obj
}

// The labs bulk create body is a discriminated union; the marshaled JSON must carry the right type
// discriminator and the type-specific definition for each channel type.
func TestUnit_BuildLabsCreateNotificationChannelRequest(t *testing.T) {
	tests := []struct {
		name       string
		channel    models.NotificationChannelsResourceElementModel
		wantType   string
		wantDefKey string
		wantDefVal string
	}{
		{
			name: "SLACK",
			channel: models.NotificationChannelsResourceElementModel{
				Name:       types.StringValue("slack-chan"),
				Type:       types.StringValue("SLACK"),
				EntityId:   types.StringValue("clentity"),
				EntityType: types.StringValue("DEPLOYMENT"),
				IsShared:   types.BoolValue(false),
				Definition: definitionObject(t, map[string]attr.Value{"webhook_url": types.StringValue("https://hooks.slack.com/x")}),
			},
			wantType: "SLACK", wantDefKey: "webhookUrl", wantDefVal: "https://hooks.slack.com/x",
		},
		{
			name: "PAGERDUTY",
			channel: models.NotificationChannelsResourceElementModel{
				Name:       types.StringValue("pd"),
				Type:       types.StringValue("PAGERDUTY"),
				EntityId:   types.StringValue("clentity"),
				EntityType: types.StringValue("DEPLOYMENT"),
				IsShared:   types.BoolValue(false),
				Definition: definitionObject(t, map[string]attr.Value{"integration_key": types.StringValue("ik-123")}),
			},
			wantType: "PAGERDUTY", wantDefKey: "integrationKey", wantDefVal: "ik-123",
		},
		{
			name: "OPSGENIE",
			channel: models.NotificationChannelsResourceElementModel{
				Name:       types.StringValue("og"),
				Type:       types.StringValue("OPSGENIE"),
				EntityId:   types.StringValue("clentity"),
				EntityType: types.StringValue("DEPLOYMENT"),
				IsShared:   types.BoolValue(false),
				Definition: definitionObject(t, map[string]attr.Value{"api_key": types.StringValue("ak-123")}),
			},
			wantType: "OPSGENIE", wantDefKey: "apiKey", wantDefVal: "ak-123",
		},
		{
			name: "DAG_TRIGGER",
			channel: models.NotificationChannelsResourceElementModel{
				Name:       types.StringValue("dt"),
				Type:       types.StringValue("DAG_TRIGGER"),
				EntityId:   types.StringValue("clentity"),
				EntityType: types.StringValue("DEPLOYMENT"),
				IsShared:   types.BoolValue(false),
				Definition: definitionObject(t, map[string]attr.Value{
					"dag_id":               types.StringValue("my_dag"),
					"deployment_api_token": types.StringValue("tok"),
					"deployment_id":        types.StringValue("cldep"),
				}),
			},
			wantType: "DAG_TRIGGER", wantDefKey: "dagId", wantDefVal: "my_dag",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, diags := BuildLabsCreateNotificationChannelRequest(context.Background(), tc.channel)
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)

			raw, err := req.MarshalJSON()
			require.NoError(t, err)

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(raw, &decoded))

			assert.Equal(t, tc.wantType, decoded["type"], "discriminator type")
			assert.Equal(t, "clentity", decoded["entityId"])
			assert.Equal(t, "DEPLOYMENT", decoded["entityType"])

			def, ok := decoded["definition"].(map[string]any)
			require.True(t, ok, "definition should be an object")
			assert.Equal(t, tc.wantDefVal, def[tc.wantDefKey])
		})
	}
}

func TestUnit_BuildLabsCreateNotificationChannelRequest_EmailRecipients(t *testing.T) {
	channel := models.NotificationChannelsResourceElementModel{
		Name:       types.StringValue("email-chan"),
		Type:       types.StringValue("EMAIL"),
		EntityId:   types.StringValue("clentity"),
		EntityType: types.StringValue("DEPLOYMENT"),
		IsShared:   types.BoolValue(false),
		Definition: definitionObject(t, map[string]attr.Value{
			"recipients": types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a@example.com"),
				types.StringValue("b@example.com"),
			}),
		}),
	}

	req, diags := BuildLabsCreateNotificationChannelRequest(context.Background(), channel)
	require.False(t, diags.HasError())

	raw, err := req.MarshalJSON()
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	assert.Equal(t, "EMAIL", decoded["type"])
	def := decoded["definition"].(map[string]any)
	recipients := def["recipients"].([]any)
	assert.ElementsMatch(t, []any{"a@example.com", "b@example.com"}, recipients)
}

func TestUnit_BuildLabsCreateNotificationChannelRequest_UnsupportedType(t *testing.T) {
	channel := models.NotificationChannelsResourceElementModel{
		Name:       types.StringValue("x"),
		Type:       types.StringValue("CARRIER_PIGEON"),
		EntityId:   types.StringValue("clentity"),
		EntityType: types.StringValue("DEPLOYMENT"),
		IsShared:   types.BoolValue(false),
		Definition: definitionObject(t, nil),
	}
	_, diags := BuildLabsCreateNotificationChannelRequest(context.Background(), channel)
	assert.True(t, diags.HasError())
}

func TestUnit_notificationChannelElementUnchanged(t *testing.T) {
	base := models.NotificationChannelsResourceElementModel{
		Id:         types.StringValue("cmid"),
		Name:       types.StringValue("chan"),
		Type:       types.StringValue("SLACK"),
		EntityId:   types.StringValue("clentity"),
		EntityType: types.StringValue("DEPLOYMENT"),
		IsShared:   types.BoolValue(false),
		Definition: definitionObject(t, map[string]attr.Value{"webhook_url": types.StringValue("https://hooks.slack.com/x")}),
	}

	t.Run("identical elements are unchanged", func(t *testing.T) {
		other := base
		assert.True(t, notificationChannelElementUnchanged(base, other))
	})

	t.Run("differing id is ignored (still unchanged)", func(t *testing.T) {
		other := base
		other.Id = types.StringValue("different")
		assert.True(t, notificationChannelElementUnchanged(base, other))
	})

	t.Run("name change is detected", func(t *testing.T) {
		other := base
		other.Name = types.StringValue("renamed")
		assert.False(t, notificationChannelElementUnchanged(base, other))
	})

	t.Run("is_shared change is detected", func(t *testing.T) {
		other := base
		other.IsShared = types.BoolValue(true)
		assert.False(t, notificationChannelElementUnchanged(base, other))
	})

	t.Run("sensitive definition change is detected", func(t *testing.T) {
		other := base
		other.Definition = definitionObject(t, map[string]attr.Value{"webhook_url": types.StringValue("https://hooks.slack.com/CHANGED")})
		assert.False(t, notificationChannelElementUnchanged(base, other))
	})
}
