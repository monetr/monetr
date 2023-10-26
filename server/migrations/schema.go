package migrations

import (
	"embed"
)

//go:embed schema/*.sql
var things embed.FS
