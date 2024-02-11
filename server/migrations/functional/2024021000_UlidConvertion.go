package functional

import "github.com/go-pg/migrations/v8"

func init() {
	FunctionalMigrations = append(FunctionalMigrations, &migrations.Migration{
		Version: 2024011000,
		UpTx:    true,
		Up: func(db migrations.DB) error {

			return nil
		},
		DownTx: true,
		Down: func(db migrations.DB) error {
			return nil
		},
	})
}
