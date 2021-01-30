package controller

import (
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/kataras/iris/v12"
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
	return &Controller{
		db:            db,
		configuration: configuration,
	}
}

func (c *Controller) RegisterRoutes(app *iris.Application) {
	app.Get("/api/config", c.configEndpoint)
	app.PartyFunc("/api/authentication", func(p router.Party) {
		p.Post("/login", c.loginEndpoint)
	})
	app.PartyFunc("/api", func(p router.Party) {
		p.Use(c.setupRepositoryMiddleware)

		p.Get("/banks", nil)
		p.Get("/transactions", nil)
		p.Get("/expenses", nil)
		p.Get("/funding", nil)
	})
}
