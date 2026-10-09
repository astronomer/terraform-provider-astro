package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// configuredBool gates what the cluster create/update requests send for
// is_private_network_egress_enabled. Only an explicitly configured value may be sent: the
// attribute is Optional+Computed, so an omitted value arrives as unknown on create and as the
// value last read from the API on update, and sending either would pin a platform default.
func TestUnit_ConfiguredBool(t *testing.T) {
	for name, tc := range map[string]struct {
		value types.Bool
		want  *bool
	}{
		"true":    {value: types.BoolValue(true), want: boolPtr(true)},
		"false":   {value: types.BoolValue(false), want: boolPtr(false)},
		"null":    {value: types.BoolNull(), want: nil},
		"unknown": {value: types.BoolUnknown(), want: nil},
	} {
		t.Run(name, func(t *testing.T) {
			got := configuredBool(tc.value)
			if tc.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, *tc.want, *got)
		})
	}
}
