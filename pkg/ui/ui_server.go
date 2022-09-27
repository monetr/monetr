package ui

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/config"
)

var (
	_ application.Controller = &UIController{}
)

type UIController struct {
	configuration config.Configuration
	fileServer    iris.Handler
}

// NewUIController creates a UI controller that uses the default embedded filesystem to serve UI files to the client.
// This requires that the UI files have been built and have been placed in the correct directory at the time that the go
// executable is compiled.
func NewUIController(configuration config.Configuration) *UIController {
	return NewUIControllerCustomFS(configuration, http.FS(builtUi))
}

// NewUIControllerCustomFS creates a UI controller that allows you to provide anything that implements the
// http.FileSystem interface in order to serve UI files to the client.
func NewUIControllerCustomFS(
	configuration config.Configuration,
	filesystem http.FileSystem,
) *UIController {
	return &UIController{
		configuration: configuration,
		fileServer: iris.FileServer(
			NewFileSystem("static", filesystem),
			iris.DirOptions{
				IndexName: "index.html",
				SPA:       true,
			},
		),
	}
}
