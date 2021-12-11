package repository

import (
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func GetTestAuthenticatedRepository(t *testing.T, db *bun.DB) Repository {
	user, _ := testutils.SeedAccount(t, db, testutils.WithPlaidAccount)

	txn, err := db.Begin()
	require.NoError(t, err, "failed to begin transaction")

	t.Cleanup(func() {
		assert.NoError(t, txn.Commit(), "should commit")
	})

	return &repositoryBase{
		userId:    user.UserId,
		accountId: user.AccountId,
		db:        txn,
	}
}
