package myownsanity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringP(t *testing.T) {
	var input string = "12345"
	result := StringP(input)
	assert.NotNil(t, result, "resulting pointer should never be nil")
	assert.Equal(t, input, *result, "and the underlying value should match the input")
}

func TestStringDefault(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		inputString := "I am a string"
		defaultString := "I am the default string"
		result := StringDefault(&inputString, defaultString)
		assert.Equal(t, inputString, result, "result should match the input string")
		assert.NotEqual(t, defaultString, result, "and should definitely not match the default string")
	})

	t.Run("nil", func(t *testing.T) {
		defaultString := "I am stil the default string"
		result := StringDefault(nil, defaultString)
		assert.Equal(t, defaultString, result, "when the input is nil, the default should be returned")
	})
}

func TestSliceContains(t *testing.T) {
	data := []string{
		"Item #1",
		"Item #2",
	}

	assert.True(t, SliceContains(data, "Item #1"), "should contain item #1")
	assert.False(t, SliceContains(data, "Item #3"), "should contain item #3")
}
