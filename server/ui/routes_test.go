//go:build !noui

package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIndexRedirect(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		filesystem := fstest.MapFS{
			"static/index.html": &fstest.MapFile{
				Data: []byte("<html></html>"),
			},
		}
		controller := NewUIControllerCustomFS(
			testutils.GetLog(t),
			config.Configuration{},
			http.FS(filesystem),
		)
		app := echo.New()
		controller.RegisterRoutes(app)

		request := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)

		assert.Equal(t, http.StatusPermanentRedirect, response.Code, "should be a permanent redirect")
		assert.Equal(t, "/", response.Header().Get("Location"), "should redirect to root")
	})

	t.Run("protocol relative path", func(t *testing.T) {
		// This test is here to make sure index.html can't be used as an open
		// redirect to some other site.
		filesystem := fstest.MapFS{
			"static/index.html": &fstest.MapFile{
				Data: []byte("<html></html>"),
			},
		}
		controller := NewUIControllerCustomFS(
			testutils.GetLog(t),
			config.Configuration{},
			http.FS(filesystem),
		)
		app := echo.New()
		controller.RegisterRoutes(app)

		request := httptest.NewRequest(http.MethodGet, "//evil.com/..%2f/index.html", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)

		assert.Equal(t, http.StatusPermanentRedirect, response.Code, "should be a permanent redirect")
		assert.Equal(t, "/", response.Header().Get("Location"), "should not redirect off site")
	})
}
