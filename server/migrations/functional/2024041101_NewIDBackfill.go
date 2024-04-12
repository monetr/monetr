package functional

import (
	"github.com/go-pg/migrations/v8"
)

func init() {
	FunctionalMigrations = append(FunctionalMigrations, &migrations.Migration{
		Version: 2024041101,
		UpTx:    true,
		Up: func(db migrations.DB) error {
			return nil
		},
		DownTx: false,
		Down: func(db migrations.DB) error {
			return nil
		},
	})
}
