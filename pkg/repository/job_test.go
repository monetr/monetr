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
)

func TestJobRepository_GetPlaidLinksByAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		_ = fixtures.GivenIHaveAPlaidLink(t, user)
		_ = fixtures.GivenIHaveAPlaidLink(t, user)

		plaidLinks, err := jobRepo.GetPlaidLinksByAccount(context.Background())
		assert.NoError(t, err, "should be able to retrieve the two links")
		assert.Len(t, plaidLinks, 1, "should retrieve the one account")
		assert.Len(t, plaidLinks[0].LinkIds, 2, "should have two links for the one account")
	})
}

func TestJobRepository_GetBankAccountsWithStaleSpending(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db)
		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAPlaidLink(t, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:      bankAccount.AccountId,
			BankAccountId:  bankAccount.BankAccountId,
			Name:           "Payday",
			Description:    "Payday",
			Rule:           fundingRule,
			NextOccurrence: fundingRule.After(time.Now(), false),
		})

		spendingRule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO")
		spendingRule.DTStart(time.Now().Add(-8 * 24 * time.Hour)) // Allow past times.
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:      bankAccount.AccountId,
			BankAccountId:  bankAccount.BankAccountId,
			SpendingType:   models.SpendingTypeExpense,
			Name:           "Test Stale Expense",
			Description:    "Description or something",
			TargetAmount:   5000,
			CurrentAmount:  5000,
			RecurrenceRule: spendingRule,
			NextRecurrence: spendingRule.Before(time.Now(), true), // Make it so it recurs next in the past. (STALE)
			DateCreated:    time.Now(),
		})

		spendingFunding := testutils.MustInsert(t, models.SpendingFunding{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingId:        spending.SpendingId,
		})
		assert.NotZero(t, spendingFunding.SpendingFundingId)

		result, err := jobRepo.GetBankAccountsWithStaleSpending(context.Background())
		assert.NoError(t, err, "must not return an error")
		assert.NotEmpty(t, result, "should return at least one expense")
		assert.Equal(t, spending.BankAccountId, result[0].BankAccountId)
	})
}
