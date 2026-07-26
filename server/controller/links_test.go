package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/links/link_jobs"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     "My personal link",
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").String().IsASCII()
		response.JSON().Path("$.linkType").IsEqual(models.ManualLinkType)
		response.JSON().Path("$.institutionName").String().NotEmpty()
		response.JSON().Path("$.description").String().IsEqual("My personal link")
	})

	t.Run("trims whitespace", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "  U.S. Bank  ",
				"description":     "  My personal link  ",
			}).
			Expect()

		// The schema parsing trims string fields before they are validated or
		// stored, so the surrounding whitespace should be gone in the response.
		response.Status(http.StatusOK)
		response.JSON().Path("$.institutionName").String().IsEqual("U.S. Bank")
		response.JSON().Path("$.description").String().IsEqual("My personal link")
	})

	t.Run("description can be explicitly null", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     nil,
			}).
			Expect()

		// A null description is a valid branch of the description validation, it
		// should just leave the link without a description.
		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").String().IsASCII()
		response.JSON().Path("$.description").IsNull()
	})

	t.Run("cannot set fields that are not part of the schema", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		// The handler forces the link type to manual and the schema does not allow
		// a linkType key at all, so trying to sneak one in should be rejected
		// outright rather than silently ignored.
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"linkType":        models.PlaidLinkType,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.linkType").String().IsEqual("key not expected")
	})

	t.Run("institution name is too long", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": gofakeit.Sentence(301),
				"description":     "My personal link",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.institutionName").String().IsEqual("Name must be between 1 and 300 characters")
	})

	t.Run("name not provided", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "",
				"description":     "My personal link",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.institutionName").String().IsEqual("Name is required")
	})

	t.Run("description is too long", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     gofakeit.Sentence(301),
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		// The description is now validated as one of a nil value or a valid text
		// field, so a too long description fails both branches and the problem
		// comes back as a oneOf envelope rather than a flat string.
		response.JSON().Path("$.problems.description.oneOf").Array().ConsistsOf(
			"must be nil",
			"Must be between 1 and 300 characters",
		)
	})

	t.Run("malformed json", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I'm not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("failed to parse request")
	})

	t.Run("unauthenticated", func(t *testing.T) {
		app, e := NewTestApplication(t)
		link := models.Link{
			LinkId:          "link_bogus",         // Set it to something so we can verify its different in the result.
			LinkType:        models.PlaidLinkType, // This should be changed to manual in the response.
			InstitutionName: "U.S. Bank",
			CreatedAt:       app.Clock.Now().Add(-1 * time.Hour), // Set these to something to make sure it gets overwritten.
			UpdatedAt:       app.Clock.Now().Add(1 * time.Hour),
		}

		response := e.POST("/api/links").
			WithJSON(link).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("lunch flow link id is malformed", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     "My personal link",
				"lunchFlowLinkId": "lfx_bogus",
			}).
			Expect()

		// A malformed lunch flow link Id is now caught by the schema before we ever
		// try to look it up, so we get a validation error instead of a not found.
		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.lunchFlowLinkId.oneOf").Array().ConsistsOf(
			"must be nil",
			"id should be between 28 and 32 characters",
		)
	})

	t.Run("lunch flow link does not exist", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		// This is a well formed Id so it passes validation, but it does not point
		// at any real lunch flow link so the lookup should come back empty.
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     "My personal link",
				"lunchFlowLinkId": models.NewID[models.LunchFlowLink](),
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Failed to retrieve lunch flow link: record does not exist")
	})

	t.Run("lunch flow is not enabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.LunchFlow.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)

		token := GivenIHaveToken(t, e)
		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     "My personal link",
				"lunchFlowLinkId": models.NewID[models.LunchFlowLink](),
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Lunch Flow is not enabled on this server")
	})

	t.Run("lunch flow link", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var lunchFlowLinkId models.ID[models.LunchFlowLink]
		{ // Create the lunch flow link first!
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
					"apiKey":       "foobar",
				}).
				Expect()

			response.Status(http.StatusOK)
			// Link should be in pending when its first created
			response.JSON().Path("$.status").IsEqual(models.LunchFlowLinkStatusPending)
			lunchFlowLinkId = models.ID[models.LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
		}

		{ // Then create the actual link!
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
					"lunchFlowLinkId": lunchFlowLinkId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.linkType").IsEqual(models.LunchFlowLinkType)
			response.JSON().Path("$.lunchFlowLinkId").String().IsEqual(lunchFlowLinkId.String())
		}

		{ // Then check the status of the lunch flow link to make sure its active
			response := e.GET("/api/lunch_flow/link/{lunchFlowLinkId}").
				WithPath("lunchFlowLinkId", lunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.status").IsEqual(models.LunchFlowLinkStatusActive)
		}

		{ // Make sure we can't reuse the lunch flow link!
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
					"lunchFlowLinkId": lunchFlowLinkId,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Cannot create a link from a Lunch Flow link that is not in a pending status")
		}
	})

	t.Run("with a valid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		// An API key created from this session belongs to the same account, so it
		// is allowed to create a link exactly like the session token can. This
		// endpoint lives in the billedKeyOrToken route group.
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)
		response := e.POST("/api/links").
			WithBasicAuth(apiKeyId, apiKeySecret).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     "My personal link",
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").String().IsASCII()
		response.JSON().Path("$.linkType").IsEqual(models.ManualLinkType)
		response.JSON().Path("$.institutionName").String().NotEmpty()
		response.JSON().Path("$.description").String().IsEqual("My personal link")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// A well formed but completely unknown API key must be rejected by the auth
		// middleware before the request body is ever considered.
		response := e.POST("/api/links").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			WithJSON(map[string]any{
				"institutionName": "U.S. Bank",
				"description":     "My personal link",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestGetLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		_, e := NewTestApplication(t)

		tokenA, tokenB := GivenIHaveToken(t, e), GivenIHaveToken(t, e)

		// We want to create a link with tokenA. This link should not be visible later when we request the link for
		// tokenB. This will help verify that we do not expose data from someone else's login.
		var linkAID models.ID[models.Link]
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenA).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.linkType").IsEqual(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			// Even if we specify a Plaid link type, it shouldn't be; so we should not
			// see a plaid link on the result.
			response.JSON().Object().Keys().NotContainsAll("plaidLink")
			linkAID = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		// Create a link for tokenB too. This way we can do a GET request for both tokens to test each scenario.
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenB).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.linkType").IsEqual(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
		}

		// Now we want to test GET with token A.
		{
			response := e.GET("/api/links").
				WithCookie(TestCookieName, tokenA).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$").Array().Length().IsEqual(1)
			response.JSON().Path("$[0].linkId").IsEqual(linkAID)
		}

		// Now we want to test GET with token B.
		{
			response := e.GET("/api/links").
				WithCookie(TestCookieName, tokenB).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$").Array().Length().IsEqual(1)
			// Make sure that we do not receive token A's link.
			response.JSON().Path("$[0].linkId").NotEqual(linkAID)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		_, e := NewTestApplication(t)
		response := e.GET("/api/links").
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("precise", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		institutionName := "U.S. Bank"

		var linkId models.ID[models.Link]
		{ // Create the link.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": institutionName,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.institutionName").String().IsEqual(institutionName)

			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Retrieve the link and make sure the linkId matches.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(linkId)
		}
	})

	t.Run("precise not found", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		{ // Try to retrieve a link that does not exist for this user.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", "link_bogus").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("failed to retrieve link: record does not exist")
		}
	})

	t.Run("precise bad Id", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		{ // Try to retrieve a link that does not exist for this user.
			response := e.GET("/api/links/0").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("must specify a valid link Id to retrieve")
		}
	})

	t.Run("plaid link", func(t *testing.T) {
		app, e := NewTestApplication(t)

		var token string
		var linkId models.ID[models.Link]
		{
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
			linkId = link.LinkId
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		// When we have a real plaid link, there will be a plaid link sub object.
		response.JSON().Path("$.plaidLink.institutionId").String().NotEmpty()
	})

	t.Run("cant get someone elses link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var link models.Link

		{ // Create a link under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link = fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to retrieve the link
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve link: record does not exist")
		}
	})

	t.Run("list with a valid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId models.ID[models.Link]
		{ // Seed a link on this account using the session token.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		// An API key on the same account should list exactly the links the session
		// token would. GET /api/links is in the billedKeyOrToken route group.
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)
		response := e.GET("/api/links").
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$").Array().Length().IsEqual(1)
		response.JSON().Path("$[0].linkId").IsEqual(linkId)
	})

	t.Run("list with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// A well formed but unknown API key must not be able to list any links.
		response := e.GET("/api/links").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})

	t.Run("precise with a valid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		institutionName := "U.S. Bank"

		var linkId models.ID[models.Link]
		{ // Create the link with the session token.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": institutionName,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.institutionName").String().IsEqual(institutionName)
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)
		{ // Retrieve the link with the API key and make sure the linkId matches.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithBasicAuth(apiKeyId, apiKeySecret).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(linkId)
		}
	})

	t.Run("precise with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// The auth check happens before we ever look up the link, so a bogus Id is
		// fine here; the unknown key is what gets rejected.
		response := e.GET("/api/links/{linkId}").
			WithPath("linkId", "link_bogus").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestPatchLink(t *testing.T) {
	t.Run("simple manual link", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId models.ID[models.Link]
		{ // Create the manual link via the API
			institutionName := "U.S. Bank"

			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": institutionName,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.institutionName").String().IsEqual(institutionName)
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Then update the link with a patch request
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "My Own Name",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(linkId)
			response.JSON().Path("$.institutionName").String().IsEqual("My Own Name")
		}

	})

	t.Run("simple plaid link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Then update the link with a patch request
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "My Own Name",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(link.LinkId)
			response.JSON().Path("$.institutionName").String().IsEqual("My Own Name")
		}
	})

	t.Run("cant update any other fields", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		token := GivenILogin(t, e, user.Login.Email, password)

		{ // Then update the link with a patch request
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkType": models.ManualLinkType,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Invalid request")
			response.JSON().Path("$.problems.linkType").String().IsEqual("key not expected")
		}
	})

	t.Run("cant patch someone elses link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var link models.Link

		{ // Create a link under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link = fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to patch the link
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"description": "my updated description",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve link: record does not exist")
		}
	})

	t.Run("update the description", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId models.ID[models.Link]
		{ // Create the manual link via the API
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Patch the description with whitespace on either side to make sure it
			// gets trimmed
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"description": "  My bank login  ",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(linkId)
			// The surrounding whitespace should have been stripped before we stored
			// it.
			response.JSON().Path("$.description").String().IsEqual("My bank login")
		}
	})

	t.Run("clear the description", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId models.ID[models.Link]
		{ // Create the manual link with a description already set
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "Something I want to remove later",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.description").String().IsEqual("Something I want to remove later")
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Now send null for the description to clear it out entirely
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"description": nil,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(linkId)
			// Passing null should null out the description and not leave the old
			// value in place.
			response.JSON().Path("$.description").IsNull()
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId models.ID[models.Link]
		{ // Create a link so that we have something real to try to patch later
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
				}).
				Expect()

			response.Status(http.StatusOK)
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Try to patch the link without providing a token at all
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithJSON(map[string]any{
					"description": "I should not be allowed to do this",
				}).
				Expect()

			response.Status(http.StatusUnauthorized)
			response.JSON().Path("$.error").String().IsEqual("unauthorized")
		}
	})

	t.Run("with a valid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId models.ID[models.Link]
		{ // Create the manual link via the API with the session token
			institutionName := "U.S. Bank"

			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": institutionName,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.institutionName").String().IsEqual(institutionName)
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)
		{ // Update the link with a patch request authenticated by the API key.
			// PATCH /api/links/{linkId} is in the billedKeyOrToken route group.
			response := e.PATCH("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithBasicAuth(apiKeyId, apiKeySecret).
				WithJSON(map[string]any{
					"institutionName": "My Own Name",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").IsEqual(linkId)
			response.JSON().Path("$.institutionName").String().IsEqual("My Own Name")
		}
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// A well formed but unknown API key must be rejected before the patch is
		// ever applied.
		response := e.PATCH("/api/links/{linkId}").
			WithPath("linkId", "link_bogus").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			WithJSON(map[string]any{
				"institutionName": "My Own Name",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestDeleteLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		clock := clock.New()
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		institutionName := "U.S. Bank"

		var linkId models.ID[models.Link]
		{ // Create the link.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": institutionName,
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.institutionName").String().IsEqual(institutionName)

			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(link_jobs.RemoveLink),
				gomock.Any(),
				gomock.Eq(link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    linkId,
				}),
			).
			MaxTimes(1).
			Return(nil)

		{ // Try to retrieve the link before it's been deleted.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Object().NotEmpty()
			response.JSON().Path("$.deletedAt").IsNull()
		}

		{ // Try to delete it.
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				WithTimeout(5 * time.Second).
				Expect()

			response.Status(http.StatusOK)
			response.NoContent()
		}

		{ // Try to retrieve the link after it's been deleted.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.deletedAt").NotNull()
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{}, "should not have made ANY plaid API calls")
	})

	t.Run("remove plaid link", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		clock := clock.New()
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, clock)
		token := GivenILogin(t, e, user.Login.Email, password)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		{ // Retrieve the link and do some tests to make sure its a plaid link
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Object().NotEmpty()
			response.JSON().Path("$.deletedAt").IsNull()
			response.JSON().Path("$.plaidLink").Object().NotEmpty()
			response.JSON().Path("$.plaidLink.status").IsEqual(models.PlaidLinkStatusSetup)
			response.JSON().Path("$.plaidLink.deletedAt").IsNull()
			response.JSON().Path("$.linkType").IsEqual(models.PlaidLinkType)
		}

		mock_plaid.MockDeactivateItemTokenSuccess(t)

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(link_jobs.RemoveLink),
				gomock.Any(),
				gomock.Eq(link_jobs.RemoveLinkArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
				}),
			).
			MaxTimes(1).
			Return(nil)

		{ // Try to delete it.
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithTimeout(5 * time.Second).
				Expect()

			response.Status(http.StatusOK)
			response.NoContent()
		}

		{ // Try to retrieve the link after it's been deleted.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.deletedAt").NotNull()
			response.JSON().Path("$.plaidLink.status").IsEqual(models.PlaidLinkStatusDeactivated)
			response.JSON().Path("$.plaidLink.deletedAt").NotNull()
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"POST https://sandbox.plaid.com/item/remove": 1,
		}, "must match expected Plaid API calls")
	})

	t.Run("wont remove link twice", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		clock := clock.New()
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, clock)
		token := GivenILogin(t, e, user.Login.Email, password)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		{ // Retrieve the link and do some tests to make sure its a plaid link
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Object().NotEmpty()
			response.JSON().Path("$.deletedAt").IsNull()
			response.JSON().Path("$.plaidLink").Object().NotEmpty()
			response.JSON().Path("$.plaidLink.status").IsEqual(models.PlaidLinkStatusSetup)
			response.JSON().Path("$.plaidLink.deletedAt").IsNull()
			response.JSON().Path("$.linkType").IsEqual(models.PlaidLinkType)
		}

		mock_plaid.MockDeactivateItemTokenSuccess(t)

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(link_jobs.RemoveLink),
				gomock.Any(),
				gomock.Eq(link_jobs.RemoveLinkArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
				}),
			).
			MaxTimes(1).
			Return(nil)

		{ // Try to delete it.
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithTimeout(5 * time.Second).
				Expect()

			response.Status(http.StatusOK)
			response.NoContent()
		}

		{ // If we try to delete it again it should return an error
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithTimeout(5 * time.Second).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Link has already been deleted and cannot be deleted again")
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"POST https://sandbox.plaid.com/item/remove": 1,
		}, "must match expected Plaid API calls")
	})

	t.Run("cant delete someone elses link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var link models.Link

		{ // Create a link under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link = fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to delete the link
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				WithTimeout(5 * time.Second).
				Expect()
			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("failed to retrieve the specified link: record does not exist")
		}
	})

	t.Run("with a valid api key", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		clock := clock.New()
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		institutionName := "U.S. Bank"

		var linkId models.ID[models.Link]
		{ // Create the link with the session token.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": institutionName,
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.institutionName").String().IsEqual(institutionName)
			linkId = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		// Create the API key before we set up the queue expectations so that the
		// key creation cannot be mistaken for one of the queued jobs below.
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(link_jobs.RemoveLink),
				gomock.Any(),
				gomock.Eq(link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    linkId,
				}),
			).
			MaxTimes(1).
			Return(nil)

		{ // Delete the link authenticated with the API key. DELETE
			// /api/links/{linkId} is in the billedKeyOrToken route group.
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", linkId).
				WithBasicAuth(apiKeyId, apiKeySecret).
				WithTimeout(5 * time.Second).
				Expect()

			response.Status(http.StatusOK)
			response.NoContent()
		}

		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{}, "should not have made ANY plaid API calls")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// A well formed but unknown API key must be rejected before we attempt to
		// retrieve or delete anything.
		response := e.DELETE("/api/links/{linkId}").
			WithPath("linkId", "link_bogus").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			WithTimeout(5 * time.Second).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}
