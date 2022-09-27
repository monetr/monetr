//go:build noui

package ui

import (
	"embed"
)

// When we are building with `noui` then do nothing here.
var builtUi embed.FS
