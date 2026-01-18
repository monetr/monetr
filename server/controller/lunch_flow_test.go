package controller_test

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/models"
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
		response.JSON().Path("$.linkId").String().IsASCII()
		response.JSON().Path("$.linkType").IsEqual(models.LunchFlowLinkType)
		response.JSON().Path("$.institutionName").String().NotEmpty()
		response.JSON().Path("$.description").IsNull()
		response.JSON().Path("$.lunchFlowLinkId").String().IsASCII()
		response.JSON().Path("$.lunchFlowLink.lunchFlowLinkId").String().IsASCII()
		response.JSON().Path("$.lunchFlowLink.status").String().IsEqual("active")
		response.JSON().Path("$.lunchFlowLink").Object().Keys().NotContainsAll("secretId", "secret")
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
