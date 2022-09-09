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

		fileHandler := iris.FileServer(
			NewFileSystem("static", http.FS(builtUi)),
			iris.DirOptions{
				IndexName: "index.html",
				SPA:       true,
			},
		)

		app.Get("/*", func(ctx iris.Context) {
			cacheExpiration := time.Now().Add(24*time.Hour).Truncate(time.Hour)
			seconds := int(cacheExpiration.Sub(time.Now()).Seconds())
			ctx.Header("Expires", cacheExpiration.Format(http.TimeFormat))
			ctx.Header("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
			fileHandler(ctx)
			c.ContentSecurityPolicyMiddleware(ctx)
		})
	})
}
