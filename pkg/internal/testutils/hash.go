package testutils

import (
	"testing"

	"github.com/monetr/monetr/pkg/hash"
	"github.com/stretchr/testify/require"
)

func MustHashLogin(t *testing.T, email, password string) string {
	require.NotEmpty(t, email, "email cannot be empty")
	require.NotEmpty(t, password, "password cannot be empty")
	return hash.HashPassword(email, password)
}
