package mock_secrets

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMockPlaidSecrets(t *testing.T) {
	plaidSecrets := NewMockPlaidSecrets()
	assert.NoError(t, plaidSecrets.Close(), "should succeed")
}

func TestMockPlaidSecrets_WithSecret(t *testing.T) {
	plaidSecrets := NewMockPlaidSecrets()

	{ // Will fail the first time because the data is not yet present.
		result, err := plaidSecrets.GetAccessTokenForPlaidLinkId(context.Background(), 1234, "abc")
		assert.Error(t, err, "should fail as secret does not exist")
		assert.Empty(t, result, "should be empty")
	}

	plaidSecrets = plaidSecrets.WithSecret(1234, "abc", "password")

	{ // Will succeed the second time because the data should now be there
		result, err := plaidSecrets.GetAccessTokenForPlaidLinkId(context.Background(), 1234, "abc")
		assert.NoError(t, err, "should succeed as secret should be present")
		assert.Equal(t, "password", result, "access token should match stored")
	}
}
