//go:build !noui

package ui

import (
	"embed"
)

//go:generate make build-ui
//go:embed static/**
var builtUi embed.FS
