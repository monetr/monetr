package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTestAuthenticatedRepository(t *testing.T, clock clock.Clock) repository.Repository {
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

	txn, err := db.Begin()
	require.NoError(t, err, "failed to begin transaction")

	t.Cleanup(func() {
		assert.NoError(t, txn.Commit(), "should commit")
	})

	return repository.NewRepositoryFromSession(
		clock,
		user.UserId,
		user.AccountId,
		txn,
		log,
	)
}
