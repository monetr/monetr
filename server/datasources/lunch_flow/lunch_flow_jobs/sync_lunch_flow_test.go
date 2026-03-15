package lunch_flow_jobs_test

import (
	"fmt"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/datasources/lunch_flow/lunch_flow_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_lunch_flow"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/similar/similar_jobs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSyncLunchFlow(t *testing.T) {
	t.Run("lunch flow not enabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			LunchFlow: config.LunchFlow{
				Enabled: false,
			},
		}).MinTimes(1)

		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			err := lunch_flow_jobs.SyncLunchFlow(
				mockqueue.NewMockContext(context),
				lunch_flow_jobs.SyncLunchFlowArguments{
					AccountId:     "acct_bogus",
					BankAccountId: "bac_bogus",
					LinkId:        "link_bogus",
				},
			)
			assert.NoError(t, err, "should return without error when lunch flow is disabled")
			assert.Empty(t, httpmock.GetCallCountInfo(), "must not make any lunch flow API calls when disabled")
		}()

	})

	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		enqueuer := mockgen.NewMockProcessor(ctrl)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		firstTransactions := []lunch_flow.Transaction{
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "19.99",
				Currency:    "USD",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Taco Bell",
				Description: "POS DEBIT - 1234 TACO BELL STPAUL MN",
			},
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "2000.00",
				Currency:    "USD",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Rocket Mortgage",
				Description: "ACH DEBIT - ROCKET MORTGAGE 1234567890",
			},
		}

		secondTransactions := []lunch_flow.Transaction{
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "500.00",
				Currency:    "USD",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Car Repair",
				Description: "ACH DEBIT - CAR REPAIR",
			},
		}

		firstCalculateCall := enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				firstTransactions,
			)

			mock_lunch_flow.MockFetchBalance(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				lunch_flow.Balance{
					Amount:   "1234.56",
					Currency: "USD",
				},
			)

			{
				context := mockgen.NewMockContext(ctrl)
				context.EXPECT().Clock().Return(clock).AnyTimes()
				context.EXPECT().Configuration().Return(config.Configuration{
					LunchFlow: config.LunchFlow{
						Enabled: true,
					},
				}).AnyTimes()
				context.EXPECT().KMS().Return(kms).AnyTimes()
				context.EXPECT().DB().Return(db).AnyTimes()
				context.EXPECT().Log().Return(log).AnyTimes()
				context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
				context.EXPECT().Publisher().Return(publisher).AnyTimes()
				context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
				err := lunch_flow_jobs.SyncLunchFlow(
					mockqueue.NewMockContext(context),
					lunch_flow_jobs.SyncLunchFlowArguments{
						AccountId:     bankAccount.AccountId,
						BankAccountId: bankAccount.BankAccountId,
						LinkId:        bankAccount.LinkId,
					},
				)
				assert.NoError(t, err, "must sync lunch flow successfully")
			}

			// We should have a few transactions now.
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 2, count, "should have one transaction now!")

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/balance", bankAccount.LunchFlowBankAccount.LunchFlowId):      1,
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
				}),
			).
			Times(1).
			After(firstCalculateCall).
			Return(nil)

		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				// Now build a new array that is the first transactions appended to the
				// end of the second transactions.
				append(secondTransactions, firstTransactions...),
			)

			mock_lunch_flow.MockFetchBalance(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				lunch_flow.Balance{
					Amount:   "5000.12",
					Currency: "USD",
				},
			)

			{
				context := mockgen.NewMockContext(ctrl)
				context.EXPECT().Clock().Return(clock).AnyTimes()
				context.EXPECT().Configuration().Return(config.Configuration{
					LunchFlow: config.LunchFlow{
						Enabled: true,
					},
				}).AnyTimes()
				context.EXPECT().KMS().Return(kms).AnyTimes()
				context.EXPECT().DB().Return(db).AnyTimes()
				context.EXPECT().Log().Return(log).AnyTimes()
				context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
				context.EXPECT().Publisher().Return(publisher).AnyTimes()
				context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
				err := lunch_flow_jobs.SyncLunchFlow(
					mockqueue.NewMockContext(context),
					lunch_flow_jobs.SyncLunchFlowArguments{
						AccountId:     bankAccount.AccountId,
						BankAccountId: bankAccount.BankAccountId,
						LinkId:        bankAccount.LinkId,
					},
				)
				assert.NoError(t, err, "must sync lunch flow successfully")
			}

			// We should have a few transactions now.
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 3, count, "should have one transaction now!")

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/balance", bankAccount.LunchFlowBankAccount.LunchFlowId):      1,
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()
	})

	t.Run("happy path but with JPY currency", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		enqueuer := mockgen.NewMockProcessor(ctrl)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		// Make sure we are in a japanese locale for this test
		user.Account.Locale = "ja_JP"
		testutils.MustDBUpdate(t, user.Account)

		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)
		bankAccount.Currency = "JPY"
		testutils.MustDBUpdate(t, &bankAccount)

		firstTransactions := []lunch_flow.Transaction{
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "1900.00",
				Currency:    "JPY",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Taco Bell",
				Description: "POS DEBIT - 1234 TACO BELL STPAUL MN",
			},
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "200000.00",
				Currency:    "JPY",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Rocket Mortgage",
				Description: "ACH DEBIT - ROCKET MORTGAGE 1234567890",
			},
		}

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				firstTransactions,
			)

			mock_lunch_flow.MockFetchBalance(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				lunch_flow.Balance{
					Amount:   "1234.00",
					Currency: "JPY",
				},
			)

			{
				context := mockgen.NewMockContext(ctrl)
				context.EXPECT().Clock().Return(clock).AnyTimes()
				context.EXPECT().Configuration().Return(config.Configuration{
					LunchFlow: config.LunchFlow{
						Enabled: true,
					},
				}).AnyTimes()
				context.EXPECT().KMS().Return(kms).AnyTimes()
				context.EXPECT().DB().Return(db).AnyTimes()
				context.EXPECT().Log().Return(log).AnyTimes()
				context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
				context.EXPECT().Publisher().Return(publisher).AnyTimes()
				context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
				err := lunch_flow_jobs.SyncLunchFlow(
					mockqueue.NewMockContext(context),
					lunch_flow_jobs.SyncLunchFlowArguments{
						AccountId:     bankAccount.AccountId,
						BankAccountId: bankAccount.BankAccountId,
						LinkId:        bankAccount.LinkId,
					},
				)
				assert.NoError(t, err, "must sync lunch flow successfully")
			}

			// We should have a few transactions now.
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 2, count, "should have one transaction now!")

			// Make sure that we parse the amounts properly even if they are JPY
			bankAccountUpdated := testutils.MustDBRead(t, bankAccount)
			assert.EqualValues(t, 1234, bankAccountUpdated.AvailableBalance)
			assert.EqualValues(t, 1234, bankAccountUpdated.CurrentBalance)

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/balance", bankAccount.LunchFlowBankAccount.LunchFlowId):      1,
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()
	})

	t.Run("fails on transaction request", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		enqueuer := mockgen.NewMockProcessor(ctrl)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactionsError(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
			)

			{
				context := mockgen.NewMockContext(ctrl)
				context.EXPECT().Clock().Return(clock).AnyTimes()
				context.EXPECT().Configuration().Return(config.Configuration{
					LunchFlow: config.LunchFlow{
						Enabled: true,
					},
				}).AnyTimes()
				context.EXPECT().KMS().Return(kms).AnyTimes()
				context.EXPECT().DB().Return(db).AnyTimes()
				context.EXPECT().Log().Return(log).AnyTimes()
				context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
				context.EXPECT().Publisher().Return(publisher).AnyTimes()
				context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
				err := lunch_flow_jobs.SyncLunchFlow(
					mockqueue.NewMockContext(context),
					lunch_flow_jobs.SyncLunchFlowArguments{
						AccountId:     bankAccount.AccountId,
						BankAccountId: bankAccount.BankAccountId,
						LinkId:        bankAccount.LinkId,
					},
				)
				assert.Error(t, err, "must return an error if the API call fails")
			}

			// If it fails we should not create any transactions
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()
	})

	t.Run("fails on balance request", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				[]lunch_flow.Transaction{
					{
						Id:          gofakeit.UUID(),
						AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
						Amount:      "19.99",
						Currency:    "USD",
						Date:        clock.Now().Format(lunch_flow.DateFormat),
						Merchant:    "Taco Bell",
						Description: "POS DEBIT - 1234 TACO BELL STPAUL MN",
					},
					{
						Id:          gofakeit.UUID(),
						AccountId:   lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
						Amount:      "2000.00",
						Currency:    "USD",
						Date:        clock.Now().Format(lunch_flow.DateFormat),
						Merchant:    "Rocket Mortgage",
						Description: "ACH DEBIT - ROCKET MORTGAGE 1234567890",
					},
				},
			)

			mock_lunch_flow.MockFetchBalanceError(
				t,
				lunch_flow.LunchFlowAccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
			)

			{
				context := mockgen.NewMockContext(ctrl)
				context.EXPECT().Clock().Return(clock).AnyTimes()
				context.EXPECT().Configuration().Return(config.Configuration{
					LunchFlow: config.LunchFlow{
						Enabled: true,
					},
				}).AnyTimes()
				context.EXPECT().KMS().Return(kms).AnyTimes()
				context.EXPECT().DB().Return(db).AnyTimes()
				context.EXPECT().Log().Return(log).AnyTimes()
				context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
				context.EXPECT().Publisher().Return(publisher).AnyTimes()
				context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
				err := lunch_flow_jobs.SyncLunchFlow(
					mockqueue.NewMockContext(context),
					lunch_flow_jobs.SyncLunchFlowArguments{
						AccountId:     bankAccount.AccountId,
						BankAccountId: bankAccount.BankAccountId,
						LinkId:        bankAccount.LinkId,
					},
				)
				assert.Error(t, err, "must return an error if the balance API call fails")
			}

			// If it fails we should not create any transactions
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/balance", bankAccount.LunchFlowBankAccount.LunchFlowId):      1,
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()
	})
}
