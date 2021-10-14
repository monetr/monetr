package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		addColumn := db.NewAddColumn().
			TableExpr("users").
			ColumnExpr("stripe_customer_id TEXT")

		_, err := addColumn.Exec(ctx)
		return errors.Wrap(err, "failed to add stripe customer Id column to users table")
	}, func(ctx context.Context, db *bun.DB) error {
		dropColumn := db.NewDropColumn().
			TableExpr("users").
			Column("stripe_customer_id")

		_, err := dropColumn.Exec(ctx)
		return errors.Wrapf(err, "failed to drop stripe_customer_id column from users table")
	})
}
