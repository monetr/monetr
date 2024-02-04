package teller_test

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/mock_teller"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/teller"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	log := testutils.GetLog(t)
	client, err := teller.NewClient(log, config.Teller{
		Enabled:       true,
		ApplicationId: "app_abc123",
	})
	assert.NoError(t, err, "must not have an error creating a client without a certificate")
	assert.NoError(t, client.GetHealth(context.Background()), "must pass health check")
}

func TestGetAccounts(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)

		account := mock_teller.BankAccountFixture(t)
		mock_teller.MockGetAccounts(t, []teller.Account{
			account,
		})

		tellerClient, err := teller.NewClient(log, config.Teller{
			Environment:   "sandbox",
			ApplicationId: "app_test123",
		})
		assert.NoError(t, err, "must be able to create the basic client")

		client := tellerClient.GetAuthenticatedClient("token_test123")
		accounts, err := client.GetAccounts(context.Background())
		assert.NoError(t, err, "should not return an error fetching accounts")
		assert.NotEmpty(t, accounts, "should have an account in the result")
	})
}
