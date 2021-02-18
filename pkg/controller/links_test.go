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
