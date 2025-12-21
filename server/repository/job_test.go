package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/google/uuid"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestJobRepository_GetBankAccountsWithStaleSpending(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		fundingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", clock.Now())

		fundingSchedule := testutils.MustInsert(t, FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday",
			Description:            "Payday",
			RuleSet:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour)) // Allow past times.
		spending := testutils.MustInsert(t, Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      SpendingTypeExpense,
			Name:              "Test Stale Expense",
			Description:       "Description or something",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true), // Make it so it recurs next in the past. (STALE)
			CreatedAt:         clock.Now(),
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

		account := user.Account
		account.SubscriptionActiveUntil = myownsanity.TimeP(clock.Now())
		testutils.MustDBUpdate(t, account)

		{ // After gaining a subscription, we should still not remove the account.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		// But if the subscription has not been updated in 100 days, then we should.
		clock.Add(100 * 24 * time.Hour)

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

		// Move the clock forward 31 days. The trial has now expired.
		clock.Add(31 * 24 * time.Hour)
		{ // But the account is not yet eligible for removal.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		// Move the clock forward 100 days, now the account should be eligible.
		clock.Add(100 * 24 * time.Hour)

		{ // 131 days after signup the account is now eligible for removal.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Len(t, result, 1, "should have one link that is expired")
			assert.EqualValues(t, link.LinkId, result[0].LinkId, "expired link should be the one created for this test")
		}
	})

	t.Run("trial then subscribe", func(t *testing.T) {
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

		// When the account is first created and still trialing, we don't want to
		// delete the data.
		{ // Then check to make sure that we don't consider this an expired account.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		// Move the clock forward 31 days. The trial has now expired.
		clock.Add(31 * 24 * time.Hour)
		{ // But the account is not yet eligible for removal.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		// On the 31st day the account becomes a subscriber. So we push their active
		// until date out 30 days.
		account.SubscriptionActiveUntil = myownsanity.TimeP(clock.Now().AddDate(0, 0, 30))

		// We are now 62 days after the account was originally created. We should
		// still not remove the account, even though their subscrition just expired
		// yesterday.
		clock.Add(31 * 24 * time.Hour)
		{ // But the account is not yet eligible for removal.
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Empty(t, result, "there should not be any expired links at the moment")
		}

		// But 100 days after the subscription expires, or 162 days after the
		// account was originally created; it is now safe to remove the link data.
		clock.Add(100 * 24 * time.Hour)
		{
			result, err := jobRepo.GetLinksForExpiredAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving links for expired accounts")
			assert.Len(t, result, 1, "should have one link that is expired")
			assert.EqualValues(t, link.LinkId, result[0].LinkId, "expired link should be the one created for this test")
		}
	})
}

func TestJobRepository_GetAccountsWithTooManyFiles(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		{
			// Check before we create the files, there shouldn't be any accounts right
			// now.
			tooMany, err := jobRepo.GetAccountsWithTooManyFiles(context.Background())
			assert.NoError(t, err, "should not return an error looking for too many files")
			assert.Empty(t, tooMany, "there should not be any accounts with too many files")
		}

		// Create a ton of files in a single account
		for i := 0; i < 100; i++ {
			testutils.MustDBInsert(t, &File{
				AccountId:   bankAccount.AccountId,
				Name:        uuid.NewString(),
				ContentType: "text/csv",
				Size:        100,
				BlobUri:     "bogus://temp",
				CreatedAt:   time.Now(),
				CreatedBy:   user.UserId,
			})
		}

		{
			// Now that some files exist we check again.
			tooMany, err := jobRepo.GetAccountsWithTooManyFiles(context.Background())
			assert.NoError(t, err, "should not return an error looking for too many files")
			assert.EqualValues(t, []repository.AccountWithTooManyFiles{
				{
					AccountId: bankAccount.AccountId,
					Count:     100,
				},
			}, tooMany, "should have 100 files now")
		}
	})
}

func TestJobRepository_GetStaleAccounts(t *testing.T) {
	t.Run("subscribed account", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		{ // Before updating the subscription, sh
			result, err := jobRepo.GetStaleAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving stale accounts")
			assert.Empty(t, result, "there should not be any stale accounts at the moment")
		}

		account := user.Account
		account.SubscriptionActiveUntil = myownsanity.Pointer(clock.Now())
		testutils.MustDBUpdate(t, account)

		{ // After gaining a subscription, we should still not remove the account.
			result, err := jobRepo.GetStaleAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving stale accounts")
			assert.Empty(t, result, "there should not be any stale accounts at the moment")
		}

		// But if the subscription has not been updated in 101 days, then we should.
		clock.Add(101 * 24 * time.Hour)

		{ // The subscription is now old enough to be considered stale
			result, err := jobRepo.GetStaleAccounts(context.Background())
			assert.NoError(t, err)
			assert.Len(t, result, 1, "should have the one stale now")
			assert.EqualValues(t, user.AccountId, result[0].AccountId, "should contain our account as stale")
		}
	})

	t.Run("trial account", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db, clock)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// Update the account to be the same as it would be in a trial state.
		account := user.Account
		account.SubscriptionActiveUntil = nil
		account.SubscriptionStatus = nil
		account.StripeCustomerId = nil
		account.StripeSubscriptionId = nil
		account.TrialEndsAt = myownsanity.Pointer(clock.Now().AddDate(0, 0, 30))
		testutils.MustDBUpdate(t, account)

		{ // Then check to make sure that we don't consider this a stale account.
			result, err := jobRepo.GetStaleAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving stale accounts")
			assert.Empty(t, result, "there should not be any stale accounts at the moment")
		}

		// Move the clock forward 31 days. The trial has now expired.
		clock.Add(31 * 24 * time.Hour)
		{ // But the account is not yet eligible for removal.
			result, err := jobRepo.GetStaleAccounts(context.Background())
			assert.NoError(t, err, "should not have an error retrieving stale accounts")
			assert.Empty(t, result, "there should not be any stale accounts at the moment")
		}

		// Move the clock forward 101 days, now the account should be eligible.
		clock.Add(101 * 24 * time.Hour)

		{ // 131 days after signup the account is now eligible for removal.
			result, err := jobRepo.GetStaleAccounts(context.Background())
			assert.NoError(t, err)
			assert.Len(t, result, 1, "should have the one stale now")
			assert.EqualValues(t, user.AccountId, result[0].AccountId, "should contain our account as stale")
		}
	})
}
