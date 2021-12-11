package repository

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestRepositoryBase_UpdateNextFundingScheduleDate(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			repo := GetTestAuthenticatedRepository(t, db)

			account, err := repo.GetAccount(ctx)
			require.NoError(t, err, "must retrieve account")

			timezone, err := account.GetTimezone()
			require.NoError(t, err, "must be able to parse account's timezone")

			bankAccounts, err := repo.GetBankAccounts(ctx)
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

			err = repo.CreateFundingSchedule(ctx, &fundingSchedule)
			assert.NoError(t, err, "must be able to create funding schedule successfully")

			assert.Nil(t, fundingSchedule.LastOccurrence, "last occurrence should still be nil")

			ok := fundingSchedule.CalculateNextOccurrence(ctx, timezone)
			assert.True(t, ok, "calculate next occurrence should return true")
			assert.Greater(t, fundingSchedule.NextOccurrence.Unix(), originalOccurrence.Unix(), "next occurrence should be greater than the original")
			assert.NotNil(t, fundingSchedule.LastOccurrence, "last occurrence should no longer be nil")

			err = repo.UpdateNextFundingScheduleDate(ctx, fundingSchedule.FundingScheduleId, fundingSchedule.NextOccurrence)
			assert.NoError(t, err, "should succeed in updating funding schedule in database")

			fundingScheduleUpdated, err := repo.GetFundingSchedule(ctx, fundingSchedule.BankAccountId, fundingSchedule.FundingScheduleId)
			assert.NoError(t, err, "should be able to retrieve updated funding schedule")
			assert.NotNil(t, fundingScheduleUpdated.LastOccurrence, "last occurrence should have changed with update")
			assert.Equal(t, originalOccurrence.Unix(), fundingScheduleUpdated.LastOccurrence.Unix(), "last occurrence should match the original occurrence")
			assert.Equal(t, fundingSchedule.NextOccurrence.Unix(), fundingScheduleUpdated.NextOccurrence.Unix(), "next occurrence should match the expected time")
		})
	})
}
