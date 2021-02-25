package testutils

import (
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func MustHashLogin(t *testing.T, email, password string) string {
	require.NotEmpty(t, email, "email cannot be empty")
	require.NotEmpty(t, password, "password cannot be empty")
	email = strings.ToLower(email)
	hash := sha256.New()
	hash.Write([]byte(email))
	hash.Write([]byte(password))
	return fmt.Sprintf("%X", hash.Sum(nil))
}
