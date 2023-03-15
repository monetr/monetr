package background

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/mockgen"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncPlaidJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, user)

		accessToken := gofakeit.UUID()
		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, accessToken))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Eq(accessToken),
				gomock.Eq(plaidLink.PlaidLink.ItemId),
			).
			Return(plaidClient, nil).
			AnyTimes()

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
			).
			Return([]platypus.BankAccount{
				platypus.PlaidBankAccount{
					AccountId: plaidBankAccount.PlaidAccountId,
					Balances: platypus.PlaidBankAccountBalances{
						Available: 100,
						Current:   100,
					},
					Mask:         plaidBankAccount.Mask,
					Name:         plaidBankAccount.Name,
					OfficialName: plaidBankAccount.PlaidOfficialName,
					Type:         "depository",
					SubType:      "checking",
				},
			}, nil).
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
						BankAccountId:          plaidBankAccount.PlaidAccountId,
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
			}, nil)

		handler := NewSyncPlaidHandler(log, db, provider, plaidPlatypus, publisher)

		{ // Do our first plaid sync.
			args := SyncPlaidArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Trigger:   "webhook",
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
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
						BankAccountId:          plaidBankAccount.PlaidAccountId,
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
			}, nil)

		{ // Do our second plaid sync.
			args := SyncPlaidArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Trigger:   "webhook",
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count = fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, count, "should have one transaction now!")

		count = fixtures.CountAllTransactions(t, user.AccountId)
		assert.EqualValues(t, 2, count, "should have a total of two transactions including the deleted one")
	})
}
