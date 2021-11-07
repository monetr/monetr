package functional

import (
	"fmt"
	"github.com/go-pg/migrations/v8"
)

func init() {
	FunctionalMigrations = append(FunctionalMigrations, &migrations.Migration{
		Version: 2021050999,
		UpTx:    false,
		Up: func(db migrations.DB) error {
			fmt.Println("TEST MIGRATION UP")
			return nil
		},
		DownTx: false,
		Down: func(db migrations.DB) error {
			fmt.Println("TEST MIGRATION DOWN")
			return nil
		},
	})
}
