package controller

import (
	"net/http"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/labstack/echo/v4"
)

func (_ *Controller) listCurrencies(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, locale.GetInstalledCurrencies())
}
