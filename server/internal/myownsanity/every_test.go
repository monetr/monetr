package myownsanity_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestEvery(t *testing.T) {
	assert.True(t, myownsanity.Every(true, true, true))
	assert.False(t, myownsanity.Every(true, false, true))
}
