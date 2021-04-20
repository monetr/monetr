//+build ui

package ui

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/config"
	"net/http"
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
		p.Any("/api", func(ctx *context.Context) {
			ctx.Next()
			ctx.StatusCode(http.StatusNotFound)
			return
		})

		routes := p.HandleDir("/", http.FS(builtUi), iris.DirOptions{
			IndexName: "index.html",
		})
		fmt.Sprint(routes)

		p.Get("/config.json", func(ctx *context.Context) {
			ctx.JSONP(map[string]interface{}{
				"apiUrl": "/api",
			})
		})

		p.Get("/*", func(ctx *context.Context) {
		})
	})
}
