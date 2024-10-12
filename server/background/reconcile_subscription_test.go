package background_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReconcileSubscriptionHandler_EnqueueTriggeredJob(t *testing.T) {
	t.Run("detect stale subscription", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		memoryCache := testutils.GetCache(t)

		accountRepo := repository.NewAccountRepository(log, memoryCache, db)
		stripeHelper := stripe_helper.NewStripeHelper(log, gofakeit.UUID())
		pubSub := pubsub.NewPostgresPubSub(log, db)
		conf := testutils.GetConfig(t)
		bill := billing.NewBilling(log, clock, conf, accountRepo, stripeHelper, pubSub)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		hasSubscription, err := bill.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "fixture account should have a subscription by default")

		// Move the clock forward 30 days, all subscriptions should be stale now.
		clock.Add(30 * 24 * time.Hour)

		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)
		handler := background.NewReconcileSubscriptionHandler(
			log, db, clock, pubSub, bill,
		)

		// Now we want to actually trigger the handler, and see if it enqueus the
		// job we want.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.ReconcileSubscription),
				testutils.NewGenericMatcher(func(args background.ReconcileSubscriptionArguments) bool {
					return assert.EqualValues(t, user.AccountId, args.AccountId, "account ID for reconcile should match the account we setup")
				}),
			).
			Times(1).
			Return(nil)

		err = handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)
	})

	t.Run("no stale subscriptions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		memoryCache := testutils.GetCache(t)

		accountRepo := repository.NewAccountRepository(log, memoryCache, db)
		stripeHelper := stripe_helper.NewStripeHelper(log, gofakeit.UUID())
		pubSub := pubsub.NewPostgresPubSub(log, db)
		conf := testutils.GetConfig(t)
		bill := billing.NewBilling(log, clock, conf, accountRepo, stripeHelper, pubSub)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		hasSubscription, err := bill.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "fixture account should have a subscription by default")

		// Same thing this time but do not move the clock forward. This way the
		// subscription should still be viewed as active.
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)
		handler := background.NewReconcileSubscriptionHandler(
			log, db, clock, pubSub, bill,
		)

		// Now we want to trigger the handler and make sure that it does not enqueue any
		// jobs since there are no stale subscriptions.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0).
			Return(nil)

		err = handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)
	})
}
