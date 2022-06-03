package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
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

func TestRepositoryBase_UpdateNextFundingScheduleDate(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestAuthenticatedRepository(t)

		account, err := repo.GetAccount(context.Background())
		require.NoError(t, err, "must retrieve account")

		timezone, err := account.GetTimezone()
		require.NoError(t, err, "must be able to parse account's timezone")

		bankAccounts, err := repo.GetBankAccounts(context.Background())
		require.NoError(t, err, "must be able to retrieve bank accounts")
		require.NotEmpty(t, bankAccounts, "must have at least one bank account to work with")

		bankAccount := bankAccounts[0]

		rule, err := models.NewRule("FREQ=DAILY")
		require.NoError(t, err, "must be able to create a rule")

		originalOccurrence := time.Now().Add(-1 * time.Minute)

		fundingSchedule := models.FundingSchedule{
			AccountId:      bankAccount.AccountId,
			BankAccountId:  bankAccount.BankAccountId,
			Name:           "Test Funding Schedule For Update",
			Description:    t.Name(),
			Rule:           rule,
			LastOccurrence: nil,
			NextOccurrence: originalOccurrence,
		}

		err = repo.CreateFundingSchedule(context.Background(), &fundingSchedule)
		assert.NoError(t, err, "must be able to create funding schedule successfully")

		assert.Nil(t, fundingSchedule.LastOccurrence, "last occurrence should still be nil")

		ok := fundingSchedule.CalculateNextOccurrence(context.Background(), timezone)
		assert.True(t, ok, "calculate next occurrence should return true")
		assert.Greater(t, fundingSchedule.NextOccurrence.Unix(), originalOccurrence.Unix(), "next occurrence should be greater than the original")
		assert.NotNil(t, fundingSchedule.LastOccurrence, "last occurrence should no longer be nil")

		err = repo.UpdateNextFundingScheduleDate(context.Background(), fundingSchedule.FundingScheduleId, fundingSchedule.NextOccurrence)
		assert.NoError(t, err, "should succeed in updating funding schedule in database")

		fundingScheduleUpdated, err := repo.GetFundingSchedule(context.Background(), fundingSchedule.BankAccountId, fundingSchedule.FundingScheduleId)
		assert.NoError(t, err, "should be able to retrieve updated funding schedule")
		assert.NotNil(t, fundingScheduleUpdated.LastOccurrence, "last occurrence should have changed with update")
		assert.Equal(t, originalOccurrence.Unix(), fundingScheduleUpdated.LastOccurrence.Unix(), "last occurrence should match the original occurrence")
		assert.Equal(t, fundingSchedule.NextOccurrence.Unix(), fundingScheduleUpdated.NextOccurrence.Unix(), "next occurrence should match the expected time")
	})
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
