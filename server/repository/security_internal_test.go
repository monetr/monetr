package repository

import (
	"testing"

	"github.com/monetr/monetr/server/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestDummyPasswordHash(t *testing.T) {
	// If the dummy hash is cheaper or more expensive than real hashes then a
	// login for a missing email is distinguishable by timing again.
	cost, err := bcrypt.Cost(dummyPasswordHash)
	require.NoError(t, err, "dummy password hash must be a valid bcrypt hash")
	assert.Equal(t, consts.BcryptCost, cost, "dummy password hash must use the same cost as real password hashes, regenerate it if consts.BcryptCost changed")
}
