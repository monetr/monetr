package background_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSyncPlaidAccountsHandler_DefaultSchedule(t *testing.T) {
	t.Run("validate cron schedule", func(t *testing.T) {
		handler := &background.SyncPlaidAccountsHandler{}
		schedule, err := cron.Parse(handler.DefaultSchedule())
		assert.NoError(t, err, "must be able too parse the schedule")
		now := time.Now()
		next := schedule.Next(now)
		assert.GreaterOrEqual(t, next, now, "next cron should always be greater or equal than now")
	})
}

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

func TestSyncPlaidAccountsJob_Run(t *testing.T) {
	t.Run("deactivate an account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		checkingAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		savingsAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.SavingsBankAccountSubType,
		)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
				gomock.Len(0), // No account IDs should be passed
			).
			Return(
				[]platypus.BankAccount{
					// Since we only return the checking account, the savings account should
					// get deactivated.
					testutils.Must(t, platypus.NewPlaidBankAccount, plaid.AccountBase{
						AccountId: checkingAccount.PlaidBankAccount.PlaidId,
					}),
				},
				nil,
			).
			Times(1)

		job, err := background.NewSyncPlaidAccountsJob(
			log,
			repository.NewRepositoryFromSession(
				clock,
				"user_system",
				user.AccountId,
				db,
				log,
			),
			clock,
			repository.NewSecretsRepository(log, clock, db, kms, user.AccountId),
			plaidPlatypus,
			background.SyncPlaidAccountsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, job)

		{ // Checking account and savings account should be active before the job
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.ActiveBankAccountStatus)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.ActiveBankAccountStatus)
		}

		// Then we run the job
		err = job.Run(context.Background())
		assert.NoError(t, err)

		{ // Now the checking is active but the savings is inactive
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.ActiveBankAccountStatus)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.InactiveBankAccountStatus)
		}
	})

	t.Run("should reactivate an account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		checkingAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		savingsAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.SavingsBankAccountSubType,
		)

		{ // Update the savings account to have an inactive status.
			savingsAccount.Status = models.InactiveBankAccountStatus
			testutils.MustDBUpdate(t, &savingsAccount)
		}

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
				gomock.Len(0), // No account IDs should be passed
			).
			Return(
				[]platypus.BankAccount{
					testutils.Must(t, platypus.NewPlaidBankAccount, plaid.AccountBase{
						AccountId: checkingAccount.PlaidBankAccount.PlaidId,
					}),
					testutils.Must(t, platypus.NewPlaidBankAccount, plaid.AccountBase{
						AccountId: savingsAccount.PlaidBankAccount.PlaidId,
					}),
				},
				nil,
			).
			Times(1)

		job, err := background.NewSyncPlaidAccountsJob(
			log,
			repository.NewRepositoryFromSession(
				clock,
				"user_system",
				user.AccountId,
				db,
				log,
			),
			clock,
			repository.NewSecretsRepository(log, clock, db, kms, user.AccountId),
			plaidPlatypus,
			background.SyncPlaidAccountsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, job)

		{ // Only checking should be active befre the job runs
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.ActiveBankAccountStatus)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.InactiveBankAccountStatus)
		}

		// Then we run the job
		err = job.Run(context.Background())
		assert.NoError(t, err)

		{ // Now both accounts should be active
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.ActiveBankAccountStatus)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.ActiveBankAccountStatus)
		}
	})
}
