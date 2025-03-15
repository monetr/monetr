package migrations

import (
	"embed"
)

//go:embed schema/*.sql
var embededMigrations embed.FS
