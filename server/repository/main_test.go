package repository

import (
	"testing"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTestAuthenticatedRepository(t *testing.T) Repository {
	db := testutils.GetPgDatabase(t)

	user, _ := testutils.SeedAccount(t, db, testutils.WithPlaidAccount)

	txn, err := db.Begin()
	require.NoError(t, err, "failed to begin transaction")

	t.Cleanup(func() {
		assert.NoError(t, txn.Commit(), "should commit")
	})

	return &repositoryBase{
		userId:    user.UserId,
		accountId: user.AccountId,
		txn:       txn,
		account:   user.Account,
	}
}
