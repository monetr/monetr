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

	t.Run("index.html redirects to root", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		assert.Equal(t, http.StatusPermanentRedirect, response.Code)
		assert.Equal(t, "/", response.Header().Get("Location"))
	})

	t.Run("protocol relative path does not redirect off site", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "//evil.com/..%2f/index.html", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		assert.Equal(t, http.StatusPermanentRedirect, response.Code)
		assert.Equal(t, "/", response.Header().Get("Location"))
	})
}
