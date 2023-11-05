package repository_test

import (
	"context"
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTestAuthenticatedRepository(t *testing.T) repository.Repository {
	db := testutils.GetPgDatabase(t)

	user, _ := testutils.SeedAccount(t, db, testutils.WithPlaidAccount)

	txn, err := db.Begin()
	require.NoError(t, err, "failed to begin transaction")

	t.Cleanup(func() {
		assert.NoError(t, txn.Commit(), "should commit")
	})

	return repository.NewRepositoryFromSession(user.UserId, user.AccountId, txn)
}

func TestRepositoryBase_UpdateFundingSchedule(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, &bankAccount, "FREQ=DAILY", false)

		repo := repository.NewRepositoryFromSession(user.UserId, user.AccountId, testutils.GetPgDatabase(t))

		fundingSchedule.Name = "Updated name"

		err := repo.UpdateFundingSchedule(context.Background(), fundingSchedule)
		assert.NoError(t, err, "must be able to update funding schedule")
		updatedSchedule := testutils.MustDBRead(t, *fundingSchedule)
		assert.Equal(t, "Updated name", updatedSchedule.Name, "name should match the new one")
	})
}
