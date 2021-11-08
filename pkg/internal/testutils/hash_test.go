package testutils

import (
	"testing"

	"github.com/monetr/monetr/pkg/hash"
	"github.com/stretchr/testify/assert"
)

func TestMustHashLogin(t *testing.T) {
	email, password := "test@test.com", "superSecretPassword"
	result := MustHashLogin(t, email, password)
	assert.NotEmpty(t, result, "resulting hash must not be empty")
	assert.Equal(t, hash.HashPassword(email, password), result, "must match the hash package's result")
}
