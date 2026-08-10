package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAzureRegionContinent(t *testing.T) {
	tests := []struct {
		name          string
		region        string
		wantContinent string
		wantOk        bool
	}{
		{"known region", "eastus2", "americas", true},
		{"another continent", "westeurope", "europe", true},
		{"case insensitive", "EastUS2", "americas", true},
		{"mixed case", "WestEurope", "europe", true},
		{"unknown region", "notarealregion", "", false},
		{"empty region", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			continent, ok := azureRegionContinent(tt.region)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantContinent, continent)
		})
	}
}

// azureClusterResourceData builds a minimal models.ClusterResource for exercising
// validateAzureConfig's Replication Time Control region-pair check in isolation.
func azureClusterResourceData(region, drRegion string, isDrEnabled bool, enableRTC *bool) *models.ClusterResource {
	data := &models.ClusterResource{
		CloudProvider: types.StringValue("AZURE"),
		Region:        types.StringValue(region),
		IsDrEnabled:   types.BoolValue(isDrEnabled),
	}
	if drRegion == "" {
		data.DrRegion = types.StringNull()
	} else {
		data.DrRegion = types.StringValue(drRegion)
	}
	if enableRTC == nil {
		data.EnableReplicationTimeControl = types.BoolNull()
	} else {
		data.EnableReplicationTimeControl = types.BoolValue(*enableRTC)
	}
	return data
}

func boolPtr(b bool) *bool { return &b }

func TestValidateAzureConfig_ReplicationTimeControlRegionPair(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		region      string
		drRegion    string
		enableRTC   *bool
		expectError string // substring expected in a diagnostic summary; empty means no error
	}{
		{
			name:      "same continent, RTC explicitly true",
			region:    "eastus2",
			drRegion:  "westus2",
			enableRTC: boolPtr(true),
		},
		{
			name:      "same continent, RTC unset",
			region:    "eastus2",
			drRegion:  "westus2",
			enableRTC: nil,
		},
		{
			name:        "different continent, RTC explicitly true",
			region:      "eastus2",
			drRegion:    "westeurope",
			enableRTC:   boolPtr(true),
			expectError: "Replication time control requires regions on the same continent",
		},
		{
			// Left unset, the provider simply won't send true (see
			// TestAzureEffectiveEnableReplicationTimeControl) - no error needed.
			name:      "different continent, RTC unset",
			region:    "eastus2",
			drRegion:  "westeurope",
			enableRTC: nil,
		},
		{
			name:      "different continent, RTC explicitly false opts out",
			region:    "eastus2",
			drRegion:  "westeurope",
			enableRTC: boolPtr(false),
		},
		{
			name:        "unrecognized primary region, RTC explicitly true",
			region:      "notarealregion",
			drRegion:    "westus2",
			enableRTC:   boolPtr(true),
			expectError: "has no defined location",
		},
		{
			name:        "unrecognized dr_region, RTC explicitly true",
			region:      "eastus2",
			drRegion:    "notarealregion",
			enableRTC:   boolPtr(true),
			expectError: "has no defined location",
		},
		{
			// Left unset, an unrecognized region just means the provider can't
			// compute a default - it isn't a validation error.
			name:      "unrecognized primary region, RTC unset",
			region:    "notarealregion",
			drRegion:  "westus2",
			enableRTC: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := azureClusterResourceData(tt.region, tt.drRegion, true, tt.enableRTC)
			diags := validateAzureConfig(ctx, data)

			if tt.expectError == "" {
				assert.False(t, diags.HasError(), "expected no error, got: %v", diags)
				return
			}

			assert.True(t, diags.HasError(), "expected an error")
			found := false
			for _, d := range diags {
				if strings.Contains(d.Summary(), tt.expectError) {
					found = true
					break
				}
			}
			assert.True(t, found, "expected a diagnostic with summary containing %q, got: %v", tt.expectError, diags)
		})
	}
}

func TestValidateAzureConfig_ReplicationTimeControlSkipsWhenDrDisabled(t *testing.T) {
	ctx := context.Background()

	// DR disabled: region-pair validation should not run at all (a different
	// diagnostic - "enable_replication_time_control is only valid..." - fires
	// instead, since RTC is only meaningful when DR is enabled).
	data := azureClusterResourceData("eastus2", "westeurope", false, boolPtr(true))
	diags := validateAzureConfig(ctx, data)

	assert.True(t, diags.HasError())
	for _, d := range diags {
		assert.NotEqual(t, "Replication time control requires regions on the same continent", d.Summary())
	}
}

func TestAzureEffectiveEnableReplicationTimeControl(t *testing.T) {
	tests := []struct {
		name      string
		region    string
		drRegion  string
		enableRTC *bool
		want      *bool
	}{
		{
			name:      "unset, same continent -> computed true",
			region:    "eastus2",
			drRegion:  "westus2",
			enableRTC: nil,
			want:      boolPtr(true),
		},
		{
			name:      "unset, different continent -> nothing sent",
			region:    "eastus2",
			drRegion:  "westeurope",
			enableRTC: nil,
			want:      nil,
		},
		{
			name:      "unset, unrecognized region -> nothing sent",
			region:    "notarealregion",
			drRegion:  "westus2",
			enableRTC: nil,
			want:      nil,
		},
		{
			name:      "explicit true, same continent -> passed through",
			region:    "eastus2",
			drRegion:  "westus2",
			enableRTC: boolPtr(true),
			want:      boolPtr(true),
		},
		{
			name: "explicit true, different continent -> passed through unchanged",
			// validateAzureConfig is what rejects this combination; this function
			// just forwards the user's explicit choice.
			region:    "eastus2",
			drRegion:  "westeurope",
			enableRTC: boolPtr(true),
			want:      boolPtr(true),
		},
		{
			name:      "explicit false, same continent -> passed through",
			region:    "eastus2",
			drRegion:  "westus2",
			enableRTC: boolPtr(false),
			want:      boolPtr(false),
		},
		{
			name:      "explicit false, different continent -> passed through",
			region:    "eastus2",
			drRegion:  "westeurope",
			enableRTC: boolPtr(false),
			want:      boolPtr(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := azureClusterResourceData(tt.region, tt.drRegion, true, tt.enableRTC)
			got := azureEffectiveEnableReplicationTimeControl(data)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestAzureEffectiveEnableReplicationTimeControl_UnknownValuesDeferred(t *testing.T) {
	// dr_region derived from an unknown value (e.g. another resource's attribute):
	// the provider can't compute a default, so it must not guess - send nothing.
	data := &models.ClusterResource{
		CloudProvider: types.StringValue("AZURE"),
		Region:        types.StringValue("eastus2"),
		IsDrEnabled:   types.BoolValue(true),
		DrRegion:      types.StringUnknown(),
	}
	assert.Nil(t, azureEffectiveEnableReplicationTimeControl(data))

	// enable_replication_time_control itself unknown: this is the normal plan-time
	// state for "left unset" on an Optional+Computed attribute with no plan
	// modifiers (Create, or Update with no prior explicit value) - it must be
	// treated the same as unset-and-null and computed from the region pair, not
	// deferred. Region/dr_region are on the same continent here, so this computes
	// to true.
	dataUnknownRTC := &models.ClusterResource{
		CloudProvider:                types.StringValue("AZURE"),
		Region:                       types.StringValue("eastus2"),
		IsDrEnabled:                  types.BoolValue(true),
		DrRegion:                     types.StringValue("westus2"),
		EnableReplicationTimeControl: types.BoolUnknown(),
	}
	got := azureEffectiveEnableReplicationTimeControl(dataUnknownRTC)
	if assert.NotNil(t, got) {
		assert.True(t, *got)
	}
}

func TestValidateAzureConfig_ReplicationTimeControlSkipsUnknownValues(t *testing.T) {
	ctx := context.Background()

	// dr_region derived from an unknown value (e.g. another resource's attribute)
	// must not trigger a false-positive continent mismatch at plan time.
	data := &models.ClusterResource{
		CloudProvider:                types.StringValue("AZURE"),
		Region:                       types.StringValue("eastus2"),
		IsDrEnabled:                  types.BoolValue(true),
		DrRegion:                     types.StringUnknown(),
		EnableReplicationTimeControl: types.BoolValue(true),
	}
	diags := validateAzureConfig(ctx, data)

	for _, d := range diags {
		assert.NotEqual(t, "Replication time control requires regions on the same continent", d.Summary())
	}
}
