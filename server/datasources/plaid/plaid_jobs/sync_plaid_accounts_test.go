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
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/secrets"
	"github.com/plaid/plaid-go/v41/plaid"
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

func TestSyncPlaidAccounts(t *testing.T) {
	t.Run("deactivate an account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
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

		{ // Checking account and savings account should be active before the job
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.BankAccountStatusActive)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.BankAccountStatusActive)
		}

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().KMS().Return(kms).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			err := plaid_jobs.SyncPlaidAccounts(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidAccountsArguments{
					AccountId: plaidLink.AccountId,
					LinkId:    plaidLink.LinkId,
				},
			)
			assert.NoError(t, err, "must sync lunch flow successfully")
		}

		{ // Now the checking is active but the savings is inactive
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.BankAccountStatusActive)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.BankAccountStatusInactive)
		}
	})

	t.Run("should reactivate an account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
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
			savingsAccount.Status = models.BankAccountStatusInactive
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

		{ // Only checking should be active befre the job runs
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.BankAccountStatusActive)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.BankAccountStatusInactive)
		}

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().KMS().Return(kms).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			err := plaid_jobs.SyncPlaidAccounts(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidAccountsArguments{
					AccountId: plaidLink.AccountId,
					LinkId:    plaidLink.LinkId,
				},
			)
			assert.NoError(t, err, "must sync lunch flow successfully")
		}

		{ // Now both accounts should be active
			checkingAccountBefore := testutils.MustRetrieve(t, checkingAccount)
			assert.Equal(t, checkingAccountBefore.Status, models.BankAccountStatusActive)
			savingsAccountBefore := testutils.MustRetrieve(t, savingsAccount)
			assert.Equal(t, savingsAccountBefore.Status, models.BankAccountStatusActive)
		}
	})
}
