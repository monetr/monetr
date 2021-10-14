package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// This is a placeholder migration to achieve parody with the old SQL migrations. It is not compatible with all
		// databases. So the latest version of this view will be applied in a later migration.
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
