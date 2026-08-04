package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_diffCidrs(t *testing.T) {
	t.Run("no changes returns no create or delete", func(t *testing.T) {
		toCreate, toDelete := diffCidrs([]string{"10.0.0.0/8"}, []string{"10.0.0.0/8"})
		assert.Empty(t, toCreate)
		assert.Empty(t, toDelete)
	})

	t.Run("new plan entries are created", func(t *testing.T) {
		toCreate, toDelete := diffCidrs([]string{"10.0.0.0/8", "192.168.0.0/16"}, []string{"10.0.0.0/8"})
		assert.Equal(t, []string{"192.168.0.0/16"}, toCreate)
		assert.Empty(t, toDelete)
	})

	t.Run("removed state entries are deleted", func(t *testing.T) {
		toCreate, toDelete := diffCidrs([]string{"10.0.0.0/8"}, []string{"10.0.0.0/8", "192.168.0.0/16"})
		assert.Empty(t, toCreate)
		assert.Equal(t, []string{"192.168.0.0/16"}, toDelete)
	})

	t.Run("disjoint sets create and delete", func(t *testing.T) {
		toCreate, toDelete := diffCidrs([]string{"172.16.0.0/12"}, []string{"192.168.0.0/16"})
		assert.Equal(t, []string{"172.16.0.0/12"}, toCreate)
		assert.Equal(t, []string{"192.168.0.0/16"}, toDelete)
	})

	t.Run("empty plan deletes everything in state", func(t *testing.T) {
		toCreate, toDelete := diffCidrs(nil, []string{"10.0.0.0/8", "192.168.0.0/16"})
		assert.Empty(t, toCreate)
		assert.ElementsMatch(t, []string{"10.0.0.0/8", "192.168.0.0/16"}, toDelete)
	})
}
