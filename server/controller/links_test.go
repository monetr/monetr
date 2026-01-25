package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/testutils"
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
		response.JSON().Path("$.problems.institutionName").String().IsEqual("Institution name must be between 1 and 300 characters")
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
		response.JSON().Path("$.problems.institutionName").String().IsEqual("Institution name is required")
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
		response.JSON().Path("$.problems.description").String().IsEqual("Description must be between 1 and 300 characters")
	})

	t.Run("malformed json", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I'm not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("Failed to parse post request")
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
}

func TestPutLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

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

		linkId := models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())

		updated := e.PUT("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": "New Name",
				"description":     "Add description",
			}).
			Expect()

		updated.Status(http.StatusOK)
		updated.JSON().Path("$.linkId").IsEqual(linkId)
		// Make sure the institution name has not changed. This cannot be changed once a link is created.
		updated.JSON().Path("$.institutionName").String().IsEqual(institutionName)
	})

	t.Run("not modified", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

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

		linkId := models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())

		updated := e.PUT("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"institutionName": institutionName,
			}).
			Expect()

		updated.Status(http.StatusNotModified)
		updated.NoContent()
	})

	t.Run("unauthenticated", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

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

		linkId := models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())

		// Try to perform an update without a token.
		updated := e.PUT("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithJSON(map[string]any{
				"institutionName": "New Name",
			}).
			Expect()

		updated.Status(http.StatusUnauthorized)
		updated.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("cannot update someone elses", func(t *testing.T) {
		_, e := NewTestApplication(t)

		tokenA, tokenB := GivenIHaveToken(t, e), GivenIHaveToken(t, e)

		// We want to create a link with tokenA. This link should not be visible later when we request the link for
		// tokenB. This will help verify that we do not expose data from someone else's login.
		var linkAID, linkBID models.ID[models.Link]
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenA).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkType").IsEqual(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			linkAID = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		// Create a link for tokenB too. This way we can do a GET request for both tokens to test each scenario.
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenB).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkType").IsEqual(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			linkBID = models.ID[models.Link](response.JSON().Path("$.linkId").String().Raw())
		}

		// Now using token A, try to update token B's link.
		{
			response := e.PUT("/api/links/{linkId}").
				WithPath("linkId", linkBID).
				WithCookie(TestCookieName, tokenA).
				WithJSON(map[string]any{
					"institutionName": "New Name",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing link for update: record does not exist")
		}

		// Now do the same thing with token B for token A's link.
		{
			response := e.PUT("/api/links/{linkId}").
				WithPath("linkId", linkAID).
				WithCookie(TestCookieName, tokenB).
				WithJSON(map[string]any{
					"institutionName": "New Name",
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve existing link for update: record does not exist")
		}
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

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.RemoveLink),
				testutils.NewGenericMatcher(func(args background.RemoveLinkArguments) bool {
					a := assert.EqualValues(t, linkId, args.LinkId, "Link ID should match")
					b := assert.EqualValues(t, user.AccountId, args.AccountId, "Account ID should match")
					return a && b
				}),
			).
			Times(1).
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

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.RemoveLink),
				testutils.NewGenericMatcher(func(args background.RemoveLinkArguments) bool {
					a := assert.EqualValues(t, link.LinkId, args.LinkId, "Link ID should match")
					b := assert.EqualValues(t, user.AccountId, args.AccountId, "Account ID should match")
					return a && b
				}),
			).
			Times(1).
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

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.RemoveLink),
				testutils.NewGenericMatcher(func(args background.RemoveLinkArguments) bool {
					a := assert.EqualValues(t, link.LinkId, args.LinkId, "Link ID should match")
					b := assert.EqualValues(t, user.AccountId, args.AccountId, "Account ID should match")
					return a && b
				}),
			).
			Times(1).
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
}
