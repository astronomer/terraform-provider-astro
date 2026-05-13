package models_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
)

// TestUnit_NotificationChannelDefinitionDataSourceTypesObject is the regression guard for CPP-737.
// Before the fix, a notification channel definition coming back from the API with camelCase keys
// (dagId, deploymentId, etc.) caused types.ObjectValue to reject the extra attributes
// ("Extra Object Attribute Name: dagId"), breaking any astro_alert that referenced such a channel.
//
// The cases cover every channel type the OpenAPI spec declares (per platform.api.gen.go):
//   - DAG_TRIGGER: dagId, deploymentApiToken, deploymentId
//   - EMAIL: recipients
//   - OPSGENIE: apiKey
//   - PAGERDUTY: integrationKey
//   - SLACK: webhookUrl
func TestUnit_NotificationChannelDefinitionDataSourceTypesObject(t *testing.T) {
	tests := []struct {
		name      string
		def       interface{}
		expectErr bool
		// expected snake_case attribute values; empty string means expect-null
		wantStrings    map[string]string
		wantRecipients []string // nil means expect-null
	}{
		{
			name: "DAG_TRIGGER with dagId, deploymentId, and deploymentApiToken",
			def: map[string]interface{}{
				"dagId":              "test_dag",
				"deploymentId":       "clxxxxxxxxxxxxxxxxxxxx",
				"deploymentApiToken": "tok_secret",
			},
			wantStrings: map[string]string{
				"dag_id":               "test_dag",
				"deployment_id":        "clxxxxxxxxxxxxxxxxxxxx",
				"deployment_api_token": "tok_secret",
			},
		},
		{
			name: "DAG_TRIGGER without deploymentApiToken (typical API response with sensitive field stripped)",
			def: map[string]interface{}{
				"dagId":        "test_dag",
				"deploymentId": "clxxxxxxxxxxxxxxxxxxxx",
			},
			wantStrings: map[string]string{
				"dag_id":        "test_dag",
				"deployment_id": "clxxxxxxxxxxxxxxxxxxxx",
			},
		},
		{
			name: "EMAIL with recipients",
			def: map[string]interface{}{
				"recipients": []interface{}{"a@example.com", "b@example.com"},
			},
			wantRecipients: []string{"a@example.com", "b@example.com"},
		},
		{
			// API response with an explicit empty recipients array — distinct from absent recipients key.
			// The decoder produces an empty non-null set here. Plain absence would produce a null set;
			// the difference is visible in Terraform plan diffs and matters for permadiff prevention.
			name: "EMAIL with empty recipients array",
			def: map[string]interface{}{
				"recipients": []interface{}{},
			},
			wantRecipients: []string{},
		},
		{
			name: "OPSGENIE with apiKey",
			def: map[string]interface{}{
				"apiKey": "ops_key",
			},
			wantStrings: map[string]string{"api_key": "ops_key"},
		},
		{
			name: "PAGERDUTY with integrationKey",
			def: map[string]interface{}{
				"integrationKey": "pd_int_key",
			},
			wantStrings: map[string]string{"integration_key": "pd_int_key"},
		},
		{
			name: "SLACK with webhookUrl",
			def: map[string]interface{}{
				"webhookUrl": "https://hooks.slack.com/services/x/y/z",
			},
			wantStrings: map[string]string{"webhook_url": "https://hooks.slack.com/services/x/y/z"},
		},
		{
			name: "extra unknown camelCase key is ignored, not raised as a diag",
			def: map[string]interface{}{
				"dagId":           "test_dag",
				"deploymentId":    "clxxxxxxxxxxxxxxxxxxxx",
				"someFutureField": "ignored",
			},
			wantStrings: map[string]string{
				"dag_id":        "test_dag",
				"deployment_id": "clxxxxxxxxxxxxxxxxxxxx",
			},
		},
		{
			name: "nil definition produces null attrs without error",
			def:  nil,
		},
		{
			name: "empty definition produces null attrs without error",
			def:  map[string]interface{}{},
		},
	}

	allStringAttrs := []string{
		"dag_id", "deployment_api_token", "deployment_id",
		"api_key", "integration_key", "webhook_url",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj, diags := models.NotificationChannelDefinitionDataSourceTypesObject(context.Background(), tc.def)
			if tc.expectErr {
				assert.True(t, diags.HasError(), "expected diags to have an error")
				return
			}
			assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
			// Named regression guard for CPP-737: the original failure surfaced as an
			// "Extra Object Attribute Name: <key>" diagnostic from types.ObjectValue when
			// raw camelCase API keys leaked into a schema-keyed attr map.
			for _, d := range diags {
				assert.NotContains(t, d.Detail(), "Extra Object Attribute Name",
					"CPP-737 regression: unexpected schema attribute mismatch")
				assert.NotContains(t, d.Summary(), "Extra Object Attribute Name",
					"CPP-737 regression: unexpected schema attribute mismatch")
			}

			attrs := obj.Attributes()

			for _, k := range allStringAttrs {
				s, _ := attrs[k].(types.String)
				if want, ok := tc.wantStrings[k]; ok {
					assert.Equal(t, want, s.ValueString(), "attribute %s", k)
				} else {
					assert.True(t, s.IsNull(), "expected %s to be null", k)
				}
			}

			recipients, _ := attrs["recipients"].(types.Set)
			if tc.wantRecipients == nil {
				assert.True(t, recipients.IsNull(), "expected recipients to be null")
			} else {
				assert.False(t, recipients.IsNull(), "expected recipients to be a non-null set")
				elems := recipients.Elements()
				got := make([]string, 0, len(elems))
				for _, e := range elems {
					s, _ := e.(types.String)
					got = append(got, s.ValueString())
				}
				assert.ElementsMatch(t, tc.wantRecipients, got)
			}
		})
	}
}
