package ui

import (
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/config"
)

var (
	_ application.Controller = &UIController{}
)

type UIController struct {
	configuration config.Configuration
}

func NewUIController() *UIController {
	return &UIController{}
}

func (c *UIController) RegisterRoutes(app *iris.Application) {
	app.PartyFunc("/", func(p router.Party) {
		p.Any("/api/*", func(ctx *context.Context) {
			ctx.Next()
			ctx.StatusCode(http.StatusNotFound)
			return
		})

		p.Get("/config.json", func(ctx *context.Context) {
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
