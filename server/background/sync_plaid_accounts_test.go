package background_test

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSyncPlaidAccountsHandler_EnqueueTriggeredJob(t *testing.T) {
	t.Run("will sync accounts who have never been synced", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		// Make sure that we trigger a sync job for the link that hasn't been
		// updated before.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.SyncPlaidAccounts),
				testutils.NewGenericMatcher(func(args background.SyncPlaidAccountsArguments) bool {
					return assert.EqualValues(t, plaidLink.LinkId, args.LinkId) &&
						assert.EqualValues(t, plaidLink.AccountId, args.AccountId)
				}),
			).
			Times(1).
			Return(nil)

		handler := background.NewSyncPlaidAccountsHandler(
			log,
			db,
			clock,
			kms,
			plaidPlatypus,
		)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")
	})

	t.Run("will not sync accounts that have been synced", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

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
			plaidLink.PlaidLink.LastAccountSync = myownsanity.TimeP(clock.Now())
			testutils.MustDBUpdate(t, plaidLink.PlaidLink)
		}

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		// Make sure that we trigger a sync job for the link that hasn't been
		// updated before.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.SyncPlaidAccounts),
				testutils.NewGenericMatcher(func(args background.SyncPlaidAccountsArguments) bool {
					return assert.EqualValues(t, plaidLink.LinkId, args.LinkId) &&
						assert.EqualValues(t, plaidLink.AccountId, args.AccountId)
				}),
			).
			Times(0).
			Return(nil)

		handler := background.NewSyncPlaidAccountsHandler(
			log,
			db,
			clock,
			kms,
			plaidPlatypus,
		)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")
	})
}

