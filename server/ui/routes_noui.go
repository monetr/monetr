//go:build noui

package ui

import (
	"github.com/labstack/echo/v5"
)

const (
	EmbeddedUI = false
)

func (c *UIController) RegisterRoutes(app *echo.Echo) {

}
