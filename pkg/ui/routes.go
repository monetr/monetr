//go:build !noui

package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

const (
	EmbeddedUI = true
)

func (c *UIController) RegisterRoutes(app *iris.Application) {
	app.PartyFunc("/", func(p router.Party) {
		p.Any("/api/*", func(ctx iris.Context) {
			ctx.Next()
			ctx.StatusCode(http.StatusNotFound)
			return
		})

		p.Get("/config.json", func(ctx iris.Context) {
			ctx.JSONP(map[string]interface{}{
				"apiUrl": "/api",
			})
		})

		app.Get("/{p:path}", func(ctx iris.Context) {
			if c.configuration.Server.UICacheHours > 0 {
				cacheExpiration := time.Now().
					Add(time.Duration(c.configuration.Server.UICacheHours) * time.Hour).
					Truncate(time.Hour)
				seconds := int(cacheExpiration.Sub(time.Now()).Seconds())
				ctx.Header("Expires", cacheExpiration.Format(http.TimeFormat))
				ctx.Header("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
			}

			ctx.Header("X-Frame-Options", "DENY")
			ctx.Header("X-Content-Type-Options", "nosniff")
			ctx.Header("Referrer-Policy", "same-origin")
			ctx.Header("Permissions-Policy", "accelerometer=(), ambient-light-sensor=(), autoplay=(), battery=(), camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), xr-spatial-tracking=(), clipboard-read=(), clipboard-write=(), gamepad=(), speaker-selection=()")
			c.ContentSecurityPolicyMiddleware(ctx)
			c.fileServer(ctx)
		})
	})
}
