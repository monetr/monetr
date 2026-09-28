package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
)

func TestListCurrencies(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The locale/currency endpoint lives on the billedKeyOrToken route group so
		// it accepts an API key. It only requires authentication (plus an active
		// subscription, which passes because billing is disabled in the test
		// config), so a valid key should reach the handler and return the installed
		// currency list.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET(`/api/locale/currency`).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()
		response.Status(http.StatusOK)
		// The handler returns the list of installed ISO currency codes as a JSON
		// array, assert the shape is present.
		response.JSON().Array().NotEmpty()
	})

	t.Run("locale query param takes priority over accept language", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/locale/currency`).
			WithQuery("locale", "ja").
			WithHeader("Accept-Language", "de").
			WithCookie(TestCookieName, token).
			Expect()
		response.Status(http.StatusOK)
		response.Header("Content-Language").IsEqual("ja")
		jpy := response.JSON().Array().Filter(func(_ int, value *httpexpect.Value) bool {
			return value.Object().Value("code").String().Raw() == "JPY"
		})
		jpy.Length().IsEqual(1)
		jpy.Value(0).Object().Value("name").IsEqual("日本円")
		jpy.Value(0).Object().Value("fractionalDigits").IsEqual(0)
	})

	t.Run("accept language is used without a locale query param", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/locale/currency`).
			WithHeader("Accept-Language", "de").
			WithCookie(TestCookieName, token).
			Expect()
		response.Status(http.StatusOK)
		response.Header("Content-Language").IsEqual("de")
	})

	t.Run("posix style locale query param", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/locale/currency`).
			WithQuery("locale", "de_CH").
			WithCookie(TestCookieName, token).
			Expect()
		response.Status(http.StatusOK)
		response.Header("Content-Language").IsEqual("de-CH")
	})

	t.Run("invalid locale query param falls back to english", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/locale/currency`).
			WithQuery("locale", "!!").
			WithHeader("Accept-Language", "de").
			WithCookie(TestCookieName, token).
			Expect()
		response.Status(http.StatusOK)
		response.Header("Content-Language").IsEqual("en")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs.
		_, e := NewTestApplication(t)

		response := e.GET(`/api/locale/currency`).
			WithBasicAuth("key_"+gofakeit.UUID(), gofakeit.UUID()).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}

func TestGetCurrency(t *testing.T) {
	t.Run("locale query param takes priority over accept language", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/locale/currency/jpy`).
			WithQuery("locale", "ja").
			WithHeader("Accept-Language", "de").
			WithCookie(TestCookieName, token).
			Expect()
		response.Status(http.StatusOK)
		response.Header("Content-Language").IsEqual("ja")
		response.JSON().Object().IsEqual(map[string]any{
			"code":             "JPY",
			"name":             "日本円",
			"symbol":           "￥",
			"fractionalDigits": 0,
		})
	})

	t.Run("unsupported currency", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/locale/currency/DEM`).
			WithQuery("locale", "ja").
			WithCookie(TestCookieName, token).
			Expect()
		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Currency is not supported")
	})
}
