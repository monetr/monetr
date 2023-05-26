package controller_test

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
)

func TestPostLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		createdAndUpdatedBogus := time.Now().Add(-1 * time.Hour)
		link := models.Link{
			LinkId:                math.MaxInt64,        // Set it to something so we can verify its different in the result.
			LinkType:              models.PlaidLinkType, // This should be changed to manual in the response.
			InstitutionName:       "U.S. Bank",
			CustomInstitutionName: "US Bank",
			Description:           myownsanity.StringP("My personal link"),
			PlaidInstitutionId:    myownsanity.StringP("ins_123"),
			CreatedAt:             createdAndUpdatedBogus, // Set these to something to make sure it gets overwritten.
			UpdatedAt:             createdAndUpdatedBogus,
		}

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(link).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").Number().NotEqual(link.LinkId)
		response.JSON().Path("$.linkId").Number().Gt(0)
		response.JSON().Path("$.linkType").Number().Equal(models.ManualLinkType)
		response.JSON().Path("$.institutionName").String().NotEmpty()
		response.JSON().Path("$.description").String().Equal("My personal link")
		response.JSON().Path("$.plaidInstitutionId").Null()
	})

	t.Run("missing name", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		link := models.Link{
			InstitutionName:       "",
			CustomInstitutionName: "",
		}

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(link).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("link must have an institution name")
	})

	t.Run("malformed json", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithBytes([]byte("I'm not json")).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").Equal("invalid JSON body")
	})

	t.Run("unauthenticated", func(t *testing.T) {
		e := NewTestApplication(t)
		link := models.Link{
			LinkId:                math.MaxInt64,        // Set it to something so we can verify its different in the result.
			LinkType:              models.PlaidLinkType, // This should be changed to manual in the response.
			InstitutionName:       "U.S. Bank",
			CustomInstitutionName: "US Bank",
			CreatedAt:             time.Now().Add(-1 * time.Hour), // Set these to something to make sure it gets overwritten.
			UpdatedAt:             time.Now().Add(1 * time.Hour),
		}

		response := e.POST("/api/links").
			WithJSON(link).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().Equal("unauthorized")
	})
}

func TestGetLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)

		tokenA, tokenB := GivenIHaveToken(t, e), GivenIHaveToken(t, e)

		link := models.Link{
			LinkId:                math.MaxInt64,        // Set it to something so we can verify its different in the result.
			LinkType:              models.PlaidLinkType, // This should be changed to manual in the response.
			InstitutionName:       "U.S. Bank",
			CustomInstitutionName: "US Bank",
			CreatedAt:             time.Now().Add(-1 * time.Hour), // Set these to something to make sure it gets overwritten.
			UpdatedAt:             time.Now().Add(1 * time.Hour),
		}

		// We want to create a link with tokenA. This link should not be visible later when we request the link for
		// tokenB. This will help verify that we do not expose data from someone else's login.
		var linkAID uint64
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenA).
				WithJSON(link).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Number().NotEqual(link.LinkId)
			response.JSON().Path("$.linkId").Number().Gt(0)
			response.JSON().Path("$.linkType").Number().Equal(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.plaidInstitutionId").Null()
			linkAID = uint64(response.JSON().Path("$.linkId").Number().Raw())
		}

		// Create a link for tokenB too. This way we can do a GET request for both tokens to test each scenario.
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenB).
				WithJSON(link).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Number().NotEqual(link.LinkId)
			response.JSON().Path("$.linkId").Number().Gt(0)
			response.JSON().Path("$.linkType").Number().Equal(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
		}

		// Now we want to test GET with token A.
		{
			response := e.GET("/api/links").
				WithCookie(TestCookieName, tokenA).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$").Array().Length().Equal(1)
			response.JSON().Path("$[0].linkId").Number().Equal(linkAID)
		}

		// Now we want to test GET with token B.
		{
			response := e.GET("/api/links").
				WithCookie(TestCookieName, tokenB).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$").Array().Length().Equal(1)
			// Make sure that we do not receive token A's link.
			response.JSON().Path("$[0].linkId").Number().NotEqual(linkAID)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		e := NewTestApplication(t)
		response := e.GET("/api/links").
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().Equal("unauthorized")
	})

	t.Run("precise", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		institutionName := "U.S. Bank"

		link := models.Link{
			LinkType:              models.ManualLinkType,
			InstitutionName:       institutionName,
			CustomInstitutionName: institutionName,
		}

		{ // Create the link.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(link).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Number().Gt(0)
			response.JSON().Path("$.institutionName").String().Equal(institutionName)
			response.JSON().Path("$.customInstitutionName").String().Equal(institutionName)

			link.LinkId = uint64(response.JSON().Path("$.linkId").Number().Raw())
		}

		{ // Retrieve the link and make sure the linkId matches.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Equal(link.LinkId)
		}
	})

	t.Run("precise not found", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		{ // Try to retrieve a link that does not exist for this user.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", math.MaxInt64).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").Equal("failed to retrieve link: record does not exist")
		}
	})

	t.Run("precise bad Id", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		{ // Try to retrieve a link that does not exist for this user.
			response := e.GET("/api/links/0").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").Equal("must specify a link Id to retrieve")
		}
	})

	t.Run("plaid link", func(t *testing.T) {
		e := NewTestApplication(t)

		var token string
		var linkId uint64
		{
			user, password := fixtures.GivenIHaveABasicAccount(t)
			link := fixtures.GivenIHaveAPlaidLink(t, user)
			linkId = link.LinkId
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		response := e.GET("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.plaidInstitutionId").String().NotEmpty()
	})
}

func TestPutLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		institutionName := "U.S. Bank"

		link := models.Link{
			LinkType:              models.ManualLinkType,
			InstitutionName:       institutionName,
			CustomInstitutionName: institutionName,
		}

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(link).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").Number().Gt(0)
		response.JSON().Path("$.institutionName").String().Equal(institutionName)
		response.JSON().Path("$.customInstitutionName").String().Equal(institutionName)

		linkId := uint64(response.JSON().Path("$.linkId").Number().Raw())

		link.LinkId = linkId
		link.CustomInstitutionName = "New Name"
		link.InstitutionName = "New Name"

		updated := e.PUT("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithCookie(TestCookieName, token).
			WithJSON(link).
			Expect()

		updated.Status(http.StatusOK)
		updated.JSON().Path("$.linkId").Number().Equal(linkId)
		// Make sure the institution name has not changed. This cannot be changed once a link is created.
		updated.JSON().Path("$.institutionName").String().Equal(institutionName)
		// But make sure that the custom institution name has changed.
		updated.JSON().Path("$.customInstitutionName").String().Equal("New Name")
	})

	t.Run("unauthenticated", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		institutionName := "U.S. Bank"

		link := models.Link{
			LinkType:              models.ManualLinkType,
			InstitutionName:       institutionName,
			CustomInstitutionName: institutionName,
		}

		response := e.POST("/api/links").
			WithCookie(TestCookieName, token).
			WithJSON(link).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").Number().Gt(0)
		response.JSON().Path("$.institutionName").String().Equal(institutionName)
		response.JSON().Path("$.customInstitutionName").String().Equal(institutionName)

		linkId := uint64(response.JSON().Path("$.linkId").Number().Raw())

		link.LinkId = linkId
		link.CustomInstitutionName = "New Name"
		link.InstitutionName = "New Name"

		// Try to perform an update without a token.
		updated := e.PUT("/api/links/{linkId}").
			WithPath("linkId", linkId).
			WithJSON(link).
			Expect()

		updated.Status(http.StatusUnauthorized)
		updated.JSON().Path("$.error").String().Equal("unauthorized")
	})

	t.Run("cannot update someone elses", func(t *testing.T) {
		e := NewTestApplication(t)

		tokenA, tokenB := GivenIHaveToken(t, e), GivenIHaveToken(t, e)

		link := models.Link{
			InstitutionName:       "U.S. Bank",
			CustomInstitutionName: "US Bank",
		}

		// We want to create a link with tokenA. This link should not be visible later when we request the link for
		// tokenB. This will help verify that we do not expose data from someone else's login.
		var linkAID, linkBID uint64
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenA).
				WithJSON(link).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Number().NotEqual(link.LinkId)
			response.JSON().Path("$.linkId").Number().Gt(0)
			response.JSON().Path("$.linkType").Number().Equal(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			linkAID = uint64(response.JSON().Path("$.linkId").Number().Raw())
		}

		// Create a link for tokenB too. This way we can do a GET request for both tokens to test each scenario.
		{
			response := e.POST("/api/links").
				WithCookie(TestCookieName, tokenB).
				WithJSON(link).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Number().NotEqual(link.LinkId)
			response.JSON().Path("$.linkId").Number().Gt(0)
			response.JSON().Path("$.linkType").Number().Equal(models.ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			linkBID = uint64(response.JSON().Path("$.linkId").Number().Raw())
		}

		// Now using token A, try to update token B's link.
		{
			link := models.Link{
				LinkId:                linkBID,
				CustomInstitutionName: "I have changed",
			}
			response := e.PUT("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, tokenA).
				WithJSON(link).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().Equal("failed to retrieve existing link for update: record does not exist")
		}

		// Now do the same thing with token B for token A's link.
		{
			link := models.Link{
				LinkId:                linkAID,
				CustomInstitutionName: "I have changed",
			}
			response := e.PUT("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, tokenB).
				WithJSON(link).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().Equal("failed to retrieve existing link for update: record does not exist")
		}
	})
}

func TestDeleteLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		institutionName := "U.S. Bank"

		link := models.Link{
			LinkType:              models.ManualLinkType,
			InstitutionName:       institutionName,
			CustomInstitutionName: institutionName,
		}

		{ // Create the link.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(link).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").Number().Gt(0)
			response.JSON().Path("$.institutionName").String().Equal(institutionName)
			response.JSON().Path("$.customInstitutionName").String().Equal(institutionName)

			link.LinkId = uint64(response.JSON().Path("$.linkId").Number().Raw())
		}

		{ // Try to retrieve the link before it's been deleted.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Object().NotEmpty()
		}

		{ // Try to delete it.
			response := e.DELETE("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.NoContent()
		}

		{ // Try to retrieve the link after it's been deleted.
			response := e.GET("/api/links/{linkId}").
				WithPath("linkId", link.LinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").Equal("failed to retrieve link: record does not exist")
		}
	})
}
