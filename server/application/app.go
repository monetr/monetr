package application

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/sentryecho"
)

type Controller interface {
	RegisterRoutes(app *echo.Echo)
}

func NewApp(configuration config.Configuration, controllers ...Controller) *echo.Echo {
	app := echo.New()
	app.HideBanner = true
	app.HidePort = true
	app.Use(sentryecho.New(sentryecho.Options{
		Repanic:         false,
		WaitForDelivery: false,
		Timeout:         30 * time.Second,
	}))
	// Right now uploads are soft limited to 5MB anyway, this gives us some head
	// room and this should also be defined at the reverse proxy layer as well to
	// prevent spam.
	app.Use(middleware.BodyLimit("6MB"))

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: configuration.CORS.AllowedOrigins,
		AllowMethods: []string{
			http.MethodDelete,
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPost,
			http.MethodPut,
		},
		AllowHeaders: []string{
			"Cookies",
			"Content-Type",
			"M-Token",
			"sentry-trace",
			"Authorization",
		},
		ExposeHeaders:    nil,
		MaxAge:           0,
		AllowCredentials: true,
	}))

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// TODO Add HSTS here if the external URL has an HTTPS protocol.
			ctx.Response().Header().Set("X-Frame-Options", "DENY")
			ctx.Response().Header().Set("X-Content-Type-Options", "nosniff")
			ctx.Response().Header().Set("Referrer-Policy", "same-origin")
			// Note: I would love to add Cross-Origin-Opener-Policy: same-origin
			// however I believe this will break legitimate Plaid OAuth flows with
			// real banks since they are not necessarily using the callback pattern?
			// I'm not 100% sure. Ditto Cross-Origin-Embedder-Policy: require-corp,
			// pretty sure this just straight up breaks plaid.
			// TODO Add `Cross-Origin-Resource-Policy: same-origin` but this breaks
			// emails since they load the logo from the server that sent the email!

			if err := next(ctx); err != nil {
				return err
			}

			ctx.Response().Header().Add(
				echo.HeaderContentLength,
				strconv.FormatInt(ctx.Response().Size, 10),
			)
			return nil
		}
	})

	for _, controller := range controllers {
		controller.RegisterRoutes(app)
	}

	return app
}
