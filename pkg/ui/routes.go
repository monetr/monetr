//go:build !noui

package ui

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	EmbeddedUI = true
	indexFile  = "index.html"
)

func (c *UIController) RegisterRoutes(app *echo.Echo) {
	app.GET("/*", func(ctx echo.Context) error {
		ctx.Response().Header().Set("X-Frame-Options", "DENY")
		ctx.Response().Header().Set("X-Content-Type-Options", "nosniff")
		ctx.Response().Header().Set("Referrer-Policy", "same-origin")
		ctx.Response().Header().Set("Permissions-Policy", "accelerometer=(), ambient-light-sensor=(), autoplay=(), battery=(), camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), xr-spatial-tracking=(), clipboard-read=(), clipboard-write=(), gamepad=(), speaker-selection=()")
		c.ContentSecurityPolicyMiddleware(ctx)

		requestedPath := ctx.Request().URL.Path
		resolvedToIndex := false
		content, err := c.filesystem.Open(requestedPath)
		switch err {
		case fs.ErrNotExist:
			content, err = c.filesystem.Open(indexFile)
			resolvedToIndex = true
			ctx.Response().Header().Set("Cache-Control", "no-cache")
			if err != nil {
				panic("could not find index file")
			}
		case nil:
			if c.configuration.Server.UICacheHours > 0 {
				cacheExpiration := time.Now().
					Add(time.Duration(c.configuration.Server.UICacheHours) * time.Hour).
					Truncate(time.Hour)
				seconds := int(time.Until(cacheExpiration).Seconds())
				ctx.Response().Header().Set("Expires", cacheExpiration.Format(http.TimeFormat))
				ctx.Response().Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
			}
		default:
			return ctx.NoContent(http.StatusInternalServerError)
		}

		data, err := ioutil.ReadAll(content)
		if err != nil {
			return ctx.NoContent(http.StatusInternalServerError)
		}

		if resolvedToIndex {
			return ctx.HTMLBlob(http.StatusOK, data)
		}

		contentType := http.DetectContentType(data)
		return ctx.Blob(http.StatusOK, contentType, data)
	})
}
