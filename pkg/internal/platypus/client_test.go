package platypus

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/internal/mock_plaid"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPlaidClient_GetAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		accountId := testutils.GetAccountIdForTest(t)

		accessToken := gofakeit.UUID()

		account := mock_plaid.BankAccountFixture(t)

		mock_plaid.MockGetAccounts(t, []plaid.AccountBase{
			account,
		})

		client := NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		link := &models.Link{
			LinkId:    1234,
			AccountId: accountId,
		}

		platypus, err := client.NewClient(context.Background(), link, accessToken)
		assert.NoError(t, err, "should create platypus")
		assert.NotNil(t, platypus, "should not be nil")

		accounts, err := platypus.GetAccounts(context.Background(), account.GetAccountId())
		assert.NoError(t, err, "should not return an error retrieving accounts")
		assert.NotEmpty(t, accounts, "should return some accounts")
	})
}

func TestPlaidClient_GetAllTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		accountId := testutils.GetAccountIdForTest(t)

		accessToken := gofakeit.UUID()

		account := mock_plaid.BankAccountFixture(t)

		end := time.Now()
		start := end.Add(-7 * 24 * time.Hour)
		mock_plaid.MockGetRandomTransactions(t, start, end, 5000, []string{
			account.GetAccountId(),
		})

		platypus := NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		link := &models.Link{
			LinkId:    1234,
			AccountId: accountId,
		}

		client, err := platypus.NewClient(context.Background(), link, accessToken)
		assert.NoError(t, err, "should create platypus")
		assert.NotNil(t, client, "should not be nil")

		transactions, err := client.GetAllTransactions(context.Background(), start, end, []string{
			account.GetAccountId(),
		})
		assert.NoError(t, err, "should not return an error")
		assert.NotEmpty(t, transactions, "should return a few transactions")
		assert.Equal(t, map[string]int{
			"POST https://sandbox.plaid.com/transactions/get": 11,
		}, httpmock.GetCallCountInfo(), "API calls should match")
	})
}

func TestPlaidClient_UpdateItem(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		accountId := testutils.GetAccountIdForTest(t)

		accessToken := gofakeit.UUID()

		mock_plaid.MockCreateLinkToken(t)

		platypus := NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
			OAuthDomain:  "localhost",
		})

		link := &models.Link{
			LinkId:    1234,
			AccountId: accountId,
		}

		client, err := platypus.NewClient(context.Background(), link, accessToken)
		assert.NoError(t, err, "should create client")
		assert.NotNil(t, client, "should not be nil")

		linkToken, err := client.UpdateItem(context.Background())
		assert.NoError(t, err, "should not return an error creating an update link token")
		assert.NotEmpty(t, linkToken.Token(), "must not be empty")
	})
}
