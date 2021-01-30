package main

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func main() {
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

	configuration := config.Configuration{
		JWTSecret:     "abc123",
		UIDomainName:  "localhost:3000",
		APIDomainName: "localhost:4000",
		PostgreSQL:    config.PostgreSQL{},
		SMTP:          config.SMTPClient{},
		ReCAPTCHA:     config.ReCAPTCHA{},
		AllowSignUp:   true,
	}

	c := controller.NewController(configuration, nil)
	c.RegisterRoutes(app)

	app.Listen(":4000")
}
