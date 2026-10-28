package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/internal/mock_simplefin"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestPostSimpleFinLink(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/happy")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.simpleFinLinkId").String().IsASCII()
		response.JSON().Path("$.name").String().IsEqual("SimpleFIN Bridge")
		response.JSON().Path("$.status").String().IsEqual("pending")
		response.JSON().Path("$.apiUrl").String().IsEqual("https://bridge.simplefin.org/simplefin")
		response.JSON().Object().Keys().NotContainsAll("secretId", "secret")
		response.Body().NotContains(mock_simplefin.TestPassword)
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should have claimed the setup token once")
	})

	t.Run("with a valid api key", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/apikey")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.POST("/api/simplefin/link").
			WithBasicAuth(apiKeyId, apiKeySecret).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.simpleFinLinkId").String().IsASCII()
		response.JSON().Path("$.status").String().IsEqual("pending")
		response.JSON().Object().Keys().NotContainsAll("secretId", "secret")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/badkey")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		response := e.POST("/api/simplefin/link").
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have claimed the setup token")
	})

	t.Run("simplefin is disabled", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/disabled")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		config := NewTestApplicationConfig(t)
		config.SimpleFin.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("SimpleFIN is not enabled on this server")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have claimed the setup token")
	})

	t.Run("no allowed hosts", func(t *testing.T) {
		// Enabled but with an empty allowlist still counts as disabled
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/nohosts")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		config := NewTestApplicationConfig(t)
		config.SimpleFin.Enabled = true
		config.SimpleFin.AllowedHosts = []string{}
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("SimpleFIN is not enabled on this server")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have claimed the setup token")
	})

	t.Run("invalid setup token", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": "this is not base64!",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid SimpleFIN setup token")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have made any requests")
	})

	t.Run("host not allowed", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := "https://evil.example.com/simplefin/claim/evil"
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("SimpleFIN setup token is for a server that is not allowed")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have claimed the setup token")
	})

	t.Run("http claim url", func(t *testing.T) {
		// The host is allowed but plain http never is, SimpleFIN sends credentials
		// with every request
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := "http://bridge.simplefin.org/simplefin/claim/http"
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("SimpleFIN setup token is for a server that is not allowed")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have claimed the setup token")
	})

	t.Run("already claimed", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/used")
		mock_simplefin.MockClaimSetupTokenAlreadyClaimed(t, claimUrl)

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("SimpleFIN setup token has already been used")
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should have tried to claim the setup token")
	})

	t.Run("access url on another host", func(t *testing.T) {
		// The claim is on an allowed host, but what it gives back points somewhere
		// else. We should not store that.
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/otherhost")
		mock_simplefin.MockClaimSetupToken(
			t,
			claimUrl,
			"https://"+mock_simplefin.TestUsername+":"+mock_simplefin.TestPassword+"@evil.example.com/simplefin",
		)

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("SimpleFIN setup token is for a server that is not allowed")
		response.Body().NotContains(mock_simplefin.TestPassword)
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should have claimed the setup token")
	})

	t.Run("upstream error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/broken")
		httpmock.RegisterResponder(
			http.MethodPost,
			claimUrl,
			httpmock.NewStringResponder(http.StatusInternalServerError, "something broke upstream"),
		)

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":       "SimpleFIN Bridge",
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().IsEqual("Failed to claim SimpleFIN setup token")
		response.Body().NotContains("something broke upstream")
	})

	t.Run("missing name", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/noname")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"setupToken": mock_simplefin.SetupToken(t, claimUrl),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.name").String().IsEqual("required key is missing")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have claimed the setup token")
	})

	t.Run("missing setup token", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/simplefin/link").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "SimpleFIN Bridge",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.setupToken").String().IsEqual("required key is missing")
		assert.Zero(t, httpmock.GetTotalCallCount(), "should not have made any requests")
	})
}

func TestGetSimpleFinLinks(t *testing.T) {
	t.Run("can list simplefin links", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		firstClaimUrl := mock_simplefin.Path(t, "/simplefin/claim/first")
		mock_simplefin.MockClaimSetupToken(t, firstClaimUrl, mock_simplefin.AccessUrl(t))
		secondClaimUrl := mock_simplefin.Path(t, "/simplefin/claim/second")
		mock_simplefin.MockClaimSetupToken(t, secondClaimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var firstLink ID[SimpleFinLink]
		var secondLink ID[SimpleFinLink]

		{ // Create two SimpleFIN links to test listing
			{
				response := e.POST("/api/simplefin/link").
					WithCookie(TestCookieName, token).
					WithJSON(map[string]any{
						"name":       "SimpleFIN Bridge",
						"setupToken": mock_simplefin.SetupToken(t, firstClaimUrl),
					}).
					Expect()

				response.Status(http.StatusOK)
				firstLink = ID[SimpleFinLink](response.JSON().Path("$.simpleFinLinkId").String().Raw())
			}

			{
				response := e.POST("/api/simplefin/link").
					WithCookie(TestCookieName, token).
					WithJSON(map[string]any{
						"name":       "Other SimpleFIN Bridge",
						"setupToken": mock_simplefin.SetupToken(t, secondClaimUrl),
					}).
					Expect()

				response.Status(http.StatusOK)
				secondLink = ID[SimpleFinLink](response.JSON().Path("$.simpleFinLinkId").String().Raw())
			}
		}

		{ // Make sure that we list both links and in the order we expect
			response := e.GET("/api/simplefin/link").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().IsArray()
			response.JSON().Array().Length().IsEqual(2)
			response.JSON().Path("$[0].simpleFinLinkId").IsEqual(secondLink)
			response.JSON().Path("$[1].simpleFinLinkId").IsEqual(firstLink)
			response.Body().NotContains(mock_simplefin.TestPassword)
		}
	})

	t.Run("get by id", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/byid")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var id ID[SimpleFinLink]
		{ // Create the SimpleFIN link
			response := e.POST("/api/simplefin/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":       "SimpleFIN Bridge",
					"setupToken": mock_simplefin.SetupToken(t, claimUrl),
				}).
				Expect()

			response.Status(http.StatusOK)
			id = ID[SimpleFinLink](response.JSON().Path("$.simpleFinLinkId").String().Raw())
			assert.NotZero(t, id, "must be able to extract the SimpleFIN link ID")
		}

		{ // Retrieve it with an API key
			apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)
			response := e.GET("/api/simplefin/link/{simpleFinLinkId}").
				WithPath("simpleFinLinkId", id).
				WithBasicAuth(apiKeyId, apiKeySecret).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.simpleFinLinkId").IsEqual(id)
			response.JSON().Path("$.name").String().IsEqual("SimpleFIN Bridge")
			response.JSON().Path("$.status").String().IsEqual("pending")
			response.JSON().Path("$.apiUrl").String().IsEqual("https://bridge.simplefin.org/simplefin")
			response.Body().NotContains(mock_simplefin.TestPassword)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.GET("/api/simplefin/link/{simpleFinLinkId}").
			WithPath("simpleFinLinkId", NewID[SimpleFinLink]()).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Failed to retrieve SimpleFIN link: record does not exist")
	})

	t.Run("no cross user reads", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		claimUrl := mock_simplefin.Path(t, "/simplefin/claim/crossuser")
		mock_simplefin.MockClaimSetupToken(t, claimUrl, mock_simplefin.AccessUrl(t))

		_, e := NewTestApplication(t)

		var id ID[SimpleFinLink]
		{ // Create the SimpleFIN link as the first user
			token := GivenIHaveToken(t, e)
			response := e.POST("/api/simplefin/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":       "SimpleFIN Bridge",
					"setupToken": mock_simplefin.SetupToken(t, claimUrl),
				}).
				Expect()

			response.Status(http.StatusOK)
			id = ID[SimpleFinLink](response.JSON().Path("$.simpleFinLinkId").String().Raw())
			assert.NotZero(t, id, "must be able to extract the SimpleFIN link ID")
		}

		token := GivenIHaveToken(t, e)
		{ // Using a different token, make sure we can't list the SimpleFIN link
			response := e.GET("/api/simplefin/link").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().IsArray()
			response.JSON().Array().IsEmpty()
		}

		{ // Or retrieve it directly
			response := e.GET("/api/simplefin/link/{simpleFinLinkId}").
				WithPath("simpleFinLinkId", id).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("Failed to retrieve SimpleFIN link: record does not exist")
		}
	})

	t.Run("simplefin is disabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.SimpleFin.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)

		{
			response := e.GET("/api/simplefin/link").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("SimpleFIN is not enabled on this server")
		}

		{
			response := e.GET("/api/simplefin/link/{simpleFinLinkId}").
				WithPath("simpleFinLinkId", NewID[SimpleFinLink]()).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("SimpleFIN is not enabled on this server")
		}
	})
}
