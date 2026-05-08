package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAdminPassword(t *testing.T) {
	const (
		lower   = "abcdefghijklmnopqrstuvwxyz"
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits  = "0123456789"
		special = "!@#%^&*()-_=+[]{}<>?"
	)
	pool := strings.Join([]string{lower, upper, digits, special}, "")

	// Run a healthy number of iterations so the "at least one of each class"
	// guarantee is exercised against the underlying RNG. A single sample would
	// not catch a regression that drops one of the anchor characters.
	for i := 0; i < 1000; i++ {
		password, err := generateAdminPassword()
		require.NoError(t, err, "must generate a password without error")
		require.Len(t, password, 16, "password must be 16 characters")

		assert.True(t, strings.ContainsAny(password, lower), "password must contain a lowercase letter: %q", password)
		assert.True(t, strings.ContainsAny(password, upper), "password must contain an uppercase letter: %q", password)
		assert.True(t, strings.ContainsAny(password, digits), "password must contain a digit: %q", password)
		assert.True(t, strings.ContainsAny(password, special), "password must contain a special character: %q", password)

		for _, r := range password {
			assert.Containsf(t, pool, string(r), "character %q from password %q is outside the allowed pool", string(r), password)
		}
	}
}

func TestGenerateAdminPasswordIsNotDeterministic(t *testing.T) {
	// Two consecutive calls returning the same password would be a sign that
	// crypto/rand is wired up incorrectly. The collision probability with a real
	// RNG is astronomically small.
	first, err := generateAdminPassword()
	require.NoError(t, err)
	second, err := generateAdminPassword()
	require.NoError(t, err)
	assert.NotEqual(t, first, second, "two generated passwords must differ")
}
