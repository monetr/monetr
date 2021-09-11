package myownsanity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat32P(t *testing.T) {
	var input float32 = 1.2345
	result := Float32P(input)
	assert.NotNil(t, result, "resulting pointer should never be nil")
	assert.Equal(t, input, *result, "and the underlying value should match the input")
}

func TestInt32P(t *testing.T) {
	var input int32 = 12345
	result := Int32P(input)
	assert.NotNil(t, result, "resulting pointer should never be nil")
	assert.Equal(t, input, *result, "and the underlying value should match the input")
}

func TestMax(t *testing.T) {
	assert.Equal(t, 2, Max(1, 2))
	assert.Equal(t, 1000, Max(1000, 100))
	assert.Equal(t, 500, Max(500, 500))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1, Min(1, 2))
	assert.Equal(t, 100, Min(1000, 100))
	assert.Equal(t, 500, Min(500, 500))
}
