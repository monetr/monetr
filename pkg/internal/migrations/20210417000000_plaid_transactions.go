package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		addColumn := db.NewAddColumn().
			TableExpr("transactions").
			ColumnExpr(`pending_plaid_transaction_id TEXT`)

		_, err := addColumn.Exec(ctx)
		return errors.Wrapf(err, "failed to add pending plaid transaction Id column")
	}, func(ctx context.Context, db *bun.DB) error {
		dropColumn := db.NewDropColumn().
			TableExpr("transactions").
			Column("pending_plaid_transaction_id")

		_, err := dropColumn.Exec(ctx)
		return errors.Wrapf(err, "failed to drop pending plaid transaction Id column")
	})
}
