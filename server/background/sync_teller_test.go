// vim: foldmethod=indent

package background_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockgenteller"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/teller"
	"github.com/robfig/cron"
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

		{ // Do our third Teller sync.
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

		{ // Make sure our balances adjust proprly for the cleared transaction!
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

	// This test handles the scenario where the amount of a posted transaction has
	// changed since the last sync. This should technically never happen, but it
	// since we don't control teller's API and they could have a bug where they
	// give us the wrong amount its best to have a way to handle this so that the
	// ledger balance can be adjusted.
	t.Run("handle posted amount change rare", func(t *testing.T) {
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
		name, _ := fixtures.GivenIHaveATransactionName(t, clock)
		transaction1Id := gofakeit.Generate("txn_????????????????")

		var firstEnqueueJob *gomock.Call

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
						Description: name,
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
						Id:          transaction1Id,
						AccountId:   tellerBankAccountId,
						Date:        "2023-02-10",
						Description: name,
						Details: teller.TransactionDetails{
							ProcessingStatus: teller.TransactionProcessingStatusComplete,
							Category:         teller.TransactionCategoryGeneral,
						},
						Status:         teller.TransactionStatusPosted,
						Amount:         "-5.12",
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

			assert.EqualValues(t, 9488, bankAccount.CurrentBalance, "current balance should match the api")
			assert.EqualValues(t, 9488, bankAccount.AvailableBalance, "available should have gone down!")
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

	t.Run("no teller link - manunal", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)

		client := mockgenteller.NewMockClient(ctrl)
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
			assert.NoError(t, err, "jobs can be retried on error, but a bad teller link should not return an error")
		}
	})

	t.Run("bad account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		client := mockgenteller.NewMockClient(ctrl)
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

		{ // Do our first Teller sync.
			args := background.SyncTellerArguments{
				AccountId: math.MaxUint32,
				LinkId:    math.MaxUint32,
				Trigger:   "initial",
			}

			argsEncoded, err := background.DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.EqualError(t, err, "failed to retrieve link to sync with Teller: failed to get link: pg: no rows in result set")
		}
	})

	t.Run("transactions fail to fetch", func(t *testing.T) {
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
				Return(nil, errors.New("failed to fetch transactions")).
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

			enqueuer.EXPECT().
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
				// Because we didn't get any transactions, we shouldn't calculate
				// transaction clusters.
				Times(0)
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

			// Even if we fail to retrieve transactions we should not fail hard.
			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}
	})

	t.Run("balance fails to sync", func(t *testing.T) {
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

			txnName, _ := fixtures.GivenIHaveATransactionName(t, clock)
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
						Description: txnName,
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
				Return(nil, errors.New("failed to retrieve balance")).
				After(firstGetTransactions).
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
				After(firstGetAccountBalance).
				Return(nil).
				// Should not calculate transaction clusters if we abort.
				Times(0)
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
			assert.EqualError(t, err, "failed to hard sync balance for account: failed to retrieve balance")
		}
	})
}

func TestSyncTellerHandler_QueueName(t *testing.T) {
	assert.NotEmpty(t, background.SyncTellerHandler{}.QueueName(), "queue name cannot be empty")
}

func TestSyncTellerHandler_DefaultSchedule(t *testing.T) {
	schedule := background.SyncTellerHandler{}.DefaultSchedule()
	cronSchedule, err := cron.Parse(schedule)
	assert.NoError(t, err, "must parse cron schedule")
	assert.NotNil(t, cronSchedule, "must get a valid schedule")
	now := time.Now()
	assert.Greater(t, cronSchedule.Next(now), now, "cron should be in the future")
}

func TestSyncTellerHandler_EnqueueTriggeredJob(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveATellerLink(t, clock, user)

		client := mockgenteller.NewMockClient(ctrl)
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

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(handler.QueueName()),
				gomock.Eq(background.SyncTellerArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "cron",
				}),
			).
			Return(nil).
			Times(1)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")
	})

	t.Run("fail gracefully", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveATellerLink(t, clock, user)

		client := mockgenteller.NewMockClient(ctrl)
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

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(handler.QueueName()),
				gomock.Eq(background.SyncTellerArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "cron",
				}),
			).
			Return(errors.New("failed to enqueue")).
			Times(1)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")
	})

	t.Run("will not enqueue a bad link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveATellerLink(t, clock, user)
		// Tweak to pending for initial setup
		link.TellerLink.Status = models.TellerLinkStatusDisconnected
		testutils.MustDBUpdate(t, link.TellerLink)

		client := mockgenteller.NewMockClient(ctrl)
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

		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(handler.QueueName()),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")
	})

	t.Run("will not enqueue a link synced too recently", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := testutils.GetKMS(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveATellerLink(t, clock, user)
		// Update the last attempted update
		link.TellerLink.LastAttemptedUpdate = myownsanity.TimeP(clock.Now())
		testutils.MustDBUpdate(t, link.TellerLink)

		client := mockgenteller.NewMockClient(ctrl)
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

		// The first time we sync, we should not call the enqueue job at all.
		firstEnqueuer := enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(handler.QueueName()),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")

		// Move the clock forward
		clock.Add(24 * time.Hour)

		// The second time we sync we should though, as it has been long enough for
		// our link to need an update.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(handler.QueueName()),
				gomock.Eq(background.SyncTellerArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "cron",
				}),
			).
			After(firstEnqueuer).
			Return(nil).
			Times(1)

		err = handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err, "should not return an error")
	})
}
