package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetBalances(t *testing.T) {
	t.Run("will read balances", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

		balances, err := repo.GetBalances(t.Context(), bank.BankAccountId)
		assert.NoError(t, err, "must not return an error when reading balances")
		assert.NotNil(t, balances, "must return a balances object")
		assert.Equal(t, bank.AvailableBalance, balances.Available, "available balance should match original")
		assert.Equal(t, bank.CurrentBalance, balances.Current, "current balance should match original")
		assert.Equal(t, bank.LimitBalance, balances.Limit, "limit balance should match original")
		assert.Equal(t, bank.AvailableBalance, balances.Free, "without expenses or goals, free should match available")
	})

	t.Run("with spending objects", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

		{ // Before we create an expense, double check balances
			balances, err := repo.GetBalances(t.Context(), bank.BankAccountId)
			assert.NoError(t, err, "must not return an error when reading balances")
			assert.NotNil(t, balances, "must return a balances object")
			assert.Equal(t, bank.AvailableBalance, balances.Available, "available balance should match original")
			assert.Equal(t, bank.CurrentBalance, balances.Current, "current balance should match original")
			assert.Equal(t, bank.LimitBalance, balances.Limit, "limit balance should match original")
			assert.Equal(t, bank.AvailableBalance, balances.Free, "without expenses or goals, free should match available")
		}

		funding := fixtures.GivenIHaveAFundingSchedule(t, clock, &bank, "FREQ=DAILY;INTERVAL=1", false)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		rule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10", clock.Now())
		nextOccurrence := util.Midnight(rule.After(clock.Now(), false), timezone)

		spending := models.Spending{
			AccountId:              user.AccountId,
			BankAccountId:          bank.BankAccountId,
			FundingScheduleId:      funding.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Test Expense",
			Description:            "Testing",
			TargetAmount:           10000,
			CurrentAmount:          1000, // $10
			UsedAmount:             0,
			RuleSet:                rule,
			LastSpentFrom:          nil,
			LastRecurrence:         nil,
			NextRecurrence:         nextOccurrence,
			NextContributionAmount: 0,
			IsBehind:               false,
			IsPaused:               false,
		}
		assert.NoError(t, repo.CreateSpending(t.Context(), &spending), "must create a spending object")

		{ // After we have created an expense, check balances again
			balances, err := repo.GetBalances(t.Context(), bank.BankAccountId)
			assert.NoError(t, err, "must not return an error when reading balances")
			assert.NotNil(t, balances, "must return a balances object")
			assert.Equal(t, bank.AvailableBalance, balances.Available, "available balance should match original")
			assert.Equal(t, bank.CurrentBalance, balances.Current, "current balance should match original")
			assert.Equal(t, bank.LimitBalance, balances.Limit, "limit balance should match original")
			assert.Equal(t, bank.AvailableBalance-spending.CurrentAmount, balances.Free, "with an expense, our free amount should be less")
		}
	})
}
