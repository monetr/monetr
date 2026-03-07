package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLog(t *testing.T) {
	originalLog := GetLog(t)
	assert.NotNil(t, originalLog, "logging object must not be nil")

	secondLog := GetLog(t)
	assert.Equal(t, originalLog, secondLog, "requesting another log in the same test should return the same log")

	t.Run("separate test", func(t *testing.T) {
		separateLog := GetLog(t)
		assert.NotEqual(t, originalLog, separateLog, "but requesting a log in a separate test should be different")
	})
}
