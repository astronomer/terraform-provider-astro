package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_chunkSlice(t *testing.T) {
	t.Run("empty input returns no chunks", func(t *testing.T) {
		assert.Empty(t, chunkSlice([]int{}, 30))
	})

	t.Run("fewer than size returns a single chunk", func(t *testing.T) {
		chunks := chunkSlice([]int{1, 2, 3}, 30)
		assert.Equal(t, [][]int{{1, 2, 3}}, chunks)
	})

	t.Run("exactly size returns a single chunk", func(t *testing.T) {
		items := make([]int, 30)
		for i := range items {
			items[i] = i
		}
		chunks := chunkSlice(items, 30)
		assert.Len(t, chunks, 1)
		assert.Len(t, chunks[0], 30)
	})

	t.Run("size+1 splits into two chunks preserving order", func(t *testing.T) {
		items := make([]int, 31)
		for i := range items {
			items[i] = i
		}
		chunks := chunkSlice(items, 30)
		assert.Len(t, chunks, 2)
		assert.Len(t, chunks[0], 30)
		assert.Equal(t, []int{30}, chunks[1])

		// concatenation must equal the original input in order
		var flat []int
		for _, c := range chunks {
			flat = append(flat, c...)
		}
		assert.Equal(t, items, flat)
	})

	t.Run("delete limit of 20 chunks correctly", func(t *testing.T) {
		items := make([]string, 45)
		for i := range items {
			items[i] = "id"
		}
		chunks := chunkSlice(items, alertsBulkDeleteLimit)
		assert.Len(t, chunks, 3) // 20 + 20 + 5
		assert.Len(t, chunks[0], 20)
		assert.Len(t, chunks[1], 20)
		assert.Len(t, chunks[2], 5)
	})

	t.Run("non-positive size returns a single chunk", func(t *testing.T) {
		chunks := chunkSlice([]int{1, 2, 3}, 0)
		assert.Equal(t, [][]int{{1, 2, 3}}, chunks)
	})
}

func TestUnit_sortedKeys(t *testing.T) {
	t.Run("returns keys in deterministic sorted order", func(t *testing.T) {
		m := map[string]int{"charlie": 3, "alpha": 1, "bravo": 2}
		assert.Equal(t, []string{"alpha", "bravo", "charlie"}, sortedKeys(m))
	})

	t.Run("empty map returns empty slice", func(t *testing.T) {
		assert.Empty(t, sortedKeys(map[string]int{}))
	})
}
