package lunch_flow_test

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/mock_lunch_flow"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestLunchFlowClient_GetAccounts(t *testing.T) {
	t.Run("happy path, retrieve a few accounts", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		accountOne := lunch_flow.Account{
			Id:              "1234",
			Name:            "Testing Checking",
			InstitutionName: "Lehman Brothers",
			Provider:        "Bogus",
			Currency:        "USD",
			Status:          "ACTIVE",
		}
		accountTwo := lunch_flow.Account{
			Id:              "1235",
			Name:            "Testing Savings",
			InstitutionName: "Lehman Brothers",
			Provider:        "Bogus",
			Currency:        "USD",
			Status:          "ACTIVE",
		}
		mock_lunch_flow.MockFetchAccounts(t, []lunch_flow.Account{
			accountOne,
			accountTwo,
		})

		log := testutils.GetLog(t)

		client, err := lunch_flow.NewLunchFlowClient(
			log,
			lunch_flow.DefaultAPIURL,
			"bogus-token",
		)
		assert.NoError(t, err, "must not return an error creating the client")
		assert.NotNil(t, client, "client must have a value")

		accounts, err := client.GetAccounts(t.Context())
		assert.NoError(t, err, "must successfully retrieve accounts")
		assert.EqualValues(t, []lunch_flow.Account{
			accountOne,
			accountTwo,
		}, accounts, "accounts response should match the expectation")

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://lunchflow.app/api/v1/accounts": 1,
		}, "must match Lunch Flow API calls")
	})

	t.Run("fail to retrieve accounts", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_lunch_flow.MockFetchAccountsError(t)

		log := testutils.GetLog(t)

		client, err := lunch_flow.NewLunchFlowClient(
			log,
			lunch_flow.DefaultAPIURL,
			"bogus-token",
		)
		assert.NoError(t, err, "must not return an error creating the client")
		assert.NotNil(t, client, "client must have a value")

		accounts, err := client.GetAccounts(t.Context())
		assert.Error(t, err, "must return an error if the request fails")
		assert.Empty(t, accounts)

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://lunchflow.app/api/v1/accounts": 1,
		}, "must match Lunch Flow API calls")
	})
}

func TestLunchFlowClient_GetBalance(t *testing.T) {
	t.Run("happy path read balance", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		expectedBalance := lunch_flow.Balance{
			Amount:   "1234.56",
			Currency: "USD",
		}
		mock_lunch_flow.MockFetchBalance(t, "1234", expectedBalance)

		log := testutils.GetLog(t)

		client, err := lunch_flow.NewLunchFlowClient(
			log,
			lunch_flow.DefaultAPIURL,
			"bogus-token",
		)
		assert.NoError(t, err, "must not return an error creating the client")
		assert.NotNil(t, client, "client must have a value")

		balance, err := client.GetBalance(t.Context(), "1234")
		assert.NoError(t, err, "must successfully retrieve accounts")
		assert.NotNil(t, balance, "balance cannot be nil here")
		assert.EqualValues(t, expectedBalance, *balance, "balance response should match expectation")

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://lunchflow.app/api/v1/accounts/1234/balance": 1,
		}, "must match Lunch Flow API calls")
	})

	t.Run("fails to read balance", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_lunch_flow.MockFetchBalanceError(t, "1234")

		log := testutils.GetLog(t)

		client, err := lunch_flow.NewLunchFlowClient(
			log,
			lunch_flow.DefaultAPIURL,
			"bogus-token",
		)
		assert.NoError(t, err, "must not return an error creating the client")
		assert.NotNil(t, client, "client must have a value")

		balance, err := client.GetBalance(t.Context(), "1234")
		assert.Error(t, err, "must get an error if the request fails")
		assert.Nil(t, balance, "balance must be nil in an error path")

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://lunchflow.app/api/v1/accounts/1234/balance": 1,
		}, "must match Lunch Flow API calls")
	})
}
