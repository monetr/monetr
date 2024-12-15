package ofx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Run("nfcu valid file", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		assert.True(t, Validate(data), "nfcu OFX file should be valid")
	})

	t.Run("nfcu valid file 2", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu-2.qfx")
		assert.True(t, Validate(data), "nfcu OFX file should be valid")
	})

	t.Run("us bank valid file", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		assert.True(t, Validate(data), "us bank OFX file should be valid")
	})

	t.Run("invalid file should return false", func(t *testing.T) {
		data := GetFixtures(t, "invalid.qfx")
		assert.False(t, Validate(data), "invalid OFX file should not be valid")
	})
}
