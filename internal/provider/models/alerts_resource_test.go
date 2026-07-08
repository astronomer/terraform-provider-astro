package models

import (
	"context"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnit_AlertsReadFromResponse verifies that a labs.Alert response maps onto the resource model,
// including notification channels (the field whose absence caused the "inconsistent result after
// apply" error) and the rules object (which exercises the client-agnostic AlertRulesResourceTypesObject).
func TestUnit_AlertsReadFromResponse(t *testing.T) {
	alert := labs.Alert{
		Id:         "clmh7vdf4000008lhhlnk9t6o",
		Name:       "my dag failure",
		Type:       labs.AlertType("DAG_FAILURE"),
		Severity:   labs.AlertSeverity("CRITICAL"),
		EntityId:   "clmh8ol3x000008jo656y4285",
		EntityType: labs.AlertEntityType("DEPLOYMENT"),
		Rules: labs.AlertRules{
			Properties: map[string]any{"deploymentId": "clmh8ol3x000008jo656y4285"},
			PatternMatches: &[]labs.PatternMatch{
				{
					EntityType:   labs.PatternMatchEntityType("DAG_ID"),
					OperatorType: labs.PatternMatchOperatorType("IS"),
					Values:       []string{"my_dag"},
				},
			},
		},
		NotificationChannels: &[]labs.AlertNotificationChannel{
			{Id: "clmk2qqia000008mhff3ndjr0"},
			{Id: "clmk2qqia000008mhff3ndjr1"},
		},
	}

	var e AlertsResourceElementModel
	diags := e.ReadFromResponse(context.Background(), &alert)
	require.False(t, diags.HasError(), "ReadFromResponse should not error: %v", diags)

	assert.Equal(t, "clmh7vdf4000008lhhlnk9t6o", e.Id.ValueString())
	assert.Equal(t, "my dag failure", e.Name.ValueString())
	assert.Equal(t, "DAG_FAILURE", e.Type.ValueString())
	assert.Equal(t, "CRITICAL", e.Severity.ValueString())
	assert.Equal(t, "DEPLOYMENT", e.EntityType.ValueString())

	// Notification channels must be populated from the response — this is the regression guard.
	var ncIds []string
	ncDiags := e.NotificationChannelIds.ElementsAs(context.Background(), &ncIds, false)
	require.False(t, ncDiags.HasError(), "extracting notification channel ids should not error: %v", ncDiags)
	assert.ElementsMatch(t, []string{"clmk2qqia000008mhff3ndjr0", "clmk2qqia000008mhff3ndjr1"}, ncIds)

	// Rules must map cleanly from labs types (AlertRulesResourceTypesObject is client-agnostic).
	assert.False(t, e.Rules.IsNull(), "rules object should be populated")
}

// TestUnit_AlertsReadFromResponse_NoNotificationChannels verifies the mapping is safe when the
// response omits notification channels.
func TestUnit_AlertsReadFromResponse_NoNotificationChannels(t *testing.T) {
	alert := labs.Alert{
		Id:         "clmh7vdf4000008lhhlnk9t6o",
		Name:       "no channels",
		Type:       labs.AlertType("DAG_DURATION"),
		Severity:   labs.AlertSeverity("WARNING"),
		EntityId:   "clmh8ol3x000008jo656y4285",
		EntityType: labs.AlertEntityType("DEPLOYMENT"),
		Rules: labs.AlertRules{
			Properties: map[string]any{
				"deploymentId":       "clmh8ol3x000008jo656y4285",
				"dagDurationSeconds": float64(3600),
			},
		},
	}

	var e AlertsResourceElementModel
	diags := e.ReadFromResponse(context.Background(), &alert)
	require.False(t, diags.HasError(), "ReadFromResponse should not error: %v", diags)

	var ncIds []string
	ncDiags := e.NotificationChannelIds.ElementsAs(context.Background(), &ncIds, false)
	require.False(t, ncDiags.HasError())
	assert.Empty(t, ncIds)
}
