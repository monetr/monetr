package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

const (
	TokenName = "H-Token"
)

type Controller struct {
	db            *pg.DB
	configuration config.Configuration
	captcha       *recaptcha.ReCAPTCHA
}

func NewController(configuration config.Configuration, db *pg.DB) *Controller {
	var captcha recaptcha.ReCAPTCHA
	var err error
	if configuration.ReCAPTCHA.Enabled {
		captcha, err = recaptcha.NewReCAPTCHA(
			configuration.ReCAPTCHA.PrivateKey,
			recaptcha.V2,
			30*time.Second,
		)
		if err != nil {
			panic(err)
		}
	}

	return &Controller{
		captcha:       &captcha,
		configuration: configuration,
		db:            db,
	}
}

func (c *Controller) RegisterRoutes(app *iris.Application) {
	app.OnAnyErrorCode(func(ctx *context.Context) {
		if err := ctx.GetErr(); err != nil {
			ctx.JSON(map[string]interface{}{
				"error": err.Error(),
			})
		}
	})
	app.OnErrorCode(http.StatusNotFound, func(ctx *context.Context) {
		ctx.JSON(map[string]interface{}{
			"path":  ctx.Path(),
			"error": "the requested path does not exist",
		})
	})

	app.Get("/health", func(ctx *context.Context) {
		dbHealthy := c.db.Ping(ctx.Request().Context()) == nil

		ctx.JSON(map[string]interface{}{
			"dbHealthy":  dbHealthy,
			"apiHealthy": true,
		})
	})

	// For the following endpoints we want to have a repository available to us.
	app.PartyFunc("/api", func(p router.Party) {
		p.Get("/config", c.configEndpoint)

		p.Use(c.setupRepositoryMiddleware)

		p.PartyFunc("/authentication", func(p router.Party) {
			p.Post("/login", c.loginEndpoint)
			p.Post("/register", c.registerEndpoint)
		})

		p.Use(c.authenticationMiddleware)

		p.PartyFunc("/users", func(p router.Party) {
			p.Get("/me", func(ctx *context.Context) {
				fmt.Println(ctx)
			})
		})

	})
}
