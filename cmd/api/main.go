package main

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func main() {
	configuration := config.LoadConfiguration()
	app := iris.New()
	app.UseRouter(cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
		},
		AllowOriginFunc: nil,
		AllowedMethods: []string{
			"HEAD",
			"OPTIONS",
			"GET",
			"POST",
		},
		AllowedHeaders:     nil,
		ExposedHeaders:     nil,
		MaxAge:             0,
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              true,
	}))

	c := controller.NewController(configuration, nil)
	c.RegisterRoutes(app)

	app.Listen(":4000")
}
