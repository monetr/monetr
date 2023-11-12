package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestJobRepository_GetPlaidLinksByAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		_ = fixtures.GivenIHaveAPlaidLink(t, clock, user)
		_ = fixtures.GivenIHaveAPlaidLink(t, clock, user)

		plaidLinks, err := jobRepo.GetPlaidLinksByAccount(context.Background())
		assert.NoError(t, err, "should be able to retrieve the two links")
		assert.Len(t, plaidLinks, 1, "should retrieve the one account")
		assert.Len(t, plaidLinks[0].LinkIds, 2, "should have two links for the one account")
	})
}

func TestJobRepository_GetBankAccountsWithStaleSpending(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		fundingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", clock.Now())

		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday",
			Description:            "Payday",
			RuleSet:                fundingRule,
			NextOccurrence:         fundingRule.After(clock.Now(), false),
			NextOccurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour)) // Allow past times.
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Test Stale Expense",
			Description:       "Description or something",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true), // Make it so it recurs next in the past. (STALE)
			DateCreated:       clock.Now(),
		})

		result, err := jobRepo.GetBankAccountsWithStaleSpending(context.Background())
		assert.NoError(t, err, "must not return an error")
		assert.NotEmpty(t, result, "should return at least one expense")
		assert.Equal(t, spending.BankAccountId, result[0].BankAccountId)
	})
}

func TestJobRepository_GetLinksForExpiredAccounts(t *testing.T) {
	t.Run("subscribed account", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		{ // Before updating the subscription
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		// Then update the account to have a subscription that has expired more than 90 days ago.
		account := user.Account
		account.SubscriptionActiveUntil = myownsanity.TimeP(clock.Now().AddDate(0, 0, -100))
		testutils.MustDBUpdate(t, account)

		{ // After updating the subscription
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Len(t, result, 1, "should have one link that is expired")
			assert.EqualValues(t, link.LinkId, result[0].LinkId, "expired link should be the one created for this test")
		}
	})

	t.Run("trial account", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		// Update the account to be the same as it would be in a trial state.
		account := user.Account
		account.SubscriptionActiveUntil = nil
		account.SubscriptionStatus = nil
		account.StripeCustomerId = nil
		account.StripeSubscriptionId = nil
		account.TrialEndsAt = myownsanity.TimeP(clock.Now().AddDate(0, 0, 30))
		testutils.MustDBUpdate(t, account)

		{ // Then check to make sure that we don't consider this an expired account.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		account.TrialEndsAt = myownsanity.TimeP(clock.Now().AddDate(0, 0, -100))
		testutils.MustDBUpdate(t, account)

		{ // After updating the trial end date we should see it as expired.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Len(t, result, 1, "should have one link that is expired")
			assert.EqualValues(t, link.LinkId, result[0].LinkId, "expired link should be the one created for this test")
		}
	})
}
