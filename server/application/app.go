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
