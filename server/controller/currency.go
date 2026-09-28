package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/currency"
)

func (*Controller) listCurrencies(ctx *echo.Context) error {
	locale := getCurrencyLocale(ctx)
	ctx.Response().Header().Set("Content-Language", locale)
	ctx.Response().Header().Add("Vary", "Accept-Language")
	return ctx.JSON(http.StatusOK, currency.GetCurrencyList(locale))
}

func (c *Controller) getCurrency(ctx *echo.Context) error {
	code := strings.ToUpper(strings.TrimSpace(ctx.Param("currencyCode")))
	locale := getCurrencyLocale(ctx)
	result, err := currency.GetCurrency(locale, code)
	if err != nil {
		return c.notFound(ctx, "Currency is not supported")
	}
	ctx.Response().Header().Set("Content-Language", locale)
	ctx.Response().Header().Add("Vary", "Accept-Language")
	return ctx.JSON(http.StatusOK, result)
}

// getCurrencyLocale returns the locale that currency names and symbols should
// be localized to, the locale query param wins over the Accept-Language header
// if they provided it
func getCurrencyLocale(ctx *echo.Context) string {
	if locale := ctx.QueryParam("locale"); locale != "" {
		tag, err := currency.ParseLocale(locale)
		if err != nil {
			// If they gave us garbage then just fall back to the default
			return currency.MatchLocale("")
		}
		return currency.MatchLocale(tag.String())
	}
	return currency.MatchLocale(ctx.Request().Header.Get("Accept-Language"))
}
