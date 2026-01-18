package controller_test

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/mock_lunch_flow"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestPostLunchFlowLink(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://lunchflow.com/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.lunchFlowLinkId").String().IsASCII()
		response.JSON().Path("$.status").String().IsEqual("pending")
		response.JSON().Object().Keys().NotContainsAll("secretId", "secret")
	})

	t.Run("lunch flow is disabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.LunchFlow.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://lunchflow.com/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Lunch Flow is not enabled on this server")
	})

	t.Run("invalid API URL, no protocol", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "lunchflow.com/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.lunchFlowURL").String().IsEqual("Lunch Flow API URL must be a full valid URL")
	})

	t.Run("invalid API URL, query params", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://lunchflow.com/api/v1?testparam=true",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.lunchFlowURL").String().IsEqual("Lunch Flow API URL must be a full valid URL")
	})

	t.Run("invalid API URL, scheme", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "ssh://lunchflow.com/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.lunchFlowURL").String().IsEqual("Lunch Flow API URL must be a full valid URL")
	})

	t.Run("invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://lunchflow.com/api/v1",
				"apiKey":       "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.apiKey").String().IsEqual("Lunch Flow API Key must be provided to setup a Lunch Flow link")
	})
}

func TestPostLunchFlowLinkBankAccountsRefresh(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		mock_lunch_flow.MockFetchAccounts(t, []lunch_flow.Account{
			{
				Id:              "1234",
				Name:            "Main Account",
				InstitutionName: "Finance",
				InstitutionLogo: nil,
				Provider:        "gocardless",
				Status:          "ACTIVE",
			},
		})

		var id ID[LunchFlowLink]
		{
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://lunchflow.com/api/v1",
					"apiKey":       "foobar",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.lunchFlowLinkId").String().IsASCII()
			response.JSON().Path("$.status").String().IsEqual("pending")
			id = ID[LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
		}

		{ // Refresh the accounts
			response := e.POST("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts/refresh").
				WithPath("lunchFlowLinkId", id).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNoContent)
			response.Body().IsEmpty()
		}

		{ // Check for bank account in the responsne
			response := e.GET("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts").
				WithPath("lunchFlowLinkId", id).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(1)
			response.JSON().Path("$[0].lunchFlowBankAccountId").String().IsASCII()
			response.JSON().Path("$[0].lunchFlowId").String().IsEqual("1234")
			response.JSON().Path("$[0].name").String().IsEqual("Main Account")
			response.JSON().Path("$[0].institutionName").String().IsEqual("Finance")
			response.JSON().Path("$[0].provider").String().IsEqual("gocardless")
			response.JSON().Path("$[0].lunchFlowStatus").String().IsEqual("ACTIVE")
			response.JSON().Path("$[0].status").String().IsEqual("inactive")
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://lunchflow.com/api/v1/accounts": 1,
		}, "must match Lunch Flow API calls")
	})

	t.Run("lunch flow is disabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.LunchFlow.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts/refresh").
			WithPath("lunchFlowLinkId", "bogus").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Lunch Flow is not enabled on this server")
	})
}
