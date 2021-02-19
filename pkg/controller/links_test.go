package controller_test

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
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

		response := e.POST("/api/links").
			WithHeader("H-Token", token).
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

		response := e.POST("/api/links").
			WithJSON(link).
			Expect()

		response.Status(http.StatusForbidden)
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
				WithHeader("H-Token", tokenA).
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
				WithHeader("H-Token", tokenB).
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
				WithHeader("H-Token", tokenA).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.links").Array().Length().Equal(1)
			response.JSON().Path("$.links[0].linkId").Number().Equal(linkAID)
		}
	})
}
