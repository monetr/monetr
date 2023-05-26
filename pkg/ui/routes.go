//go:build !noui

package ui

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
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

func (c *UIController) RegisterRoutes(app *echo.Echo) {
	app.GET("/*", func(ctx echo.Context) error {
		ctx.Response().Header().Set("X-Frame-Options", "DENY")
		ctx.Response().Header().Set("X-Content-Type-Options", "nosniff")
		ctx.Response().Header().Set("Referrer-Policy", "same-origin")
		c.ApplyContentSecurityPolicy(ctx)
		c.ApplyPermissionsPolicy(ctx)

		requestedPath := ctx.Request().URL.Path

		log := c.log.WithFields(logrus.Fields{
			"path":            requestedPath,
			"ext":             path.Ext(requestedPath),
			"resolvedToIndex": false,
		})

		resolvedToIndex := false
		content, err := c.filesystem.Open(requestedPath)
		switch c.fixFilesystemError(err) {
		case fs.ErrNotExist, fs.ErrInvalid:
			content, err = c.filesystem.Open(indexFile)
			resolvedToIndex = true
			log = log.WithField("resolvedToIndex", true)
			ctx.Response().Header().Set("Cache-Control", "no-cache")
			if err != nil {
				panic("could not find index file")
			}
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

		data, err := ioutil.ReadAll(content)
		if err != nil {
			log.WithError(err).Error("failed to read content from embedded file")
			return ctx.NoContent(http.StatusInternalServerError)
		}

		if resolvedToIndex {
			log.WithField("contentType", "text/html").Tracef("%s %s", ctx.Request().Method, ctx.Request().URL.Path)
			return ctx.HTMLBlob(http.StatusOK, data)
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
	})
}
