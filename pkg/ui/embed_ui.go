//+build ui

package ui

import "embed"

//go:embed *.js
//go:embed *.html
var builtUi embed.FS
