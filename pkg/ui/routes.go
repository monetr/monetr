//go:build !noui

package ui

import (
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
		p.Use(c.ContentSecurityPolicyMiddleware)

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
			ctx.Header("Expires", time.Now().Add(24*time.Hour).Truncate(time.Hour).Format(http.TimeFormat))
			fileHandler(ctx)
		})
	})
}
