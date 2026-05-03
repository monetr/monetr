package plaid_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/datasources/plaid/plaid_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/similar/similar_jobs"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v41/plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSyncPlaidJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		plaidBankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		nextCursor := gofakeit.UUID()
		pendingTxnId := gofakeit.UUID()
		firstSyncCall := plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				gomock.Nil(),
			).
			Return(&platypus.SyncResult{
				NextCursor: nextCursor,
				HasMore:    false,
				New: []platypus.Transaction{
					platypus.PlaidTransaction{
						Amount:                 1250,
						BankAccountId:          plaidBankAccount.PlaidBankAccount.PlaidId,
						Category:               []string{},
						Date:                   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "USD",
						IsPending:              true,
						MerchantName:           "Acme Corp",
						Name:                   "Acme Corp",
						OriginalDescription:    "ACME CORP",
						PendingTransactionId:   nil,
						TransactionId:          pendingTxnId,
					},
				},
				Updated: []platypus.Transaction{},
				Deleted: []string{},
				Accounts: []platypus.BankAccount{
					platypus.PlaidBankAccount{
						AccountId: plaidBankAccount.PlaidBankAccount.PlaidId,
						Balances: platypus.PlaidBankAccountBalances{
							Available: 100,
							Current:   100,
						},
						Mask:         *plaidBankAccount.Mask,
						Name:         plaidBankAccount.Name,
						OfficialName: plaidBankAccount.PlaidBankAccount.OfficialName,
						Type:         "depository",
						SubType:      "checking",
					},
				},
			}, nil)

		firstCalculateCall := enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		{ // Do our first plaid sync.
			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().KMS().Return(kms).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := plaid_jobs.SyncPlaid(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
					Trigger:   "webhook",
				},
			)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, count, "should have one transaction now!")

		count = fixtures.CountAllTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, count, "there should not be any EXTRA transactions that are deleted yet")

		plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				testutils.NewGenericMatcher(func(cursor *string) bool {
					return assert.EqualValues(t, nextCursor, *cursor, "provided cursor should match")
				}),
			).
			After(firstSyncCall).
			Return(&platypus.SyncResult{
				NextCursor: nextCursor,
				HasMore:    false,
				New: []platypus.Transaction{
					platypus.PlaidTransaction{
						Amount:                 1250,
						BankAccountId:          plaidBankAccount.PlaidBankAccount.PlaidId,
						Category:               []string{},
						Date:                   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "USD",
						IsPending:              false, // Replaces the pending transaction.
						MerchantName:           "Acme Corp",
						Name:                   "Acme Corp",
						OriginalDescription:    "ACME CORP",
						PendingTransactionId:   &pendingTxnId,
						TransactionId:          gofakeit.UUID(),
					},
				},
				Updated: []platypus.Transaction{},
				Deleted: []string{
					pendingTxnId,
				},
				Accounts: []platypus.BankAccount{
					platypus.PlaidBankAccount{
						AccountId: plaidBankAccount.PlaidBankAccount.PlaidId,
						Balances: platypus.PlaidBankAccountBalances{
							Available: 100,
							Current:   100,
						},
						Mask:         *plaidBankAccount.Mask,
						Name:         plaidBankAccount.Name,
						OfficialName: plaidBankAccount.PlaidBankAccount.OfficialName,
						Type:         "depository",
						SubType:      "checking",
					},
				},
			}, nil)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			MinTimes(1).
			After(firstCalculateCall).
			Return(nil)

		{ // Do our second plaid sync.
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().KMS().Return(kms).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := plaid_jobs.SyncPlaid(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
					Trigger:   "webhook",
				},
			)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count = fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, count, "should have one transaction now!")

		// There should be only one transaction since we handle merging.
		count = fixtures.CountAllTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, count, "should have a total of two transactions including the deleted one")
	})

	t.Run("recovers from mutation-during-pagination", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		plaidBankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		// Pre-seed a PlaidSync row so that SyncPlaid believes a prior sync already
		// stored a cursor. This is the exact condition the user hit: a stored
		// cursor from a successful first sync that Plaid now considers
		// mid-pagination because its data mutated in the meantime.
		storedCursor := "stored-cursor-from-previous-sync"
		{
			repo := repository.NewRepositoryFromSession(
				clock,
				user.UserId,
				user.AccountId,
				db,
				log,
			)
			require.NoError(
				t,
				repo.RecordPlaidSync(
					t.Context(),
					*plaidLink.PlaidLinkId,
					storedCursor,
					"webhook",
					0, 0, 0,
				),
				"must be able to seed prior plaid sync",
			)
		}

		// Advance the mock clock so the pre-seeded PlaidSync row has an earlier
		// timestamp than the row the recovery path will write. GetLastPlaidSync
		// orders by timestamp DESC with Limit(1), so the test would be
		// non-deterministic if both rows shared the same mock timestamp.
		clock.Add(time.Second)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		// First call uses the stored cursor. Plaid responds with the
		// mutation-during-pagination error, exactly as produced by
		// platypus/sync.go's after() wrapper.
		mutationErr := errors.Wrap(
			&platypus.PlatypusError{PlaidError: plaid.PlaidError{
				ErrorType:    "TRANSACTIONS_ERROR",
				ErrorCode:    platypus.ErrorCodeTransactionsSyncMutationDuringPagination,
				ErrorMessage: "Underlying transaction data changed since last page was fetched. Please restart pagination from last update.",
			}},
			"failed to sync data with Plaid",
		)
		firstSyncCall := plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				testutils.NewGenericMatcher(func(cursor *string) bool {
					return assert.NotNil(t, cursor, "first call must use the stored cursor, not nil") &&
						assert.EqualValues(t, storedCursor, *cursor, "first call must use the stored cursor")
				}),
			).
			Return(nil, mutationErr).
			Times(1)

		// Second call is the restart. SyncPlaid must detect the mutation error and
		// reset the cursor to nil; this mock expectation is the assertion that the
		// reset happened.
		freshCursor := gofakeit.UUID()
		freshTxnId := gofakeit.UUID()
		plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				gomock.Nil(),
			).
			After(firstSyncCall).
			Return(&platypus.SyncResult{
				NextCursor: freshCursor,
				HasMore:    false,
				New: []platypus.Transaction{
					platypus.PlaidTransaction{
						Amount:                 1250,
						BankAccountId:          plaidBankAccount.PlaidBankAccount.PlaidId,
						Category:               []string{},
						Date:                   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "USD",
						IsPending:              false,
						MerchantName:           "Acme Corp",
						Name:                   "Acme Corp",
						OriginalDescription:    "ACME CORP",
						PendingTransactionId:   nil,
						TransactionId:          freshTxnId,
					},
				},
				Updated: []platypus.Transaction{},
				Deleted: []string{},
				Accounts: []platypus.BankAccount{
					platypus.PlaidBankAccount{
						AccountId: plaidBankAccount.PlaidBankAccount.PlaidId,
						Balances: platypus.PlaidBankAccountBalances{
							Available: 100,
							Current:   100,
						},
						Mask:         *plaidBankAccount.Mask,
						Name:         plaidBankAccount.Name,
						OfficialName: plaidBankAccount.PlaidBankAccount.OfficialName,
						Type:         "depository",
						SubType:      "checking",
					},
				},
			}, nil).
			Times(1)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		// Before we start the database should be clean.
		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
		context.EXPECT().KMS().Return(kms).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := plaid_jobs.SyncPlaid(
			mockqueue.NewMockContext(context),
			plaid_jobs.SyncPlaidArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Trigger:   "webhook",
			},
		)
		assert.NoError(t, err, "must recover from mutation-during-pagination and finish without error")

		// The transaction from the second (restart) call must have been persisted;
		// this proves the recovery path actually produced real data.
		count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, count, "must persist the transaction returned by the null-cursor restart")

		// The latest PlaidSync row must hold the fresh cursor, not the stale
		// pre-seeded one. GetLastPlaidSync orders by timestamp DESC, so writing the
		// fresh cursor naturally supersedes the pre-seeded stale row.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		latestSync, err := repo.GetLastPlaidSync(t.Context(), *plaidLink.PlaidLinkId)
		require.NoError(t, err, "must be able to read the latest plaid sync")
		require.NotNil(t, latestSync, "must have a plaid sync row after recovery")
		assert.Equal(
			t,
			freshCursor,
			latestSync.NextCursor,
			"the latest plaid sync row must hold the cursor from the null-cursor restart, not the stale one",
		)
	})

	t.Run("gives up after repeated mutation-during-pagination errors", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
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

		storedCursor := "stored-cursor-from-previous-sync"
		{
			repo := repository.NewRepositoryFromSession(
				clock,
				user.UserId,
				user.AccountId,
				db,
				log,
			)
			require.NoError(
				t,
				repo.RecordPlaidSync(
					t.Context(),
					*plaidLink.PlaidLinkId,
					storedCursor,
					"webhook",
					0, 0, 0,
				),
				"must be able to seed prior plaid sync",
			)
		}

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		mutationErr := errors.Wrap(
			&platypus.PlatypusError{PlaidError: plaid.PlaidError{
				ErrorType:    "TRANSACTIONS_ERROR",
				ErrorCode:    platypus.ErrorCodeTransactionsSyncMutationDuringPagination,
				ErrorMessage: "Underlying transaction data changed since last page was fetched. Please restart pagination from last update.",
			}},
			"failed to sync data with Plaid",
		)
		// Both calls return the same error. The job should call Sync exactly twice;
		// once with the stored cursor, once with the null restart; then bail out
		// rather than looping forever.
		plaidClient.EXPECT().
			Sync(gomock.Any(), gomock.Any()).
			Return(nil, mutationErr).
			Times(2)

		// No transactions get written so no similarity recalc gets enqueued.
		enqueuer.EXPECT().EnqueueAt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().KMS().Return(kms).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := plaid_jobs.SyncPlaid(
			mockqueue.NewMockContext(context),
			plaid_jobs.SyncPlaidArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Trigger:   "webhook",
			},
		)
		assert.Error(t, err, "must surface the error once the restart budget is exhausted")
		assert.True(
			t,
			platypus.IsPlaidErrorCode(err, platypus.ErrorCodeTransactionsSyncMutationDuringPagination),
			"returned error must still carry the underlying plaid error code",
		)
	})

	t.Run("recovery preserves transactions from prior successful sync", func(t *testing.T) {
		// End-to-end reproduction of the user-reported scenario:
		//   1. An initial sync runs successfully, persists transactions, and stores
		//      a cursor in plaid_syncs.
		//   2. Days later, a webhook-triggered sync uses that stored cursor and
		//      receives TRANSACTIONS_SYNC_MUTATION_DURING_PAGINATION because Plaid
		//      mutated the data in the meantime.
		//   3. The job auto-recovers by restarting pagination with a null cursor.
		//      Plaid replays every current transaction as "added".
		//
		// This test proves the recovery does not clobber the previously- synced
		// data: existing monetr Transaction rows keep their original TransactionId
		// values (they are updated in place via plaid_id lookup, not recreated), no
		// rows are marked deleted, and no duplicates are inserted.
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		plaidBankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		// Two transactions that will exist across both sync runs; these are what we
		// are proving the recovery path does not clobber.
		txn1PlaidId := gofakeit.UUID()
		txn2PlaidId := gofakeit.UUID()
		initialCursor := gofakeit.UUID()

		buildPlaidTxn := func(plaidId string, amount int64, date time.Time) platypus.PlaidTransaction {
			return platypus.PlaidTransaction{
				Amount:                 amount,
				BankAccountId:          plaidBankAccount.PlaidBankAccount.PlaidId,
				Category:               []string{},
				Date:                   date,
				ISOCurrencyCode:        "USD",
				UnofficialCurrencyCode: "USD",
				IsPending:              false,
				MerchantName:           "Acme Corp",
				Name:                   "Acme Corp",
				OriginalDescription:    "ACME CORP",
				PendingTransactionId:   nil,
				TransactionId:          plaidId,
			}
		}

		plaidAccounts := []platypus.BankAccount{
			platypus.PlaidBankAccount{
				AccountId: plaidBankAccount.PlaidBankAccount.PlaidId,
				Balances: platypus.PlaidBankAccountBalances{
					Available: 100,
					Current:   100,
				},
				Mask:         *plaidBankAccount.Mask,
				Name:         plaidBankAccount.Name,
				OfficialName: plaidBankAccount.PlaidBankAccount.OfficialName,
				Type:         "depository",
				SubType:      "checking",
			},
		}

		// Step 1: first sync; starts from nil cursor, returns two transactions and
		// finishes cleanly with the initial cursor.
		firstSyncCall := plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				gomock.Nil(),
			).
			Return(&platypus.SyncResult{
				NextCursor: initialCursor,
				HasMore:    false,
				New: []platypus.Transaction{
					buildPlaidTxn(txn1PlaidId, 1250, time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local)),
					buildPlaidTxn(txn2PlaidId, 4599, time.Date(2023, 01, 02, 0, 0, 0, 0, time.Local)),
				},
				Updated:  []platypus.Transaction{},
				Deleted:  []string{},
				Accounts: plaidAccounts,
			}, nil).
			Times(1)

		firstCalculateCall := enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		{ // Run the initial sync.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().KMS().Return(kms).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := plaid_jobs.SyncPlaid(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
					Trigger:   "webhook",
				},
			)
			require.NoError(t, err, "initial sync must succeed")
		}

		// Capture the monetr TransactionIds assigned during the initial sync; the
		// no-clobber assertion below compares these against what is in the database
		// after recovery.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		initialTxns, err := repo.GetTransactionsByPlaidId(
			t.Context(),
			plaidLink.LinkId,
			[]string{txn1PlaidId, txn2PlaidId},
		)
		require.NoError(t, err, "must be able to read transactions after initial sync")
		require.Len(t, initialTxns, 2, "initial sync must have persisted both transactions")
		originalTxn1Id := initialTxns[txn1PlaidId].TransactionId
		originalTxn2Id := initialTxns[txn2PlaidId].TransactionId

		// Step 2: simulate the gap between the webhook-triggered syncs. Three days
		// is enough time for Plaid's upstream data to mutate and invalidate the
		// stored cursor.
		clock.Add(72 * time.Hour)

		// Step 3a: second sync's first call; uses the stored cursor and Plaid
		// returns TRANSACTIONS_SYNC_MUTATION_DURING_PAGINATION.
		mutationErr := errors.Wrap(
			&platypus.PlatypusError{PlaidError: plaid.PlaidError{
				ErrorType:    "TRANSACTIONS_ERROR",
				ErrorCode:    platypus.ErrorCodeTransactionsSyncMutationDuringPagination,
				ErrorMessage: "Underlying transaction data changed since last page was fetched. Please restart pagination from last update.",
			}},
			"failed to sync data with Plaid",
		)
		mutationErrorCall := plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				testutils.NewGenericMatcher(func(cursor *string) bool {
					return assert.NotNil(t, cursor, "second sync must start from the stored cursor, not nil") &&
						assert.EqualValues(t, initialCursor, *cursor, "second sync must start from the cursor recorded by the initial sync")
				}),
			).
			After(firstSyncCall).
			Return(nil, mutationErr).
			Times(1)

		// Step 3b: recovery call; SyncPlaid must detect the mutation error
		// and retry with a nil cursor. Plaid replays both existing
		// transactions plus one new one.
		txn3PlaidId := gofakeit.UUID()
		recoveryCursor := gofakeit.UUID()
		plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				gomock.Nil(),
			).
			After(mutationErrorCall).
			Return(&platypus.SyncResult{
				NextCursor: recoveryCursor,
				HasMore:    false,
				New: []platypus.Transaction{
					buildPlaidTxn(txn1PlaidId, 1250, time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local)),
					buildPlaidTxn(txn2PlaidId, 4599, time.Date(2023, 01, 02, 0, 0, 0, 0, time.Local)),
					buildPlaidTxn(txn3PlaidId, 750, time.Date(2023, 01, 03, 0, 0, 0, 0, time.Local)),
				},
				Updated:  []platypus.Transaction{},
				Deleted:  []string{},
				Accounts: plaidAccounts,
			}, nil).
			Times(1)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			After(firstCalculateCall).
			Return(nil).
			MinTimes(1)

		{ // Run the second sync, which triggers the mutation-error recovery.
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().KMS().Return(kms).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := plaid_jobs.SyncPlaid(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
					Trigger:   "webhook",
				},
			)
			require.NoError(t, err, "recovery sync must succeed without clobbering data")
		}

		// No-clobber assertions:
		//
		// Three transactions total (the two from the initial sync plus the one the
		// recovery call added). No duplicates, no soft-deletes.
		assert.EqualValues(
			t,
			3,
			fixtures.CountNonDeletedTransactions(t, user.AccountId),
			"recovery must produce exactly three live transactions (original two + one new)",
		)
		assert.EqualValues(
			t,
			3,
			fixtures.CountAllTransactions(t, user.AccountId),
			"recovery must not soft-delete any of the original transactions",
		)

		// Same plaid transaction ids must still resolve to the same monetr
		// TransactionId values; proving recovery updates in place rather than
		// re-creating the rows.
		postRecoveryTxns, err := repo.GetTransactionsByPlaidId(
			t.Context(),
			plaidLink.LinkId,
			[]string{txn1PlaidId, txn2PlaidId, txn3PlaidId},
		)
		require.NoError(t, err, "must be able to read transactions after recovery")
		require.Len(t, postRecoveryTxns, 3, "all three plaid ids must resolve to monetr transactions")
		assert.Equal(
			t,
			originalTxn1Id,
			postRecoveryTxns[txn1PlaidId].TransactionId,
			"txn1's monetr TransactionId must survive the recovery unchanged",
		)
		assert.Equal(
			t,
			originalTxn2Id,
			postRecoveryTxns[txn2PlaidId].TransactionId,
			"txn2's monetr TransactionId must survive the recovery unchanged",
		)

		// The latest plaid_syncs row must hold the cursor returned by the recovery
		// call, not the stale initialCursor.
		latestSync, err := repo.GetLastPlaidSync(t.Context(), *plaidLink.PlaidLinkId)
		require.NoError(t, err, "must be able to read latest plaid sync")
		require.NotNil(t, latestSync, "must have a plaid sync row after recovery")
		assert.Equal(
			t,
			recoveryCursor,
			latestSync.NextCursor,
			"latest plaid sync row must hold the recovery cursor, not the stale initial one",
		)
	})

	t.Run("dont't trample custom name", func(t *testing.T) {
		// Reproduction of the user-reported scenario from
		// https://github.com/monetr/monetr/issues/3167:
		//   1. An initial sync persists a pending transaction with whatever Name
		//      plaid sent.
		//   2. The user renames the transaction; the Name column is the user
		//      editable display label.
		//   3. The cleared version of the same transaction arrives in a later sync,
		//      with a Name field that is different from both the original plaid
		//      Name and the user's customized one.
		//
		// Before the fix this test exercises, the second sync would clobber the
		// user's rename back to whatever plaid sent in the cleared payload. The new
		// code in syncPlaidTransaction leaves Name alone and only updates
		// OriginalName when the transaction transitions to non-pending, which is
		// the original https://github.com/monetr/monetr/issues/1714 intent (capture
		// the cleared original description) without the trampling side effect.
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		plaidBankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&plaidLink,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Any(),
				gomock.Eq(plaidLink.PlaidLink.PlaidId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		initialCursor := gofakeit.UUID()
		pendingTxnId := gofakeit.UUID()
		clearedTxnId := gofakeit.UUID()

		plaidAccounts := []platypus.BankAccount{
			platypus.PlaidBankAccount{
				AccountId: plaidBankAccount.PlaidBankAccount.PlaidId,
				Balances: platypus.PlaidBankAccountBalances{
					Available: 100,
					Current:   100,
				},
				Mask:         *plaidBankAccount.Mask,
				Name:         plaidBankAccount.Name,
				OfficialName: plaidBankAccount.PlaidBankAccount.OfficialName,
				Type:         "depository",
				SubType:      "checking",
			},
		}

		// First sync: plaid sends a pending transaction. monetr stores it with name
		// "Acme Corp", which is what the user will then go and rename below.
		firstSyncCall := plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				gomock.Nil(),
			).
			Return(&platypus.SyncResult{
				NextCursor: initialCursor,
				HasMore:    false,
				New: []platypus.Transaction{
					platypus.PlaidTransaction{
						Amount:                 1250,
						BankAccountId:          plaidBankAccount.PlaidBankAccount.PlaidId,
						Category:               []string{},
						Date:                   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "USD",
						IsPending:              true,
						MerchantName:           "Acme Corp",
						Name:                   "Acme Corp",
						OriginalDescription:    "ACME CORP",
						PendingTransactionId:   nil,
						TransactionId:          pendingTxnId,
					},
				},
				Updated:  []platypus.Transaction{},
				Deleted:  []string{},
				Accounts: plaidAccounts,
			}, nil).
			Times(1)

		firstCalculateCall := enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		{ // Initial sync that creates the pending transaction.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().KMS().Return(kms).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := plaid_jobs.SyncPlaid(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
					Trigger:   "webhook",
				},
			)
			require.NoError(t, err, "initial sync must succeed")
		}

		// Simulate the user renaming the transaction. The user-facing API path
		// ultimately writes the new Name to the transactions row; MustDBUpdate is
		// the same UPDATE without any auditing or queueing side effects that could
		// perturb the gomock expectations.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		afterFirstSync, err := repo.GetTransactionsByPlaidId(
			t.Context(),
			plaidLink.LinkId,
			[]string{pendingTxnId},
		)
		require.NoError(t, err, "must be able to read the pending transaction after the first sync")
		require.Len(t, afterFirstSync, 1, "first sync must have persisted exactly one transaction")
		renamed := afterFirstSync[pendingTxnId]
		originalMonetrTxnId := renamed.TransactionId
		renamed.Name = "My Custom Coffee Shop"
		testutils.MustDBUpdate(t, &renamed)

		// Second sync: plaid replays the same logical transaction, now cleared.
		// Both Name and OriginalDescription are deliberately different from the
		// initial sync. The Name diff is what would have triggered the old
		// overwrite of the user's rename; the OriginalDescription diff is what
		// exercises the new OriginalName update branch.
		plaidClient.EXPECT().
			Sync(
				gomock.Any(),
				testutils.NewGenericMatcher(func(cursor *string) bool {
					return assert.NotNil(t, cursor, "second sync must pass a cursor, not nil") &&
						assert.EqualValues(t, initialCursor, *cursor, "second sync must use the cursor returned by the first sync")
				}),
			).
			After(firstSyncCall).
			Return(&platypus.SyncResult{
				NextCursor: gofakeit.UUID(),
				HasMore:    false,
				New: []platypus.Transaction{
					platypus.PlaidTransaction{
						Amount:                 1250,
						BankAccountId:          plaidBankAccount.PlaidBankAccount.PlaidId,
						Category:               []string{},
						Date:                   time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "USD",
						IsPending:              false,
						MerchantName:           "Acme Corp",
						Name:                   "Pos Debit 5988 Acme",
						OriginalDescription:    "ACME CORP CLEARED",
						PendingTransactionId:   &pendingTxnId,
						TransactionId:          clearedTxnId,
					},
				},
				Updated: []platypus.Transaction{},
				Deleted: []string{
					pendingTxnId,
				},
				Accounts: plaidAccounts,
			}, nil).
			Times(1)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Eq(similar_jobs.CalculateTransactionClustersArguments{
					AccountId:     plaidBankAccount.AccountId,
					BankAccountId: plaidBankAccount.BankAccountId,
				}),
			).
			After(firstCalculateCall).
			Return(nil).
			MinTimes(1)

		{ // Second sync that clears the transaction.
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).MinTimes(1)
			context.EXPECT().DB().Return(db).MinTimes(1)
			context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
			context.EXPECT().KMS().Return(kms).MinTimes(1)
			context.EXPECT().Log().Return(log).MinTimes(1)
			context.EXPECT().Platypus().Return(plaidPlatypus).MinTimes(1)
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := plaid_jobs.SyncPlaid(
				mockqueue.NewMockContext(context),
				plaid_jobs.SyncPlaidArguments{
					AccountId: user.AccountId,
					LinkId:    plaidLink.LinkId,
					Trigger:   "webhook",
				},
			)
			require.NoError(t, err, "clearing sync must succeed")
		}

		// The actual claims under test:
		//   - Name still reflects the user's rename (the #3167 fix).
		//   - OriginalName tracks the cleared payload's OriginalDescription (the
		//     #1714 intent, now sourced from the right field).
		//   - Same monetr row updated in place, not recreated.
		//   - No duplicates and no soft-deletes left behind.
		afterSecondSync, err := repo.GetTransactionsByPlaidId(
			t.Context(),
			plaidLink.LinkId,
			[]string{clearedTxnId},
		)
		require.NoError(t, err, "must be able to read the transaction after it clears")
		cleared, found := afterSecondSync[clearedTxnId]
		require.True(t, found, "cleared plaid id must resolve to the existing monetr transaction")
		assert.Equal(
			t,
			"My Custom Coffee Shop",
			cleared.Name,
			"the user's renamed Name must survive the clearing sync",
		)
		assert.Equal(
			t,
			"ACME CORP CLEARED",
			cleared.OriginalName,
			"OriginalName must update to the new original description on the pending-to-cleared transition",
		)
		assert.False(t, cleared.IsPending, "transaction must now be cleared")
		assert.Equal(
			t,
			originalMonetrTxnId,
			cleared.TransactionId,
			"monetr's TransactionId must survive the clearing sync (updated in place, not recreated)",
		)

		assert.EqualValues(
			t,
			1,
			fixtures.CountNonDeletedTransactions(t, user.AccountId),
			"must still have exactly one live monetr transaction",
		)
		assert.EqualValues(
			t,
			1,
			fixtures.CountAllTransactions(t, user.AccountId),
			"must not have soft-deleted the original monetr transaction",
		)
	})

	t.Run("initial setup", func(t *testing.T) {
		// TODO!!
	})

	t.Run("remove with spending", func(t *testing.T) {
		// TODO!!
	})

	t.Run("no updates", func(t *testing.T) {
		// TODO!!
	})
}
