package ui

import (
	"net/http"

	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/config"
	"github.com/sirupsen/logrus"
)

var (
	_ application.Controller = &UIController{}
)

type UIController struct {
	log           *logrus.Entry
	configuration config.Configuration
	filesystem    http.FileSystem
}

// NewUIController creates a UI controller that uses the default embedded filesystem to serve UI files to the client.
// This requires that the UI files have been built and have been placed in the correct directory at the time that the go
// executable is compiled.
func NewUIController(log *logrus.Entry, configuration config.Configuration) *UIController {
	return NewUIControllerCustomFS(log, configuration, http.FS(builtUi))
}

// NewUIControllerCustomFS creates a UI controller that allows you to provide anything that implements the
// http.FileSystem interface in order to serve UI files to the client.
func NewUIControllerCustomFS(
	log *logrus.Entry,
	configuration config.Configuration,
	filesystem http.FileSystem,
) *UIController {
	return &UIController{
		log:           log,
		configuration: configuration,
		filesystem:    NewFileSystem("static", filesystem),
		// fileServer: iris.FileServer(
		// 	NewFileSystem("static", filesystem),
		// 	iris.DirOptions{
		// 		IndexName: "index.html",
		// 		SPA:       true,
		// 	},
		// ),
	}
}
