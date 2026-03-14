package background_test

import (
	"fmt"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_lunch_flow"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSyncLunchFlowJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		handler := background.NewSyncLunchFlowHandler(
			log,
			db,
			clock,
			kms,
			publisher,
			enqueuer,
		)

		firstTransactions := []lunch_flow.Transaction{
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "19.99",
				Currency:    "USD",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Taco Bell",
				Description: "POS DEBIT - 1234 TACO BELL STPAUL MN",
			},
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
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
				AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "500.00",
				Currency:    "USD",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Car Repair",
				Description: "ACH DEBIT - CAR REPAIR",
			},
		}

		firstCalculateCall := enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.CalculateTransactionClustersName),
				testutils.NewGenericMatcher(func(args background.CalculateTransactionClustersArguments) bool {
					return myownsanity.Every(
						assert.Equal(t, bankAccount.BankAccountId, args.BankAccountId),
						assert.Equal(t, bankAccount.AccountId, args.AccountId),
					)
				}),
			).
			Times(1).
			Return(nil)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			args := background.SyncLunchFlowArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
				LinkId:        bankAccount.LinkId,
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				firstTransactions,
			)

			mock_lunch_flow.MockFetchBalance(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				lunch_flow.Balance{
					Amount:   "1234.56",
					Currency: "USD",
				},
			)

			err = handler.HandleConsumeJob(t.Context(), log, argsEncoded)
			assert.NoError(t, err, "must process job successfully")

			// We should have a few transactions now.
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 2, count, "should have one transaction now!")

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/balance", bankAccount.LunchFlowBankAccount.LunchFlowId):      1,
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.CalculateTransactionClustersName),
				testutils.NewGenericMatcher(func(args background.CalculateTransactionClustersArguments) bool {
					return myownsanity.Every(
						assert.Equal(t, bankAccount.BankAccountId, args.BankAccountId),
						assert.Equal(t, bankAccount.AccountId, args.AccountId),
					)
				}),
			).
			Times(1).
			After(firstCalculateCall).
			Return(nil)

		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			args := background.SyncLunchFlowArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
				LinkId:        bankAccount.LinkId,
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				// Now build a new array that is the first transactions appended to the
				// end of the second transactions.
				append(secondTransactions, firstTransactions...),
			)

			mock_lunch_flow.MockFetchBalance(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				lunch_flow.Balance{
					Amount:   "5000.12",
					Currency: "USD",
				},
			)

			err = handler.HandleConsumeJob(t.Context(), log, argsEncoded)
			assert.NoError(t, err, "must process job successfully")

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
		kms := secrets.NewPlaintextKMS()

		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		// Make sure we are in a japanese locale for this test
		user.Account.Locale = "ja_JP"
		testutils.MustDBUpdate(t, user.Account)

		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)
		bankAccount.Currency = "JPY"
		testutils.MustDBUpdate(t, &bankAccount)

		handler := background.NewSyncLunchFlowHandler(
			log,
			db,
			clock,
			kms,
			publisher,
			enqueuer,
		)

		firstTransactions := []lunch_flow.Transaction{
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "1900.00",
				Currency:    "JPY",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Taco Bell",
				Description: "POS DEBIT - 1234 TACO BELL STPAUL MN",
			},
			{
				Id:          gofakeit.UUID(),
				AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				Amount:      "200000.00",
				Currency:    "JPY",
				Date:        clock.Now().Format(lunch_flow.DateFormat),
				Merchant:    "Rocket Mortgage",
				Description: "ACH DEBIT - ROCKET MORTGAGE 1234567890",
			},
		}

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.CalculateTransactionClustersName),
				testutils.NewGenericMatcher(func(args background.CalculateTransactionClustersArguments) bool {
					return myownsanity.Every(
						assert.Equal(t, bankAccount.BankAccountId, args.BankAccountId),
						assert.Equal(t, bankAccount.AccountId, args.AccountId),
					)
				}),
			).
			Times(1).
			Return(nil)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			args := background.SyncLunchFlowArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
				LinkId:        bankAccount.LinkId,
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				firstTransactions,
			)

			mock_lunch_flow.MockFetchBalance(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				lunch_flow.Balance{
					Amount:   "1234.00",
					Currency: "JPY",
				},
			)

			err = handler.HandleConsumeJob(t.Context(), log, argsEncoded)
			assert.NoError(t, err, "must process job successfully")

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
		kms := secrets.NewPlaintextKMS()

		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		handler := background.NewSyncLunchFlowHandler(
			log,
			db,
			clock,
			kms,
			publisher,
			enqueuer,
		)

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.CalculateTransactionClustersName),
				gomock.Any(),
			).
			Times(0)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			args := background.SyncLunchFlowArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
				LinkId:        bankAccount.LinkId,
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactionsError(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
			)

			err = handler.HandleConsumeJob(t.Context(), log, argsEncoded)
			assert.Error(t, err, "job will fail if an API call fails")

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

		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveALunchFlowBankAccount(t, clock, &link)

		handler := background.NewSyncLunchFlowHandler(
			log,
			db,
			clock,
			kms,
			publisher,
			enqueuer,
		)

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.CalculateTransactionClustersName),
				gomock.Any(),
			).
			Times(0)

		// Do the first sync in an anonymous function so the http mocks reset
		// properly. I'm not doing this in a t.Run because I don't want the sub
		// function to be invoked on its own because I'm going to use multiple.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			args := background.SyncLunchFlowArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
				LinkId:        bankAccount.LinkId,
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			mock_lunch_flow.MockFetchTransactions(
				t,
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
				[]lunch_flow.Transaction{
					{
						Id:          gofakeit.UUID(),
						AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
						Amount:      "19.99",
						Currency:    "USD",
						Date:        clock.Now().Format(lunch_flow.DateFormat),
						Merchant:    "Taco Bell",
						Description: "POS DEBIT - 1234 TACO BELL STPAUL MN",
					},
					{
						Id:          gofakeit.UUID(),
						AccountId:   lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
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
				lunch_flow.AccountId(bankAccount.LunchFlowBankAccount.LunchFlowId),
			)

			err = handler.HandleConsumeJob(t.Context(), log, argsEncoded)
			assert.Error(t, err, "job will fail if an API call fails")

			// If it fails we should not create any transactions
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/balance", bankAccount.LunchFlowBankAccount.LunchFlowId):      1,
				fmt.Sprintf("GET https://lunchflow.app/api/v1/accounts/%s/transactions", bankAccount.LunchFlowBankAccount.LunchFlowId): 1,
			}, "must match Lunch Flow API calls")
		}()
	})
}
