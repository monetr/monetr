//go:build !noui

package ui

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/monetr/monetr/server/internal/sentryecho"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	EmbeddedUI = true
	indexFile  = "/index.html"
)

func (c *UIController) fixFilesystemError(err error) error {
	unwrappedError := errors.Unwrap(err)
	switch unwrappedError {
	case fs.ErrNotExist, fs.ErrInvalid:
		return unwrappedError
	default:
		return err
	}
}

func (c *UIController) registerIndexRenderer(app *echo.Echo) {
	index, err := c.filesystem.Open(indexFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read index.html file: %+v", err))
	}

	indexData, err := io.ReadAll(index)
	if err != nil {
		panic(fmt.Sprintf("failed to read contents of index.html file: %+v", err))
	}

	indexTemplate := template.New(indexFile)
	indexTemplate, err = indexTemplate.Parse(string(indexData))
	if err != nil {
		panic(fmt.Sprintf("failed to parse index.html template: %+v", err))
	}

	app.Renderer = &indexRenderer{
		index: indexTemplate,
	}
}

func (c *UIController) RegisterRoutes(app *echo.Echo) {
	c.registerIndexRenderer(app)

	app.GET("/*", func(ctx echo.Context) error {
		defer func(ctx echo.Context) {
			if err := recover(); err != nil {
				hub := sentryecho.GetHubFromContext(ctx)
				hub.Recover(err)
				c.log.Errorf("panic for request: %+v\n%s", err, string(debug.Stack()))
				_ = ctx.String(http.StatusInternalServerError, "Something went very wrong!")
			}
		}(ctx)
		requestedPath := path.Clean(ctx.Request().URL.Path)

		// Even though we are using an embedded filesystem for the UI, we still want
		// to make sure we do not use relative paths.
		if !path.IsAbs(requestedPath) {
			return ctx.NoContent(http.StatusNotFound)
		}

		// If they request `/index.html` simply redirect them to `/`.
		if requestedPath == indexFile {
			url := ctx.Request().URL.String()
			url = strings.TrimSuffix(url, requestedPath)
			return ctx.Redirect(http.StatusPermanentRedirect, url)
		}

		log := c.log.WithFields(logrus.Fields{
			"path":            requestedPath,
			"ext":             path.Ext(requestedPath),
			"resolvedToIndex": false,
		})

		content, err := c.filesystem.Open(requestedPath)
		switch c.fixFilesystemError(err) {
		case fs.ErrNotExist, fs.ErrInvalid:
			log = log.WithField("resolvedToIndex", true)

			// Only apply these headers and content security permissions to the
			// index.html return result.
			ctx.Response().Header().Set("Cache-Control", "no-cache")
			ctx.Response().Header().Set("X-Frame-Options", "DENY")
			ctx.Response().Header().Set("X-Content-Type-Options", "nosniff")
			ctx.Response().Header().Set("Referrer-Policy", "same-origin")
			c.ApplyContentSecurityPolicy(ctx)
			c.ApplyPermissionsPolicy(ctx)

			log.WithField("contentType", "text/html").Tracef("%s %s", ctx.Request().Method, ctx.Request().URL.Path)
			return ctx.Render(http.StatusOK, indexFile, indexParams{
				SentryDSN: c.configuration.Sentry.ExternalDSN,
			})
		case nil:
			if c.configuration.Server.UICacheHours > 0 {
				cacheExpiration := time.Now().
					Add(time.Duration(c.configuration.Server.UICacheHours) * time.Hour).
					Truncate(time.Hour)
				seconds := int(time.Until(cacheExpiration).Seconds())
				ctx.Response().Header().Set("Expires", cacheExpiration.Format(http.TimeFormat))
				ctx.Response().Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
			}
		default:
			log.WithError(err).Error("failed to read the embedded file specified")
			return ctx.NoContent(http.StatusInternalServerError)
		}

		data, err := io.ReadAll(content)
		if err != nil {
			log.WithError(err).Error("failed to read content from embedded file")
			return ctx.NoContent(http.StatusInternalServerError)
		}

		contentType := "text/plain"
		switch strings.ToLower(path.Ext(requestedPath)) {
		case ".json", "json":
			contentType = "application/json"
		case ".js", "js":
			contentType = "text/javascript"
		case ".css", "css":
			contentType = "text/css"
		default:
			contentType = http.DetectContentType(data)
		}
		log.WithField("contentType", contentType).Tracef("%s %s", ctx.Request().Method, ctx.Request().URL.Path)
		return ctx.Blob(http.StatusOK, contentType, data)
	}, middleware.Gzip())
}
