package migrations

import (
	"embed"
)

//go:embed schema/*.sql
var embeddedMigrations embed.FS
