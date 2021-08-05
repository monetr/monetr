package controller_test

import (
	"fmt"
	"github.com/monetr/rest-api/pkg/models"
	"math"
	"net/http"
	"testing"
	"time"
)

func TestPostLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		link := models.Link{
			LinkId:                math.MaxInt64,        // Set it to something so we can verify its different in the result.
			LinkType:              models.PlaidLinkType, // This should be changed to manual in the response.
			InstitutionName:       "U.S. Bank",
			CustomInstitutionName: "US Bank",
			CreatedAt:             time.Now().Add(-1 * time.Hour), // Set these to something to make sure it gets overwritten.
			UpdatedAt:             time.Now().Add(1 * time.Hour),
		}

		response := e.POST("/links").
			WithHeader("M-Token", token).
			WithJSON(link).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkId").Number().NotEqual(link.LinkId)
		response.JSON().Path("$.linkId").Number().Gt(0)
		response.JSON().Path("$.linkType").Number().Equal(models.ManualLinkType)
		response.JSON().Path("$.institutionName").String().NotEmpty()
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

		response := e.POST("/links").
			WithJSON(link).
			Expect()

		response.Status(http.StatusForbidden)
		response.JSON().Path("$.error").String().Equal("token must be provided")
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
			response := e.POST("/links").
				WithHeader("M-Token", tokenA).
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
			response := e.POST("/links").
				WithHeader("M-Token", tokenB).
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
			response := e.GET("/links").
				WithHeader("M-Token", tokenA).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$").Array().Length().Equal(1)
			response.JSON().Path("$[0].linkId").Number().Equal(linkAID)
		}

		// Now we want to test GET with token B.
		{
			response := e.GET("/links").
				WithHeader("M-Token", tokenB).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$").Array().Length().Equal(1)
			// Make sure that we do not receive token A's link.
			response.JSON().Path("$[0].linkId").Number().NotEqual(linkAID)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		e := NewTestApplication(t)
		response := e.GET("/links").
			Expect()

		response.Status(http.StatusForbidden)
		response.JSON().Path("$.error").String().Equal("token must be provided")
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

		response := e.POST("/links").
			WithHeader("M-Token", token).
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

		updated := e.PUT(fmt.Sprintf("/links/%d", linkId)).
			WithHeader("M-Token", token).
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

		response := e.POST("/links").
			WithHeader("M-Token", token).
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
		updated := e.PUT(fmt.Sprintf("/links/%d", linkId)).
			WithJSON(link).
			Expect()

		updated.Status(http.StatusForbidden)
		updated.JSON().Path("$.error").String().Equal("token must be provided")
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
			response := e.POST("/links").
				WithHeader("M-Token", tokenA).
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
			response := e.POST("/links").
				WithHeader("M-Token", tokenB).
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
			response := e.PUT(fmt.Sprintf("/links/%d", link.LinkId)).
				WithHeader("M-Token", tokenA).
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
			response := e.PUT(fmt.Sprintf("/links/%d", link.LinkId)).
				WithHeader("M-Token", tokenB).
				WithJSON(link).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().Equal("failed to retrieve existing link for update: record does not exist")
		}
	})
}
