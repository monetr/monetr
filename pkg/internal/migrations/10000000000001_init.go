package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/migrate"
)

var (
	Migrations = migrate.NewMigrations()
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		switch db.Dialect().Name() {
		case dialect.SQLite:
			if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = true;`); err != nil {
				return errors.Wrap(err, "failed to enable foreign keys for SQLite")
			}
		case dialect.PG, dialect.MySQL:
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
