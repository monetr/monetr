package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func GetTestUnauthenticatedRepository(t *testing.T) UnauthenticatedRepository {
	txn := testutils.GetPgDatabaseTxn(t)
	return NewUnauthenticatedRepository(txn)
}

func TestUnauthenticatedRepo_CreateAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		account, err := repo.CreateAccount(time.UTC)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
	})
}
