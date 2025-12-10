package myownsanity_test

import (
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestPointer(t *testing.T) {
	t.Run("float32", func(t *testing.T) {
		var input float32 = 1.2345
		result := myownsanity.Pointer(input)
		assert.NotNil(t, result, "resulting pointer should never be nil")
		assert.Equal(t, input, *result, "and the underlying value should match the input")
	})

	t.Run("int32", func(t *testing.T) {
		var input int32 = 12345
		result := myownsanity.Pointer(input)
		assert.NotNil(t, result, "resulting pointer should never be nil")
		assert.Equal(t, input, *result, "and the underlying value should match the input")
	})

	t.Run("time", func(t *testing.T) {
		input := time.Now()
		result := myownsanity.Pointer(input)
		assert.NotNil(t, result, "output must not be nil")
		assert.EqualValues(t, input, *result, "output's value must match input")
	})
}
