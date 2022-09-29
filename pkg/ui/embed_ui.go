//go:build !noui

package ui

import (
	"embed"
)

//go:embed static/**
//go:embed static/index.html
var builtUi embed.FS
