package application

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/controller"
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
			"PUT",
			"DELETE",
		},
		AllowedHeaders: []string{
			"Cookies",
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
