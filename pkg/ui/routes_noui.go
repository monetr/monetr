//go:build noui

package ui

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	EmbeddedUI = false
)

func (c *UIController) RegisterRoutes(app *echo.Echo) {

}
