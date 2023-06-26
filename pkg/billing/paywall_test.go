package billing

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
)

func TestBaseBasicPaywall_GetHasSubscription(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)
		memoryCache := testutils.GetCache(t)
		log := testutils.GetLog(t)

		accountRepo := NewAccountRepository(log, memoryCache, db)

		paywall := NewBasicPaywall(log, accountRepo)

		user, _ := fixtures.GivenIHaveABasicAccount(t)

		hasSubscription, err := paywall.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "fixture account should have a subscription by default")

		account := user.Account
		canceledStatus := stripe.SubscriptionStatusCanceled
		account.SubscriptionActiveUntil = myownsanity.TimeP(time.Now().Add(-1 * time.Hour))
		account.SubscriptionStatus = &canceledStatus

		err = accountRepo.UpdateAccount(context.Background(), account)
		assert.NoError(t, err, "failed to update account with new status")

		hasSubscription, err = paywall.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.False(t, hasSubscription, "account should no longer have a subscription")
	})

	t.Run("payment past due", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)
		memoryCache := testutils.GetCache(t)
		log := testutils.GetLog(t)

		accountRepo := NewAccountRepository(log, memoryCache, db)

		paywall := NewBasicPaywall(log, accountRepo)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		account := user.Account
		subscriptionStatus := stripe.SubscriptionStatusPastDue
		account.SubscriptionActiveUntil = myownsanity.TimeP(time.Now().Add(7 * 24 * time.Hour))
		account.SubscriptionStatus = &subscriptionStatus

		err := accountRepo.UpdateAccount(context.Background(), account)
		assert.NoError(t, err, "failed to update account with new status")

		hasSubscription, err := paywall.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "the subscription should be present for past due")
	})

	t.Run("subscription is canceled", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)
		memoryCache := testutils.GetCache(t)
		log := testutils.GetLog(t)

		accountRepo := NewAccountRepository(log, memoryCache, db)

		paywall := NewBasicPaywall(log, accountRepo)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		account := user.Account
		subscriptionStatus := stripe.SubscriptionStatusCanceled
		account.SubscriptionActiveUntil = myownsanity.TimeP(time.Now().Add(-7 * 24 * time.Hour))
		account.SubscriptionStatus = &subscriptionStatus

		err := accountRepo.UpdateAccount(context.Background(), account)
		assert.NoError(t, err, "failed to update account with new status")

		hasSubscription, err := paywall.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.False(t, hasSubscription, "the subscription is not present when canceled")
	})

	t.Run("status is nil", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)
		memoryCache := testutils.GetCache(t)
		log := testutils.GetLog(t)

		accountRepo := NewAccountRepository(log, memoryCache, db)

		paywall := NewBasicPaywall(log, accountRepo)

		user, _ := fixtures.GivenIHaveABasicAccount(t)

		hasSubscription, err := paywall.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "fixture account should have a subscription by default")

		account := user.Account
		account.SubscriptionStatus = nil

		err = accountRepo.UpdateAccount(context.Background(), account)
		assert.NoError(t, err, "failed to update account with new status")

		hasSubscription, err = paywall.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "the subscription should be present for past due")
	})

	t.Run("account not found", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)
		memoryCache := testutils.GetCache(t)
		log := testutils.GetLog(t)

		accountRepo := NewAccountRepository(log, memoryCache, db)

		paywall := NewBasicPaywall(log, accountRepo)

		hasSubscription, err := paywall.GetHasSubscription(context.Background(), math.MaxUint64)
		assert.EqualError(t, err, "could not determine whether subscription was present: failed to retrieve account by Id: pg: no rows in result set")
		assert.False(t, hasSubscription, "account that does not exist should return false")
	})
}
