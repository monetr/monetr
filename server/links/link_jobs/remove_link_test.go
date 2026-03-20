package link_jobs_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/links/link_jobs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRemoveLink(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bankAccount, 100)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    bankAccount.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			for i := range transactions {
				transaction := transactions[i]
				testutils.MustDBNotExist(t, *transaction.PlaidTransaction)
				testutils.MustDBNotExist(t, transaction)
			}

			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
			testutils.MustDBNotExist(t, models.Secret{
				SecretId:  link.PlaidLink.SecretId,
				AccountId: link.AccountId,
			})
		}
	})

	t.Run("no transactions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    bankAccount.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("no bank accounts", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure link data has been removed
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
			testutils.MustDBNotExist(t, models.Secret{
				SecretId:  link.PlaidLink.SecretId,
				AccountId: link.AccountId,
			})
		}
	})

	t.Run("manual link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bankAccount, 10)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			for i := range transactions {
				testutils.MustDBNotExist(t, transactions[i])
			}
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("with spending and funding schedules", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bankAccount, 10)

		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(
			t,
			clock,
			&bankAccount,
			"FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15",
			false,
		)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		rule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Test Spending",
			TargetAmount:           10000,
			CurrentAmount:          5000,
			NextRecurrence:         rule.After(clock.Now(), false),
			NextContributionAmount: 5000,
			RuleSet:                rule,
			CreatedAt:              clock.Now(),
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			for i := range transactions {
				testutils.MustDBNotExist(t, *transactions[i].PlaidTransaction)
				testutils.MustDBNotExist(t, transactions[i])
			}
			testutils.MustDBNotExist(t, spending)
			testutils.MustDBNotExist(t, *fundingSchedule)
			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("with transaction clusters and uploads", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		cluster := testutils.MustInsert(t, models.TransactionCluster{
			AccountId:     bankAccount.AccountId,
			BankAccountId: bankAccount.BankAccountId,
			Name:          "Test Cluster",
			OriginalName:  "test cluster",
			Members:       []models.ID[models.Transaction]{},
		})

		file := testutils.MustInsert(t, models.File{
			AccountId:   bankAccount.AccountId,
			Kind:        "transactions/uploads",
			Name:        "test-upload.csv",
			ContentType: models.TextCSVContentType,
			Size:        1024,
			CreatedBy:   link.CreatedBy,
		})
		upload := testutils.MustInsert(t, models.TransactionUpload{
			AccountId:     bankAccount.AccountId,
			BankAccountId: bankAccount.BankAccountId,
			FileId:        file.FileId,
			Status:        models.TransactionUploadStatusComplete,
			CreatedBy:     link.CreatedBy,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			testutils.MustDBNotExist(t, cluster)
			testutils.MustDBNotExist(t, upload)
			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("with plaid syncs", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		plaidSync := testutils.MustInsert(t, models.PlaidSync{
			AccountId:   bankAccount.AccountId,
			PlaidLinkId: link.PlaidLink.PlaidLinkId,
			Timestamp:   clock.Now(),
			Trigger:     "webhook",
			NextCursor:  "cursor-abc123",
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			testutils.MustDBNotExist(t, plaidSync)
			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("lunch flow link with transactions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		lfTxn1 := testutils.MustInsert(t, models.LunchFlowTransaction{
			AccountId:              bankAccount.AccountId,
			LunchFlowBankAccountId: bankAccount.LunchFlowBankAccount.LunchFlowBankAccountId,
			LunchFlowId:            "lf-txn-1",
			Merchant:               "Coffee Shop",
			Description:            "Morning coffee",
			Date:                   clock.Now(),
			Currency:               "USD",
			Amount:                 450,
			IsPending:              false,
		})
		lfTxn2 := testutils.MustInsert(t, models.LunchFlowTransaction{
			AccountId:              bankAccount.AccountId,
			LunchFlowBankAccountId: bankAccount.LunchFlowBankAccount.LunchFlowBankAccountId,
			LunchFlowId:            "lf-txn-2",
			Merchant:               "Grocery Store",
			Description:            "Weekly groceries",
			Date:                   clock.Now(),
			Currency:               "USD",
			Amount:                 3500,
			IsPending:              false,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			testutils.MustDBNotExist(t, lfTxn1)
			testutils.MustDBNotExist(t, lfTxn2)
			testutils.MustDBNotExist(t, *bankAccount.LunchFlowBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.LunchFlowLink)
			testutils.MustDBNotExist(t, link)
			testutils.MustDBNotExist(t, models.Secret{
				SecretId:  link.LunchFlowLink.SecretId,
				AccountId: link.AccountId,
			})
		}
	})

	t.Run("link not found", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    models.NewID[models.Link](),
				},
			)
			assert.NoError(t, err, "remove link job should succeed even when link not found")
		})
	})

	t.Run("multiple bank accounts", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount1 := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		bankAccount2 := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.SavingsBankAccountSubType,
		)
		transactions1 := fixtures.GivenIHaveNTransactions(t, clock, bankAccount1, 10)
		transactions2 := fixtures.GivenIHaveNTransactions(t, clock, bankAccount2, 10)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed for both bank accounts
			for i := range transactions1 {
				testutils.MustDBNotExist(t, *transactions1[i].PlaidTransaction)
				testutils.MustDBNotExist(t, transactions1[i])
			}
			for i := range transactions2 {
				testutils.MustDBNotExist(t, *transactions2[i].PlaidTransaction)
				testutils.MustDBNotExist(t, transactions2[i])
			}

			testutils.MustDBNotExist(t, *bankAccount1.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount1)
			testutils.MustDBNotExist(t, *bankAccount2.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount2)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("does not affect other links", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link1 := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		link2 := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount1 := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link1,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		bankAccount2 := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link2,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions1 := fixtures.GivenIHaveNTransactions(t, clock, bankAccount1, 5)
		transactions2 := fixtures.GivenIHaveNTransactions(t, clock, bankAccount2, 5)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    link1.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // link1 data should be removed
			for i := range transactions1 {
				testutils.MustDBNotExist(t, *transactions1[i].PlaidTransaction)
				testutils.MustDBNotExist(t, transactions1[i])
			}
			testutils.MustDBNotExist(t, *bankAccount1.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount1)
			testutils.MustDBNotExist(t, *link1.PlaidLink)
			testutils.MustDBNotExist(t, link1)
		}

		{ // link2 data should still exist
			for i := range transactions2 {
				testutils.MustDBExist(t, *transactions2[i].PlaidTransaction)
				testutils.MustDBExist(t, transactions2[i])
			}
			testutils.MustDBExist(t, *bankAccount2.PlaidBankAccount)
			testutils.MustDBExist(t, bankAccount2)
			testutils.MustDBExist(t, *link2.PlaidLink)
			testutils.MustDBExist(t, link2)
		}
	})

	t.Run("does not affect other link types", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		lfLink := fixtures.GivenIHaveALunchFlowLink(t, clock, user)

		plaidBankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		lfBankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &lfLink)

		plaidTransactions := fixtures.GivenIHaveNTransactions(t, clock, plaidBankAccount, 5)
		lfTxn := testutils.MustInsert(t, models.LunchFlowTransaction{
			AccountId:              lfBankAccount.AccountId,
			LunchFlowBankAccountId: lfBankAccount.LunchFlowBankAccount.LunchFlowBankAccountId,
			LunchFlowId:            "lf-isolation-1",
			Merchant:               "Test Store",
			Description:            "Test purchase",
			Date:                   clock.Now(),
			Currency:               "USD",
			Amount:                 1000,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Plaid link data should be removed
			for i := range plaidTransactions {
				testutils.MustDBNotExist(t, *plaidTransactions[i].PlaidTransaction)
				testutils.MustDBNotExist(t, plaidTransactions[i])
			}
			testutils.MustDBNotExist(t, *plaidBankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, plaidBankAccount)
			testutils.MustDBNotExist(t, *plaidLink.PlaidLink)
			testutils.MustDBNotExist(t, plaidLink)
		}

		{ // LunchFlow link data should still exist
			testutils.MustDBExist(t, lfTxn)
			testutils.MustDBExist(t, *lfBankAccount.LunchFlowBankAccount)
			testutils.MustDBExist(t, lfBankAccount)
			testutils.MustDBExist(t, *lfLink.LunchFlowLink)
			testutils.MustDBExist(t, lfLink)
		}
	})
}
