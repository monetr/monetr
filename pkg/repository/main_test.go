package repository

import (
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
