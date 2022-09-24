package application

import (
	"github.com/getsentry/sentry-go"
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/util"
)

type Controller interface {
	RegisterRoutes(app *iris.Application)
}

func NewApp(configuration config.Configuration, controllers ...Controller) *iris.Application {
	app := iris.New()

	app.Configure(iris.WithoutBanner)

	// This will properly display IP addresses as most of the time the API will not be able to see
	// the real IP due to being behind several networking layers. Masquerade only works so much and
	// I'm not sure of a better way.
	app.UseGlobal(func(ctx *context.Context) {
		ipAddress := util.GetForwardedFor(ctx)
		ctx.Request().RemoteAddr = util.GetForwardedFor(ctx)

		// This way we still have a way to correlate users even if they are not authenticated.
		if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetUser(sentry.User{
				IPAddress: ipAddress,
			})
		}

		ctx.Next()
	})

	app.Use(sentryiris.New(sentryiris.Options{
		Repanic: false,
	}))

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
			"sentry-trace",
			"Authorization",
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
