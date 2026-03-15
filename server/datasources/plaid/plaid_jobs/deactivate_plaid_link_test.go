package plaid_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/datasources/plaid/plaid_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeactivateLinksCron(t *testing.T) {
	t.Run("plaid link is not old", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		assert.Nil(t, plaidLink.DeletedAt, "deleted at should be nil")

		// Even a bit more than a month later we do not want to deactivate links
		clock.Add(32 * 24 * time.Hour)

		// Make sure that if our plaid link is not old then we will not deactivate it
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.DeactivatePlaidLink),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Configuration().Return(config.Configuration{
				Stripe: config.Stripe{
					Enabled: true,
				},
			}).Times(1)
			err := plaid_jobs.DeactivatePlaidLinkCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err, "should not return an error")
		}
	})

	t.Run("deactivate after 90 days", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		assert.Nil(t, plaidLink.DeletedAt, "deleted at should be nil")

		// Make sure that if our plaid link is not old then we will not deactivate it
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.DeactivatePlaidLink),
				gomock.Any(),
				gomock.Eq(plaid_jobs.DeactivateLinksArguments{
					AccountId: plaidLink.AccountId,
					LinkId:    plaidLink.LinkId,
				}),
			).
			Return(nil).
			Times(1)

		// Move time forward 90 days and 1 hour
		clock.Add(90*24*time.Hour + 1*time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Configuration().Return(config.Configuration{
				Stripe: config.Stripe{
					Enabled: true,
				},
			}).Times(1)
			err := plaid_jobs.DeactivatePlaidLinkCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err, "should not return an error")
		}
	})

	t.Run("dont do anything if billing is disabled", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		assert.Nil(t, plaidLink.DeletedAt, "deleted at should be nil")

		// Make sure that if our plaid link is not old then we will not deactivate it
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.DeactivatePlaidLink),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		// Move time forward 90 days and 1 hour
		clock.Add(90*24*time.Hour + 1*time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).Times(0)
			context.EXPECT().DB().Return(db).Times(0)
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Configuration().Return(config.Configuration{
				Stripe: config.Stripe{
					Enabled: false,
				},
			}).Times(1)
			err := plaid_jobs.DeactivatePlaidLinkCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err, "should not return an error")
		}
	})
}
