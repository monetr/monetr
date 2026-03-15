package plaid_jobs_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/datasources/plaid/plaid_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSyncPlaidAccountsCron(t *testing.T) {
	t.Run("will sync accounts who have never been synced", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		enqueuer := mockgen.NewMockProcessor(ctrl)

		// Make sure that we trigger a sync job for the link that hasn't been
		// updated before.
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.SyncPlaidAccounts),
				gomock.Any(),
				gomock.Eq(plaid_jobs.SyncPlaidAccountsArguments{
					AccountId: plaidLink.AccountId,
					LinkId:    plaidLink.LinkId,
				}),
			).
			Return(nil).
			Times(1)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			err := plaid_jobs.SyncPlaidAccountsCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err, "should not return an error")
		}
	})

	t.Run("will not sync accounts that have been synced", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		{ // Set a timestamp on the last account sync and store it.
			plaidLink.PlaidLink.LastAccountSync = myownsanity.Pointer(clock.Now())
			testutils.MustDBUpdate(t, plaidLink.PlaidLink)
		}

		enqueuer := mockgen.NewMockProcessor(ctrl)

		// Make sure that we trigger a sync job for the link that hasn't been
		// updated before.
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.SyncPlaidAccounts),
				gomock.Any(),
				gomock.Eq(plaid_jobs.SyncPlaidAccountsArguments{
					AccountId: plaidLink.AccountId,
					LinkId:    plaidLink.LinkId,
				}),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).MinTimes(1)
			err := plaid_jobs.SyncPlaidAccountsCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err, "should not return an error")
		}
	})
}
