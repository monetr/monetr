package simplefin_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/datasources/simplefin"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/mock_simplefin"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestClaimSetupToken(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/abc123")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			mock_simplefin.SetupToken(t, claimUrl),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "should claim the setup token")
		assert.Equal(t, mock_simplefin.AccessUrl(t), accessUrl, "should return the access url")

		assert.EqualValues(t, map[string]int{
			"POST https://bridge.simplefin.org/simplefin/claim/abc123": 1,
		}, httpmock.GetCallCountInfo(), "must match SimpleFIN API calls")
	})

	t.Run("token without padding", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// This claim url is long enough that the base64 has padding on the end
		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/abc1234")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		setupToken := mock_simplefin.SetupToken(t, claimUrl)
		assert.True(t, strings.HasSuffix(setupToken, "="), "must have padding to test without it")

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			strings.TrimRight(setupToken, "="),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "should claim the setup token")
		assert.Equal(t, mock_simplefin.AccessUrl(t), accessUrl, "should return the access url")
	})

	t.Run("invalid token", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			"this is not a setup token!",
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "invalid SimpleFIN setup token")
		assert.Empty(t, accessUrl, "should not return an access url")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not make any requests")
	})

	t.Run("host not allowed", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			mock_simplefin.SetupToken(t, "https://169.254.169.254/simplefin/claim/abc123"),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "SimpleFIN host is not in the configured allowlist")
		assert.Empty(t, accessUrl, "should not return an access url")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not make any requests")
	})

	t.Run("already claimed", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/abc123")
		mock_simplefin.MockClaimSetupTokenAlreadyClaimed(t, claimUrl)

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			mock_simplefin.SetupToken(t, claimUrl),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "SimpleFIN setup token has already been claimed")
		assert.Empty(t, accessUrl, "should not return an access url")
	})

	t.Run("access url on another host", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/abc123")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, "https://user:password@169.254.169.254/simplefin")

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			mock_simplefin.SetupToken(t, claimUrl),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "SimpleFIN host is not in the configured allowlist")
		assert.Empty(t, accessUrl, "should not return an access url")
	})

	t.Run("access url without credentials", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/abc123")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.Path(t, "/simplefin"))

		log := testutils.GetLog(t)
		accessUrl, err := simplefin.ClaimSetupToken(
			t.Context(),
			log,
			mock_simplefin.SetupToken(t, claimUrl),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "SimpleFIN access URL is missing credentials")
		assert.Empty(t, accessUrl, "should not return an access url")
	})
}

func TestNewSimpleFinClient(t *testing.T) {
	t.Run("host not allowed", func(t *testing.T) {
		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			"https://user:password@169.254.169.254/simplefin",
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "SimpleFIN host is not in the configured allowlist")
		assert.Nil(t, client, "should not return a client")
	})

	t.Run("missing credentials", func(t *testing.T) {
		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			mock_simplefin.Path(t, "/simplefin"),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.EqualError(t, err, "SimpleFIN access URL is missing credentials")
		assert.Nil(t, client, "should not return a client")
	})
}

func TestSimpleFinClient_GetAccounts(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		accountSet := simplefin.AccountSet{
			Errlist: []simplefin.Error{
				{
					Code:         "con.auth",
					Message:      "You need to reauthenticate with Lehman Brothers",
					ConnectionId: new("CON-1234"),
					AccountId:    nil,
				},
			},
			Errors: []string{
				"You need to reauthenticate with Lehman Brothers",
			},
			Connections: []simplefin.Connection{
				{
					ConnectionId: "CON-1234",
					Name:         "Lehman Brothers",
				},
			},
			Accounts: []simplefin.Account{
				{
					Id:               "ACT-1234",
					Name:             "Testing Checking",
					ConnectionId:     "CON-1234",
					Currency:         "USD",
					Balance:          "1234.56",
					AvailableBalance: new(json.Number("1200.00")),
					BalanceDate:      1729728000,
					Transactions: []simplefin.Transaction{
						{
							Id:           "TRN-1234",
							Posted:       1729641600,
							Amount:       "-12.34",
							Description:  "Coffee",
							TransactedAt: new(int64(1729600000)),
							Pending:      false,
						},
					},
				},
				{
					Id:               "ACT-1235",
					Name:             "Testing Savings",
					ConnectionId:     "CON-1234",
					Currency:         "USD",
					Balance:          "100.00",
					AvailableBalance: nil,
					BalanceDate:      1729728000,
					Transactions:     nil,
				},
			},
		}
		mock_simplefin.MockFetchAccounts(t, accountSet)

		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			mock_simplefin.AccessUrl(t),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "must be able to create the client")
		assert.NotNil(t, client, "client must have a value")

		result, err := client.GetAccounts(t.Context(), simplefin.GetAccountsOptions{})
		assert.NoError(t, err, "should retrieve accounts")
		assert.EqualValues(t, &accountSet, result, "account set should match the expectation")

		assert.EqualValues(t, map[string]int{
			"GET https://bridge.simplefin.org/simplefin/accounts": 1,
		}, httpmock.GetCallCountInfo(), "must match SimpleFIN API calls")
	})

	t.Run("start date and balances only", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_http_helper.NewHttpMockJsonResponder(
			t,
			http.MethodGet,
			mock_simplefin.Path(t, "/simplefin/accounts"),
			func(t *testing.T, request *http.Request) (any, int) {
				query := request.URL.Query()
				assert.Equal(t, "2", query.Get("version"), "should request version 2")
				assert.Equal(t, "1729728000", query.Get("start-date"), "should send the start date as epoch seconds")
				assert.Equal(t, "1", query.Get("balances-only"), "should request balances only")
				return simplefin.AccountSet{}, http.StatusOK
			},
			nil,
		)

		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			mock_simplefin.AccessUrl(t),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "must be able to create the client")

		_, err = client.GetAccounts(t.Context(), simplefin.GetAccountsOptions{
			StartDate:    time.Unix(1729728000, 0),
			BalancesOnly: true,
		})
		assert.NoError(t, err, "should retrieve accounts")
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "should make one request")
	})

	t.Run("payment required", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccountsError(t, http.StatusPaymentRequired)

		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			mock_simplefin.AccessUrl(t),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "must be able to create the client")

		result, err := client.GetAccounts(t.Context(), simplefin.GetAccountsOptions{})
		assert.EqualError(t, err, "SimpleFIN requires payment")
		assert.Nil(t, result, "should not return an account set")
	})

	t.Run("bad credentials", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{})

		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			"https://"+mock_simplefin.TestUsername+":wrong-password@bridge.simplefin.org/simplefin",
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "must be able to create the client")

		result, err := client.GetAccounts(t.Context(), simplefin.GetAccountsOptions{})
		assert.EqualError(t, err, "SimpleFIN access was denied")
		assert.Nil(t, result, "should not return an account set")
	})

	t.Run("server error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccountsError(t, http.StatusInternalServerError)

		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			mock_simplefin.AccessUrl(t),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "must be able to create the client")

		// The error has the request URL in it, it should not have the credentials
		result, err := client.GetAccounts(t.Context(), simplefin.GetAccountsOptions{})
		assert.EqualError(t, err, "SimpleFIN request failed https://bridge.simplefin.org/simplefin/accounts?version=2 [500]: \"Internal Server Error\"\n")
		assert.NotContains(t, err.Error(), mock_simplefin.TestPassword, "error should not have the password")
		assert.Nil(t, result, "should not return an account set")
	})
}

func TestSimpleFinClient_Redirects(t *testing.T) {
	t.Run("redirects are blocked", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_http_helper.NewHttpMockJsonResponder(
			t,
			http.MethodGet,
			mock_simplefin.Path(t, "/simplefin/accounts"),
			func(_ *testing.T, _ *http.Request) (any, int) {
				return map[string]any{}, http.StatusFound
			},
			func(_ *testing.T, _ *http.Request, _ any, _ int) map[string][]string {
				return map[string][]string{
					"Location": {
						"https://169.254.169.254/latest/meta-data",
					},
				}
			},
		)

		log := testutils.GetLog(t)
		client, err := simplefin.NewSimpleFinClient(
			log,
			mock_simplefin.AccessUrl(t),
			testutils.GetConfig(t).SimpleFin,
		)
		assert.NoError(t, err, "must be able to create the client")

		result, err := client.GetAccounts(t.Context(), simplefin.GetAccountsOptions{})
		assert.EqualError(t, err, "failed to send HTTP request: Get \"https://169.254.169.254/latest/meta-data\": blocked SimpleFIN redirect to https://169.254.169.254/latest/meta-data")
		assert.Nil(t, result, "should not return an account set")

		assert.EqualValues(t, map[string]int{
			"GET https://bridge.simplefin.org/simplefin/accounts": 1,
		}, httpmock.GetCallCountInfo(), "should not follow the redirect")
	})
}
