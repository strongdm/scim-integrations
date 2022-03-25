package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	array := []string{"a", "b"}

	t.Run("should return true when an item exists in array", func(t *testing.T) {
		assert.True(t, StringArrayContains(array, "a"))
	})

	t.Run("should return false when an item don't exists in array", func(t *testing.T) {
		assert.False(t, StringArrayContains(array, "c"))
	})
}
