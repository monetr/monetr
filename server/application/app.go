package application

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/sentryecho"
)

type Controller interface {
	RegisterRoutes(app *echo.Echo)
}

// ConfigureServer applies our HTTP server timeouts to mitigate slowloris-style
// attacks. In echo v4 these lived on app.Server, but echo v5 no longer exposes
// the http.Server, so we set them through StartConfig.BeforeServeFunc instead.
// WriteTimeout must exceed the 30s Plaid long-poll timeout in
// controller.getWaitForPlaid; 45s keeps 15s of headroom.
// TODO migrate that long-poll to a websocket so this write deadline can be
// tightened further.
func ConfigureServer(server *http.Server) {
	server.ReadHeaderTimeout = 5 * time.Second
	server.ReadTimeout = 30 * time.Second
	server.WriteTimeout = 45 * time.Second
	server.IdleTimeout = 120 * time.Second
}

func NewApp(configuration config.Configuration, controllers ...Controller) *echo.Echo {
	app := echo.New()

	app.Use(sentryecho.New(sentryecho.Options{
		Repanic:         false,
		WaitForDelivery: false,
		Timeout:         30 * time.Second,
	}))
	// Right now uploads are soft limited to 5MB anyway, this gives us some head
	// room and this should also be defined at the reverse proxy layer as well to
	// prevent spam. echo v5 takes the limit as raw bytes instead of a string.
	app.Use(middleware.BodyLimit(6 * 1024 * 1024)) // 6MB

	// CORS is opt in. We only register the middleware when origins are explicitly
	// configured, that way monetr never advertises cross origin access by default
	// and the browser keeps enforcing same origin. echo v4 used to default an
	// unset AllowOrigins to "*" which allowed any origin, but that is not what we
	// want and echo v5 panics on "*" + AllowCredentials anyway.
	if len(configuration.CORS.AllowedOrigins) > 0 {
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
	}

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			if configuration.Server.GetIsSecureProtocol() {
				ctx.Response().Header().Set("Strict-Transport-Security", "max-age=31536000")
			}
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

			// In echo v5 the response byte count lives on the *echo.Response
			// behind the http.ResponseWriter, so we have to unwrap it to read
			// the size for the Content-Length header.
			if response, err := echo.UnwrapResponse(ctx.Response()); err == nil {
				ctx.Response().Header().Add(
					echo.HeaderContentLength,
					strconv.FormatInt(response.Size, 10),
				)
			}
			return nil
		}
	})

	for _, controller := range controllers {
		controller.RegisterRoutes(app)
	}

	return app
}
