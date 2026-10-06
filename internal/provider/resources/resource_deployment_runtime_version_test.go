package resources_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/resources"
)

// Deployment options are sorted newest-first and can lead with a pre-release, which the
// create deployment endpoint rejects ("3.4-1-nightly20260630 is not a valid astro runtime
// version"). The newest stable release must be chosen instead of the first entry.
func TestUnit_LatestStableRuntimeVersion_SkipsLeadingPreReleases(t *testing.T) {
	version := resources.LatestStableRuntimeVersion(context.Background(), []platform.RuntimeRelease{
		{Version: "3.4-1-nightly20260630", Channel: "nightly"},
		{Version: "3.4-1-beta1", Channel: "beta"},
		{Version: "3.3-5", Channel: "stable"},
		{Version: "3.3-4", Channel: "stable"},
	})
	assert.Equal(t, "3.3-5", version)
}

// The common case: the newest release is already stable.
func TestUnit_LatestStableRuntimeVersion_TakesNewestStable(t *testing.T) {
	version := resources.LatestStableRuntimeVersion(context.Background(), []platform.RuntimeRelease{
		{Version: "3.4-2", Channel: "stable"},
		{Version: "3.4-1", Channel: "stable"},
	})
	assert.Equal(t, "3.4-2", version)
}

// With nothing stable on offer, fall back to the newest release rather than failing a create
// that may still succeed — this is what the provider did before stable filtering existed.
func TestUnit_LatestStableRuntimeVersion_FallsBackWhenNoStable(t *testing.T) {
	version := resources.LatestStableRuntimeVersion(context.Background(), []platform.RuntimeRelease{
		{Version: "3.4-1-nightly20260630", Channel: "nightly"},
		{Version: "3.4-1-beta1", Channel: "beta"},
	})
	assert.Equal(t, "3.4-1-nightly20260630", version)
}

// An empty channel is not stable and must not be selected over an explicitly stable release.
func TestUnit_LatestStableRuntimeVersion_EmptyChannelIsNotStable(t *testing.T) {
	version := resources.LatestStableRuntimeVersion(context.Background(), []platform.RuntimeRelease{
		{Version: "3.4-1-unknown", Channel: ""},
		{Version: "3.3-5", Channel: "stable"},
	})
	assert.Equal(t, "3.3-5", version)
}
