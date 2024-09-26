package background_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeactivateLinksHandler_EnqueueTriggeredJob(t *testing.T) {
	t.Run("plaid link is not old", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		handler := background.NewDeactivateLinksHandler(
			log,
			db,
			clock,
			config.Configuration{
				Stripe: config.Stripe{
					Enabled: true,
				},
			},
			kms,
			nil,
		)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		assert.Nil(t, plaidLink.DeletedAt, "deleted at should be nil")

		// Even a bit more than a month later we do not want to deactivate links
		clock.Add(32 * 24 * time.Hour)

		// Make sure that if our plaid link is not old then we will not deactivate it
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.DeactivateLinks),
				gomock.Any(),
			).
			Times(0)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "enqueue triggered job should not return an error")
	})

	t.Run("deactivate after 90 days", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		handler := background.NewDeactivateLinksHandler(
			log,
			db,
			clock,
			config.Configuration{
				Stripe: config.Stripe{
					Enabled: true,
				},
			},
			kms,
			nil,
		)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		assert.Nil(t, plaidLink.DeletedAt, "deleted at should be nil")

		// Make sure that if our plaid link is not old then we will not deactivate it
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.DeactivateLinks),
				testutils.NewGenericMatcher(func(args background.DeactivateLinksArguments) bool {
					return assert.EqualValues(t, user.AccountId, args.AccountId, "account ID should match the account we created")
				}),
			).
			Times(1)

		// Move time forward 90 days and 1 hour
		clock.Add(90*24*time.Hour + 1*time.Hour)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "enqueue triggered job should not return an error")
	})

	t.Run("dont do anything if billing is disabled", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		handler := background.NewDeactivateLinksHandler(
			log,
			db,
			clock,
			config.Configuration{
				Stripe: config.Stripe{
					Enabled: false,
				},
			},
			kms,
			nil,
		)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		assert.Nil(t, plaidLink.DeletedAt, "deleted at should be nil")

		// Make sure that if our plaid link is not old then we will not deactivate it
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.DeactivateLinks),
				gomock.Any(),
			).
			Times(0)

		// Move time forward 90 days and 1 hour
		clock.Add(90*24*time.Hour + 1*time.Hour)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "enqueue triggered job should not return an error")
	})
}
