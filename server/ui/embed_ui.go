//go:build !noui

package ui

import (
	"embed"
)

//go:embed static/**
var builtUi embed.FS
