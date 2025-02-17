package platypus

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/stretchr/testify/assert"
)

func TestPlaidSync(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.New()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		kms := secrets.NewPlaintextKMS()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		secret, err := repository.NewSecretsRepository(
			log,
			clock,
			db,
			kms,
			plaidLink.AccountId,
		).Read(context.Background(), plaidLink.PlaidLink.SecretId)
		assert.NoError(t, err, "must be able to read the secret")

		end := clock.Now().UTC().Truncate(time.Hour)
		start := clock.Now().UTC().Add(-30 * 24 * time.Hour).Truncate(time.Hour)
		bankAccounts := []string{
			"1234",
		}
		transactions := mock_plaid.GenerateTransactions(t, start, end, 10, bankAccounts)

		mock_plaid.MockSync(t, transactions)

		platypus := NewPlaid(log, clock, kms, db, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
			OAuthDomain:  "localhost",
		})

		link := &models.Link{
			LinkId:    "link_foo",
			AccountId: user.AccountId,
		}

		client, err := platypus.NewClient(
			context.Background(),
			link,
			secret.Value,
			gofakeit.UUID(),
		)
		assert.NoError(t, err, "should create platypus")
		assert.NotNil(t, platypus, "should not be nil")

		result, err := client.Sync(t.Context(), nil)
		assert.NoError(t, err)
		assert.Len(t, result.New, len(transactions))
	})

	t.Run("pagination error", func(t *testing.T) {
		clock := clock.New()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		kms := secrets.NewPlaintextKMS()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		secret, err := repository.NewSecretsRepository(
			log,
			clock,
			db,
			kms,
			plaidLink.AccountId,
		).Read(context.Background(), plaidLink.PlaidLink.SecretId)
		assert.NoError(t, err, "must be able to read the secret")

		mock_plaid.MockSyncError(t, plaid.PlaidError{
			ErrorType:    "TRANSACTIONS_ERROR",
			ErrorCode:    "TRANSACTIONS_SYNC_MUTATION_DURING_PAGINATION",
			ErrorMessage: "Underlying transaction data changed since last page was fetched. Please restart pagination from last update.",
		})

		platypus := NewPlaid(log, clock, kms, db, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
			OAuthDomain:  "localhost",
		})

		link := &models.Link{
			LinkId:    "link_foo",
			AccountId: user.AccountId,
		}

		client, err := platypus.NewClient(
			context.Background(),
			link,
			secret.Value,
			gofakeit.UUID(),
		)
		assert.NoError(t, err, "should create platypus")
		assert.NotNil(t, platypus, "should not be nil")

		result, err := client.Sync(t.Context(), nil)
		assert.EqualError(t, err, "failed to sync data with Plaid: plaid API call failed with [TRANSACTIONS_ERROR - TRANSACTIONS_SYNC_MUTATION_DURING_PAGINATION]Underlying transaction data changed since last page was fetched. Please restart pagination from last update.")
		assert.Nil(t, result)
		assert.IsType(t, errors.Cause(err), new(PlatypusError), "Should be a platypus error")
		assert.Equal(t, "TRANSACTIONS_SYNC_MUTATION_DURING_PAGINATION", (errors.Cause(err).(*PlatypusError)).ErrorCode, "Must be able to extract error code")
	})
}
