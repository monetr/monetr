package billing_jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/billing/billing_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
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

		enqueuer := mockgen.NewMockProcessor(ctrl)

		// Now we want to actually trigger the handler, and see if it enqueus the
		// job we want.
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(billing_jobs.ReconcileSubscription),
				gomock.Any(),
				gomock.Eq(billing_jobs.ReconcileSubscriptionArguments{
					AccountId: user.AccountId,
				}),
			).
			Return(nil).
			Times(1)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			err := billing_jobs.ReconcileSubscriptionCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("no stale subscriptions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		memoryCache := testutils.GetCache(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		accountRepo := repository.NewAccountRepository(log, memoryCache, db)
		stripeHelper := stripe_helper.NewStripeHelper(log, gofakeit.UUID())
		pubSub := pubsub.NewPostgresPubSub(log, db)
		conf := testutils.GetConfig(t)
		bill := billing.NewBilling(log, clock, conf, accountRepo, stripeHelper, pubSub)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		hasSubscription, err := bill.GetHasSubscription(context.Background(), user.AccountId)
		assert.NoError(t, err, "must not return an error checking for subscription")
		assert.True(t, hasSubscription, "fixture account should have a subscription by default")

		// Now we want to trigger the handler and make sure that it does not enqueue any
		// jobs since there are no stale subscriptions.
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(billing_jobs.ReconcileSubscription),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		// Same thing this time but do not move the clock forward. This way the
		// subscription should still be viewed as active.
		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			err := billing_jobs.ReconcileSubscriptionCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})
}
