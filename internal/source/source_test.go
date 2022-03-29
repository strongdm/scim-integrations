package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByFlag(t *testing.T) {
	t.Run("should return a GoogleSource when passing \"google\"", func(t *testing.T) {
		source := ByFlag("google")

		assert.IsType(t, &GoogleSource{}, source)
	})

	t.Run("should return nil when passing a random value", func(t *testing.T) {
		source := ByFlag("random")

		assert.Nil(t, source)
	})
}
