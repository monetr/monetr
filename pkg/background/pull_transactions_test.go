package background

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/mock_plaid"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullTransactionsJob_Run(t *testing.T) {
	t.Skipf("skipping for now")
	t.Run("invalid link", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    math.MaxInt64,
			Start:     time.Now(),
			End:       time.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.EqualError(t, err, "failed to retrieve link to pull transactions: failed to get link: pg: no rows in result set")
		assert.Equal(t, "ROLLBACK", hook.LastEntry().Message, "Should rollback")
	})

	t.Run("manual link", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    link.LinkId,
			Start:     time.Now(),
			End:       time.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should not return an error, this way the job is not retried")

		// Make sure that if the job fails, nothing changes.
		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)
	})

	t.Run("no bank accounts", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, user)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    plaidLink.LinkId,
			Start:     time.Now(),
			End:       time.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "no error should be returned as the test should not be retried")
	})

	t.Run("missing access token", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, user)
		// Need at least one bank account.
		fixtures.GivenIHaveABankAccount(t, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    plaidLink.LinkId,
			Start:     time.Now(),
			End:       time.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.EqualError(t, err, "failed to retrieve access token for plaid link: failed to retrieve access token for plaid link: pg: no rows in result set")
	})

	t.Run("happy path, pull new transactions", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, user)

		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, gofakeit.UUID()))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		{ // Seeding plaid data.
			end := time.Now()
			start := end.Add(-365 * 24 * time.Hour)
			mock_plaid.MockGetRandomTransactions(t, start, end, 5000, []string{
				plaidBankAccount.PlaidAccountId,
			})
		}

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		{ // Do our first pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     time.Now().Add(-1 * 24 * time.Hour),
				End:       time.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count := fixtures.CountTransactions(t, user.AccountId)
		assert.NotZero(t, count, "should have more than zero transactions now")

		{ // Now try to retrieve a few more.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     time.Now().Add(-7 * 24 * time.Hour),
				End:       time.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		newCount := fixtures.CountTransactions(t, user.AccountId)
		assert.Greater(t, newCount, count, "must have more transactions after second run")

		assert.Equal(t, map[string]int{
			"POST https://sandbox.plaid.com/transactions/get": 2,
		}, httpmock.GetCallCountInfo())
	})

	t.Run("clear pending status on existing transaction", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, user)

		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, gofakeit.UUID()))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		end := time.Now()
		start := end.Add(-2 * 24 * time.Hour)
		transactions := mock_plaid.GenerateTransactions(t, start, end, 10, []string{
			plaidBankAccount.PlaidAccountId,
		})

		// Start with a pending transaction.
		transactions[0].Pending = true

		mock_plaid.MockGetTransactions(t, transactions)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		{ // Do our first pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     time.Now().Add(-7 * 24 * time.Hour),
				End:       time.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count := fixtures.CountTransactions(t, user.AccountId)
		assert.NotZero(t, count, "should have more than zero transactions now")

		// Update the transactions returned from the API, now it will not be pending.
		transactions[0].Pending = false

		{ // Now try to retrieve a few more.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     time.Now().Add(-7 * 24 * time.Hour),
				End:       time.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		newCount := fixtures.CountTransactions(t, user.AccountId)
		assert.EqualValues(t, newCount, count, "must have more transactions after second run")

		assert.Equal(t, map[string]int{
			"POST https://sandbox.plaid.com/transactions/get": 2,
		}, httpmock.GetCallCountInfo())
	})

	t.Run("no transactions", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, user)

		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, gofakeit.UUID()))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		end := time.Now()
		start := end.Add(-2 * 24 * time.Hour)
		transactions := mock_plaid.GenerateTransactions(t, start, end, 0, []string{
			plaidBankAccount.PlaidAccountId,
		})

		mock_plaid.MockGetTransactions(t, transactions)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    plaidLink.LinkId,
			Start:     time.Now().Add(-7 * 24 * time.Hour),
			End:       time.Now(),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		// Make sure that before we start there isn't anything in the database.
		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "must process job successfully")

		// No transactions should have been retrieved, so we should still be at zero.
		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

		assert.Equal(t, map[string]int{
			"POST https://sandbox.plaid.com/transactions/get": 1,
		}, httpmock.GetCallCountInfo())
	})
}
