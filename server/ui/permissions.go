package ui

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

var (
	permissionHeader string = ""
)

func init() {
	permissions := map[string][]string{
		"accelerometer":             {},
		"autoplay":                  {},
		"camera":                    {},
		"clipboard-read":            {},
		"clipboard-write":           {},
		"cross-origin-isolated":     {},
		"display-capture":           {},
		"encrypted-media":           {},
		"fullscreen":                {},
		"gamepad":                   {},
		"geolocation":               {},
		"gyroscope":                 {},
		"keyboard-map":              {},
		"magnetometer":              {},
		"microphone":                {},
		"midi":                      {},
		"payment":                   {},
		"picture-in-picture":        {},
		"publickey-credentials-get": {},
		"screen-wake-lock":          {},
		"sync-xhr":                  {},
		"usb":                       {},
		"xr-spatial-tracking":       {},
	}

	items := make([]string, 0, len(permissions))
	for permission, props := range permissions {
		items = append(items, fmt.Sprintf("%s=(%s)", permission, strings.Join(props, " ")))
	}

	if len(items) > 0 {
		permissionHeader = strings.Join(items, ", ")
	}
}

func (c *UIController) ApplyPermissionsPolicy(ctx echo.Context) {
	if permissionHeader != "" {
		ctx.Response().Header().Set("Permissions-Policy", permissionHeader)
	}
}
