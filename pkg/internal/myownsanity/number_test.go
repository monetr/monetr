package myownsanity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMax(t *testing.T) {
	assert.Equal(t, 2, Max(1, 2))
	assert.Equal(t, 1000, Max(1000, 100))
	assert.Equal(t, 500, Max(500, 500))
}
