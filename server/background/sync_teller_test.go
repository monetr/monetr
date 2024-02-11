// vim: foldmethod=indent

package background_test

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockgenteller"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/teller"
	"github.com/stretchr/testify/assert"
)

func TestSyncTellerJob_Run(t *testing.T) {
	t.Run("happy initial setup", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveATellerLink(t, clock, user)
		// Tweak to pending for initial setup
		link.TellerLink.Status = models.TellerLinkStatusPending
		testutils.MustDBUpdate(t, link.TellerLink)

		client := mockgenteller.NewMockClient(ctrl)
		authenticatedClient := mockgenteller.NewMockAuthenticatedClient(ctrl)
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		handler := background.NewSyncTellerHandler(
			log,
			db,
			clock,
			kms,
			client,
			publisher,
			enqueuer,
		)

		tellerBankAccountId := gofakeit.Generate("acc_????????????????")
		transaction1Id := gofakeit.Generate("txn_????????????????")
		transaction2Id := gofakeit.Generate("txn_????????????????")
		transaction3Id := gofakeit.Generate("txn_????????????????")

		var firstEnqueueJob, secondEnqueueJob *gomock.Call

		{ // Setup the mock calls for the first sync.
			firstGetAuthenticatedClient := client.EXPECT().
				GetAuthenticatedClient(
					gomock.AssignableToTypeOf("string"),
				).
				Return(authenticatedClient).
				Times(1)

			firstGetAccounts := authenticatedClient.EXPECT().
				GetAccounts(
					gomock.Any(),
				).
				Return([]teller.Account{
					{
						Id:           tellerBankAccountId,
						Currency:     "USD",
						EnrollmentId: link.TellerLink.EnrollmentId,
						Institution: struct {
							Id   string "json:\"id\""
							Name string "json:\"name\""
						}{
							Id:   "navy_federal",
							Name: "Navy Federal",
						},
						Mask:    "1234",
						Links:   map[string]string{},
						Name:    "Primary Checking",
						Type:    teller.AccountTypeDepository,
						SubType: teller.AccountSubTypeChecking,
						Status:  teller.AccountStatusOpen,
					},
				}, nil).
				After(firstGetAuthenticatedClient).
				Times(1)

			firstGetTransactions := authenticatedClient.EXPECT().
				GetTransactions(
					gomock.Any(),
					gomock.Eq(tellerBankAccountId),
					gomock.Nil(),
					testutils.EqVal(25),
				).
				Return([]teller.Transaction{
					{
						Id:          transaction1Id,
						AccountId:   tellerBankAccountId,
						Date:        "2023-02-10",
						Description: "I am the first transaction",
						Details: teller.TransactionDetails{
							ProcessingStatus: teller.TransactionProcessingStatusComplete,
							Category:         teller.TransactionCategoryGeneral,
						},
						Status:         teller.TransactionStatusPosted,
						Amount:         "-10.12",
						RunningBalance: nil,
						Type:           "card_payment",
					},
				}, nil).
				After(firstGetAccounts).
				Times(1)

			firstGetAccountBalance := authenticatedClient.EXPECT().
				GetAccountBalance(
					gomock.Any(),
					gomock.Eq(tellerBankAccountId),
				).
				Return(&teller.Balance{
					AccountId: tellerBankAccountId,
					Ledger:    "89.88",
					Available: "89.88",
					Links:     map[string]string{},
				}, nil).
				After(firstGetTransactions).
				Times(1)

			firstEnqueueJob = enqueuer.EXPECT().
				EnqueueJob(
					gomock.Any(),
					gomock.Eq(background.CalculateTransactionClusters),
					testutils.NewGenericMatcher(func(args background.CalculateTransactionClustersArguments) bool {
						return assert.Equal(t, link.AccountId, args.AccountId) &&
							assert.NotZero(t, args.BankAccountId)
					}),
				).
				After(firstGetAccountBalance).
				Return(nil).
				Times(1)
		}
		{ // Do our first Teller sync.
			args := background.SyncTellerArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Trigger:   "initial",
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		{ // Assert that our bank account was setup properly.
			bankAccounts := fixtures.ReadBankAccounts(t, clock, link)
			assert.Len(t, bankAccounts, 1, "should only have one bank account")
			bankAccount := bankAccounts[0]
			assert.Equal(t, tellerBankAccountId, bankAccount.TellerBankAccount.TellerId, "should have created the teller bank account")

			assert.EqualValues(t, 8988, bankAccount.CurrentBalance, "current balance should match the api")
			assert.EqualValues(t, 8988, bankAccount.AvailableBalance, "available balance should match the api")
		}

		{ // Make sure we have the one transaction now.
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 1, count, "should have one transaction now!")

			count = fixtures.CountAllTransactions(t, user.AccountId)
			assert.EqualValues(t, 1, count, "there should not be any EXTRA transactions that are deleted yet")
		}

		{ // Setup the mock calls for the second sync
			getAuthenticatedClient := client.EXPECT().
				GetAuthenticatedClient(
					gomock.AssignableToTypeOf("string"),
				).
				After(firstEnqueueJob).
				Return(authenticatedClient).
				Times(1)

			getAccounts := authenticatedClient.EXPECT().
				GetAccounts(
					gomock.Any(),
				).
				Return([]teller.Account{
					{
						Id:           tellerBankAccountId,
						Currency:     "USD",
						EnrollmentId: link.TellerLink.EnrollmentId,
						Institution: struct {
							Id   string "json:\"id\""
							Name string "json:\"name\""
						}{
							Id:   "navy_federal",
							Name: "Navy Federal",
						},
						Mask:    "1234",
						Links:   map[string]string{},
						Name:    "Primary Checking",
						Type:    teller.AccountTypeDepository,
						SubType: teller.AccountSubTypeChecking,
						Status:  teller.AccountStatusOpen,
					},
				}, nil).
				After(getAuthenticatedClient).
				Times(1)

			getTransactions := authenticatedClient.EXPECT().
				GetTransactions(
					gomock.Any(),
					gomock.Eq(tellerBankAccountId),
					gomock.Nil(),
					testutils.EqVal(25),
				).
				Return([]teller.Transaction{
					{
						Id:          transaction2Id,
						AccountId:   tellerBankAccountId,
						Date:        "2023-02-10",
						Description: "I am a new pending transaction!",
						Details: teller.TransactionDetails{
							ProcessingStatus: teller.TransactionProcessingStatusComplete,
							Category:         teller.TransactionCategoryGeneral,
						},
						Status:         teller.TransactionStatusPending,
						Amount:         "-1.00", // $1 pending
						RunningBalance: nil,
						Type:           "card_payment",
					},
					{
						Id:          transaction1Id,
						AccountId:   tellerBankAccountId,
						Date:        "2023-02-10",
						Description: "I am the first transaction",
						Details: teller.TransactionDetails{
							ProcessingStatus: teller.TransactionProcessingStatusComplete,
							Category:         teller.TransactionCategoryGeneral,
						},
						Status:         teller.TransactionStatusPosted,
						Amount:         "-10.12",
						RunningBalance: nil,
						Type:           "card_payment",
					},
				}, nil).
				After(getAccounts).
				Times(1)

			secondEnqueueJob = enqueuer.EXPECT().
				EnqueueJob(
					gomock.Any(),
					gomock.Eq(background.CalculateTransactionClusters),
					testutils.NewGenericMatcher(func(args background.CalculateTransactionClustersArguments) bool {
						return assert.Equal(t, link.AccountId, args.AccountId) &&
							assert.NotZero(t, args.BankAccountId)
					}),
				).
				After(getTransactions).
				Return(nil).
				Times(1)
		}

		{ // Do our second Teller sync.
			args := background.SyncTellerArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Trigger:   "visit",
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		{ // Make sure the available balance went down!
			bankAccounts := fixtures.ReadBankAccounts(t, clock, link)
			assert.Len(t, bankAccounts, 1, "should only have one bank account")
			bankAccount := bankAccounts[0]

			assert.EqualValues(t, 8988, bankAccount.CurrentBalance, "current balance should match the api")
			assert.EqualValues(t, 8888, bankAccount.AvailableBalance, "available should have gone down!")
		}

		{ // Make sure we have two transactions
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 2, count, "should have one transaction now!")

			count = fixtures.CountAllTransactions(t, user.AccountId)
			assert.EqualValues(t, 2, count, "there should not be any EXTRA transactions that are deleted yet")
		}

		{ // Setup the mock calls for the third sync
			getAuthenticatedClient := client.EXPECT().
				GetAuthenticatedClient(
					gomock.AssignableToTypeOf("string"),
				).
				After(secondEnqueueJob).
				Return(authenticatedClient).
				Times(1)

			getAccounts := authenticatedClient.EXPECT().
				GetAccounts(
					gomock.Any(),
				).
				Return([]teller.Account{
					{
						Id:           tellerBankAccountId,
						Currency:     "USD",
						EnrollmentId: link.TellerLink.EnrollmentId,
						Institution: struct {
							Id   string "json:\"id\""
							Name string "json:\"name\""
						}{
							Id:   "navy_federal",
							Name: "Navy Federal",
						},
						Mask:    "1234",
						Links:   map[string]string{},
						Name:    "Primary Checking",
						Type:    teller.AccountTypeDepository,
						SubType: teller.AccountSubTypeChecking,
						Status:  teller.AccountStatusOpen,
					},
				}, nil).
				After(getAuthenticatedClient).
				Times(1)

			getTransactions := authenticatedClient.EXPECT().
				GetTransactions(
					gomock.Any(),
					gomock.Eq(tellerBankAccountId),
					gomock.Nil(),
					testutils.EqVal(25),
				).
				Return([]teller.Transaction{
					{
						Id:          transaction3Id,
						AccountId:   tellerBankAccountId,
						Date:        "2023-02-10",
						Description: "I'm not pending anymore!",
						Details: teller.TransactionDetails{
							ProcessingStatus: teller.TransactionProcessingStatusComplete,
							Category:         teller.TransactionCategoryGeneral,
						},
						Status:         teller.TransactionStatusPosted,
						Amount:         "-5.00", // $5 cleared
						RunningBalance: nil,
						Type:           "card_payment",
					},
					{
						Id:          transaction1Id,
						AccountId:   tellerBankAccountId,
						Date:        "2023-02-10",
						Description: "I am the first transaction",
						Details: teller.TransactionDetails{
							ProcessingStatus: teller.TransactionProcessingStatusComplete,
							Category:         teller.TransactionCategoryGeneral,
						},
						Status:         teller.TransactionStatusPosted,
						Amount:         "-10.12",
						RunningBalance: nil,
						Type:           "card_payment",
					},
				}, nil).
				After(getAccounts).
				Times(1)

			enqueuer.EXPECT().
				EnqueueJob(
					gomock.Any(),
					gomock.Eq(background.CalculateTransactionClusters),
					testutils.NewGenericMatcher(func(args background.CalculateTransactionClustersArguments) bool {
						return assert.Equal(t, link.AccountId, args.AccountId) &&
							assert.NotZero(t, args.BankAccountId)
					}),
				).
				After(getTransactions).
				Return(nil).
				Times(1)
		}

		{ // Do our second Teller sync.
			args := background.SyncTellerArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Trigger:   "visit",
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		{ // Make sure the available balance went down!
			bankAccounts := fixtures.ReadBankAccounts(t, clock, link)
			assert.Len(t, bankAccounts, 1, "should only have one bank account")
			bankAccount := bankAccounts[0]

			assert.EqualValues(t, 8488, bankAccount.CurrentBalance, "current balance should match the api")
			assert.EqualValues(t, 8488, bankAccount.AvailableBalance, "available should have gone down!")
		}

		{ // Make sure we have two transactions
			count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
			assert.EqualValues(t, 2, count, "should have one transaction now!")

			count = fixtures.CountAllTransactions(t, user.AccountId)
			assert.EqualValues(t, 3, count, "there should be 3 transactions included the 1 deleted one")
		}
	})

	t.Run("no teller bank accounts found", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveATellerLink(t, clock, user)
		// Tweak to pending for initial setup
		link.TellerLink.Status = models.TellerLinkStatusPending
		testutils.MustDBUpdate(t, link.TellerLink)

		client := mockgenteller.NewMockClient(ctrl)
		authenticatedClient := mockgenteller.NewMockAuthenticatedClient(ctrl)
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		client.EXPECT().
			GetAuthenticatedClient(
				gomock.AssignableToTypeOf("string"),
			).
			Return(authenticatedClient).
			Times(1)

		authenticatedClient.EXPECT().
			GetAccounts(gomock.Any()).
			Return([]teller.Account{}, nil).
			Times(1)

		handler := background.NewSyncTellerHandler(
			log,
			db,
			clock,
			kms,
			client,
			publisher,
			enqueuer,
		)

		{ // Do our first Teller sync.
			args := background.SyncTellerArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Trigger:   "initial",
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.EqualError(t, err, "no Teller accounts found")
		}
	})
}
