package myownsanity_test

import (
	"testing"

	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestAssert(t *testing.T) {
	t.Run("does panic", func(t *testing.T) {
		assert.PanicsWithValue(t, "ASSERT FAILED: I'm a bad assertion", func() {
			myownsanity.Assert(false, "I'm a bad assertion")
		})
	})

	t.Run("does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			myownsanity.Assert(true, "I'm a good assertion")
		})
	})
}
