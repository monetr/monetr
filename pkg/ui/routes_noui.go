//go:build noui

package ui

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

const (
	EmbeddedUI = false
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

		app.Get("/*", func(ctx iris.Context) {
			ctx.StatusCode(http.StatusNotFound)
			return
		})
	})
}
