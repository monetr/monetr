package controller_test

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_lunch_flow"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
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
				"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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
				"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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
				"lunchFlowURL": "lunchflow.app/api/v1",
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
				"lunchFlowURL": "https://www.lunchflow.app/api/v1?testparam=true",
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
				"lunchFlowURL": "ssh://lunchflow.app/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.lunchFlowURL").String().IsEqual("Lunch Flow API URL must be a full valid URL")
	})

	t.Run("IP address is not in the allowlist", func(t *testing.T) {
		// A link local IP like the cloud metadata endpoint is exactly the kind of
		// SSRF target the allowlist is supposed to keep out. The allowlist check
		// happens in the handler now so it comes back as a plain bad request
		// instead of a per field problem.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "Not Allowed",
				"lunchFlowURL": "http://169.254.169.254/latest/meta-data",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid Lunch Flow API URL provided")
		response.JSON().Object().NotContainsKey("problems")
	})

	t.Run("localhost is not in the allowlist", func(t *testing.T) {
		// Pointing the API URL back at ourselves should never be allowed either.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "Not Allowed",
				"lunchFlowURL": "http://localhost",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid Lunch Flow API URL provided")
		response.JSON().Object().NotContainsKey("problems")
	})

	t.Run("attacker controlled URL is not in the allowlist", func(t *testing.T) {
		// And a perfectly well formed external URL that just is not one we trust
		// should be rejected too.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "Not Allowed",
				"lunchFlowURL": "https://attacker.example.com/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid Lunch Flow API URL provided")
		response.JSON().Object().NotContainsKey("problems")
	})

	t.Run("allowlist with multiple entries accepts any", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.LunchFlow.AllowedApiUrls = []string{
			"https://www.lunchflow.app/api/v1",
			"https://lunchflow.compatible.app/api/v1",
		}
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "Staging",
				"lunchFlowURL": "https://lunchflow.compatible.app/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()
		response.Status(http.StatusOK)
	})

	t.Run("invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://www.lunchflow.app/api/v1",
				"apiKey":       "",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.apiKey").String().IsEqual("Lunch Flow API Key must be provided to setup a Lunch Flow link")
	})

	t.Run("empty name", func(t *testing.T) {
		// The name validation now comes from the shared schemas.Name helper, make
		// sure it is actually wired up. An empty string trips the Required rule on
		// that helper which has its own custom message.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "",
				"lunchFlowURL": "https://www.lunchflow.app/api/v1",
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.name").String().IsEqual("Name is required")
	})

	t.Run("api key with invalid characters", func(t *testing.T) {
		// The api key has to be letters and numbers only, anything with spaces or
		// symbols in it should be rejected by the is.UTFLetterNumeric rule.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://www.lunchflow.app/api/v1",
				"apiKey":       "not a valid key!",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.apiKey").String().IsEqual("must contain unicode letters and numbers only")
	})

	t.Run("lunch flow URL is not a string", func(t *testing.T) {
		// The schema now asserts the URL is actually a string before it tries to
		// parse it. A caller sending a number should trip the IsString rule rather
		// than blow up somewhere downstream.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": 1234,
				"apiKey":       "foobar",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.lunchFlowURL").String().IsEqual("must be a string")
	})

	t.Run("api key is not a string", func(t *testing.T) {
		// Same idea as the URL, the api key has to be a string and the IsString
		// rule should catch a non string value.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":         "US Bank",
				"lunchFlowURL": "https://www.lunchflow.app/api/v1",
				"apiKey":       1234,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.apiKey").String().IsEqual("must be a string")
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

		mock_lunch_flow.MockFetchBalance(t, "1234", lunch_flow.Balance{
			Amount:   "1234.56",
			Currency: "USD",
		})

		var id ID[LunchFlowLink]
		{
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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
			"GET https://www.lunchflow.app/api/v1/accounts":              1,
			"GET https://www.lunchflow.app/api/v1/accounts/1234/balance": 1,
		}, "must match Lunch Flow API calls")
	})

	t.Run("no accounts returned", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		// Empty account list from lunch flow!
		mock_lunch_flow.MockFetchAccounts(t, []lunch_flow.Account{})

		var id ID[LunchFlowLink]
		{
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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
			response.JSON().Array().IsEmpty()
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://www.lunchflow.app/api/v1/accounts": 1,
		}, "must match Lunch Flow API calls")
	})

	t.Run("lunch flow API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		// Simulate an error, the user gave us a bad token
		mock_lunch_flow.MockFetchAccountsError(t)

		var id ID[LunchFlowLink]
		{
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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

			response.Status(http.StatusInternalServerError)
			response.JSON().Path("$.error").String().Contains("Failed to retrieve Lunch Flow accounts:")
		}

		{ // Check for bank account in the responsne
			response := e.GET("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts").
				WithPath("lunchFlowLinkId", id).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"GET https://www.lunchflow.app/api/v1/accounts": 1,
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

func TestGetLunchFlowLinks(t *testing.T) {
	t.Run("can list lunch flow links", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var firstLink models.ID[models.LunchFlowLink]
		var secondLink models.ID[models.LunchFlowLink]

		{ // Create two lunch flow links to test listing!
			{
				response := e.POST("/api/lunch_flow/link").
					WithCookie(TestCookieName, token).
					WithJSON(map[string]any{
						"name":         "US Bank",
						"lunchFlowURL": "https://www.lunchflow.app/api/v1",
						"apiKey":       "foobar",
					}).
					Expect()

				response.Status(http.StatusOK)
				firstLink = models.ID[models.LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
			}

			{
				response := e.POST("/api/lunch_flow/link").
					WithCookie(TestCookieName, token).
					WithJSON(map[string]any{
						"name":         "Chase Bank",
						"lunchFlowURL": "https://www.lunchflow.app/api/v1",
						"apiKey":       "foobar2",
					}).
					Expect()

				response.Status(http.StatusOK)
				secondLink = models.ID[models.LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
			}
		}

		{ // Make sure that we list both links and in the order we expect
			response := e.GET("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().IsArray()
			response.JSON().Array().Length().IsEqual(2)
			response.JSON().Path("$[0].lunchFlowLinkId").IsEqual(secondLink)
			response.JSON().Path("$[1].lunchFlowLinkId").IsEqual(firstLink)
		}
	})

	t.Run("no cross user reads", func(t *testing.T) {
		_, e := NewTestApplication(t)

		{
			token := GivenIHaveToken(t, e)
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
					"apiKey":       "foobar",
				}).
				Expect()

			response.Status(http.StatusOK)
		}

		{ // Using a different token, make sure we can't read the lunch flow link
			token := GivenIHaveToken(t, e)
			response := e.GET("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().IsArray()
			response.JSON().Array().IsEmpty()
		}
	})

	t.Run("lunch flow is disabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.LunchFlow.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)
		response := e.GET("/api/lunch_flow/link").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Lunch Flow is not enabled on this server")
	})
}

func TestPostLunchFlowLinkSync(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// An active Lunch Flow link that has never been manually synced should kick
		// off a sync and come back accepted. We dont have any bank accounts on it
		// so nothing actually gets enqueued, but the endpoint still accepts the
		// request.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": link.LinkId,
			}).
			Expect()

		response.Status(http.StatusAccepted)
		response.Body().IsEmpty()
	})

	t.Run("missing link Id", func(t *testing.T) {
		// The endpoint used to just bind the body and run with a zero value link
		// Id, now it validates that a link Id was actually provided.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.linkId").String().IsEqual("required key is missing")
	})

	t.Run("link Id is not a valid Id", func(t *testing.T) {
		// A link Id that is present but does not look like one of our IDs should be
		// rejected by the ValidID rule before we ever go to the database.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": "potato",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.linkId").String().IsEqual("id does not match format link_...")
	})

	t.Run("link does not exist", func(t *testing.T) {
		// A well formed link Id that just is not in the database should come back
		// as a not found rather than blowing up.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": "link_bogusbogusbogusbogusbo",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve link: record does not exist")
	})

	t.Run("cannot sync a non-Lunch Flow link", func(t *testing.T) {
		// Manual sync only makes sense for Lunch Flow links. Pointing it at a
		// manual link should be rejected.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": link.LinkId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("cannot manually sync a non-Lunch Flow link")
	})

	t.Run("deactivated link will not sync", func(t *testing.T) {
		// A link that the user has deactivated should not be syncable. We flip the
		// status directly since theres no API to deactivate a link.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		link.LunchFlowLink.Status = LunchFlowLinkStatusDeactivated
		testutils.MustDBUpdate(t, link.LunchFlowLink)

		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": link.LinkId,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Link is not active and will not be synced")
	})

	t.Run("cannot sync too recently", func(t *testing.T) {
		// The first sync sets the last manual sync timestamp, a second one right
		// after it should be rejected because we only allow a manual sync every 30
		// minutes. The mock clock does not advance between the two calls so the
		// second is always inside that window.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		{ // The first sync should be accepted.
			response := e.POST("/api/lunch_flow/link/sync").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId": link.LinkId,
				}).
				Expect()

			response.Status(http.StatusAccepted)
		}

		{ // The second sync should be rejected because it is too soon.
			response := e.POST("/api/lunch_flow/link/sync").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId": link.LinkId,
				}).
				Expect()

			response.Status(http.StatusTooEarly)
			response.JSON().Path("$.error").String().IsEqual("Link has been manually synced too recently")
		}
	})

	t.Run("lunch flow is disabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.LunchFlow.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/lunch_flow/link/sync").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": "link_bogusbogusbogusbogusbo",
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Lunch Flow is not enabled on this server")
	})
}
