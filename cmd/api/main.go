package main

import (
	"fmt"

	"github.com/go-pg/pg/v10"
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
		AllowedHeaders: []string{
			"Content-Type",
			"H-Token",
		},
		ExposedHeaders:     nil,
		MaxAge:             0,
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              true,
	}))

	db := pg.Connect(&pg.Options{
		Addr: fmt.Sprintf("%s:%d",
			configuration.PostgreSQL.Address,
			configuration.PostgreSQL.Port,
		),
		User:            configuration.PostgreSQL.Username,
		Password:        configuration.PostgreSQL.Password,
		Database:        configuration.PostgreSQL.Database,
		ApplicationName: "harder - api",
	})

	c := controller.NewController(configuration, db)
	c.RegisterRoutes(app)

	app.Listen(":4000")
}
