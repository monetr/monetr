package jobs

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/gocraft/work"
	"github.com/jarcoal/httpmock"
	"github.com/monetrapp/rest-api/pkg/internal/mock_plaid"
	"github.com/monetrapp/rest-api/pkg/internal/mock_secrets"
	"github.com/monetrapp/rest-api/pkg/internal/plaid_helper"
	"github.com/monetrapp/rest-api/pkg/internal/testutils"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestPullAccountBalances(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)

		db := testutils.GetPgDatabase(t)
		cache := testutils.GetRedisPool(t)

		account, plaidData := testutils.SeedAccount(t, db, testutils.WithPlaidAccount)

		var linkId uint64
		require.NoError(t, db.RunInTransaction(context.Background(), func(txn *pg.Tx) error {
			repo := repository.NewRepositoryFromSession(account.UserId, account.AccountId, txn)
			links, err := repo.GetLinks()
			require.NoError(t, err, "must retrieve links for account")
			require.Len(t, links, 1, "should have exactly one link")

			linkId = links[0].LinkId // There should only be one so we want to use that one.
			return nil
		}), "must retrieve linkId")

		plaidClient := plaid_helper.NewPlaidClient(log, plaid.ClientOptions{
			ClientID:    gofakeit.UUID(),
			Secret:      gofakeit.UUID(),
			Environment: plaid.Sandbox,
			HTTPClient:  http.DefaultClient,
		})

		plaidSecrets := mock_secrets.NewMockPlaidSecrets()
		for accessToken, data := range plaidData.PlaidTokens {
			plaidSecrets = plaidSecrets.WithSecret(account.AccountId, data.ItemId, accessToken)
		}

		job := NewJobManager(log, cache, db, plaidClient, nil, plaidSecrets).(*jobManagerBase)
		defer require.NoError(t, job.Close(), "must close job manager")

		// TODO (elliotcourant) Tweak the plaid data balances before we make our request. This way we can add proper
		//  assertions around whether or not the balances in our database have actually updated.
		// Mock the plaid get accounts endpoint with our seeded plaid data.
		mock_plaid.MockGetAccountsExtended(t, plaidData)

		err := job.pullAccountBalances(&work.Job{
			Name:       PullAccountBalances,
			ID:         gofakeit.UUID(),
			EnqueuedAt: time.Now().Unix(),
			Args: map[string]interface{}{
				"accountId": account.AccountId,
				"linkId":    linkId,
			},
			Unique:   true,
			Fails:    0,
			LastErr:  "",
			FailedAt: 0,
		})
		assert.NoError(t, err, "job should succeed")

		// Make sure the API calls we made match what we expect.
		assert.Equal(t, map[string]int{
			"POST https://sandbox.plaid.com/accounts/get": 1,
		}, httpmock.GetCallCountInfo(), "call counts should match")
	})
}
