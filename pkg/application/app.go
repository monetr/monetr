package application

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func NewApp(configuration config.Configuration, controller *controller.Controller) *iris.Application {
	app := iris.New()
	app.UseRouter(cors.New(cors.Options{
		AllowedOrigins:  configuration.CORS.AllowedOrigins,
		AllowOriginFunc: nil,
		AllowedMethods: []string{
			"HEAD",
			"OPTIONS",
			"GET",
			"POST",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"H-Token",
		},
		ExposedHeaders:     nil,
		MaxAge:             0,
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              configuration.CORS.Debug,
	}))
	controller.RegisterRoutes(app)

	return app
}
