package application

import (
	"github.com/getsentry/sentry-go"
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetrapp/rest-api/pkg/config"
)

type Controller interface {
	RegisterRoutes(app *iris.Application)
}

func NewApp(configuration config.Configuration, controllers ...Controller) *iris.Application {
	app := iris.New()


	// This will properly display IP addresses as most of the time the API will not be able to see
	// the real IP due to being behind several networking layers. Masquerade only works so much and
	// I'm not sure of a better way.
	app.UseGlobal(func(ctx *context.Context) {
		if forwardedFor := ctx.GetHeader("X-Forwarded-For"); forwardedFor != "" {
			ctx.Request().RemoteAddr = forwardedFor
		}

		// This way we still have a way to correlate users even if they are not authenticated.
		if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetUser(sentry.User{
				IPAddress: ctx.GetHeader("X-Forwarded-For"),
			})
		}

		ctx.Next()
	})

	if configuration.Sentry.Enabled {
		app.Use(sentryiris.New(sentryiris.Options{
			Repanic: false,
		}))
	}

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
			"M-Token",
		},
		ExposedHeaders:     nil,
		MaxAge:             0,
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              configuration.CORS.Debug,
	}))

	for _, controller := range controllers {
		controller.RegisterRoutes(app)
	}

	return app
}
