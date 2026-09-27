package migrations

import (
	"embed"
)

//go:embed schema/pg/*.sql
var embeddedMigrations embed.FS
