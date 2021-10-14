package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			{ // First name
				// SQLite does not easily let us change a column constraint after the column has been created. This
				// particular migration is relatively simple. But to work around how SQLite works, we will create the
				// column with a default value.
				// In MySQL or Postgres this would be really bad; But I don't anticipate a user ever really having so
				// much data in their SQLite database that this is going to be worse than the alternative.
				addColumn := tx.NewAddColumn().
					TableExpr("logins")
				switch db.Dialect().Name() {
				case dialect.SQLite:
					addColumn.ColumnExpr(`first_name TEXT NOT NULL DEFAULT 'N/A'`)
				case dialect.PG, dialect.MySQL:
					addColumn.ColumnExpr(`first_name TEXT`)
				}

				if _, err := addColumn.Exec(ctx); err != nil {
					return errors.Wrapf(err, "failed to add first_name columns to login")
				}
			}
			{ // Last name
				addColumn := tx.NewAddColumn().
					TableExpr("logins").
					ColumnExpr(`last_name TEXT`)
				if _, err := addColumn.Exec(ctx); err != nil {
					return errors.Wrapf(err, "failed to add first_name columns to login")
				}
			}

			var firstName, lastName bun.Ident
			switch db.Dialect().Name() {
			case dialect.MySQL:
				firstName = "logins.first_name"
				lastName = "logins.last_name"
			default:
				firstName = `first_name`
				lastName = `last_name`
			}

			_, err := tx.NewUpdate().
				Table(`logins`, `users`).
				Set(`? = users.first_name`, firstName).
				Set(`? = users.last_name`, lastName).
				Where(`logins.login_id = users.login_id`).
				Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to move names from users table to login table")
			}

			// Then once we are done, if we are on SQLite then we don't have any more work to do. But if we are on MySQL
			// or PostgreSQL then we need to update the column constraint.
			switch db.Dialect().Name() {
			case dialect.MySQL:
				_, err = tx.ExecContext(ctx, `ALTER TABLE logins MODIFY first_name TEXT NOT NULL;`)
			case dialect.PG:
				_, err = tx.ExecContext(ctx, `ALTER TABLE logins ALTER COLUMN first_name SET NOT NULL;`)
			case dialect.SQLite:
				// Nothing to be done.
			}

			return errors.Wrap(err, "failed to set first_name to be not null")
		})
	}, func(ctx context.Context, db *bun.DB) error {
		{
			dropColumn := db.NewDropColumn().
				TableExpr("logins").
				Column("first_name")

			if _, err := dropColumn.Exec(ctx); err != nil {
				return errors.Wrapf(err, "failed to drop first_name columns from login")
			}
		}
		{
			dropColumn := db.NewDropColumn().
				TableExpr("logins").
				Column("last_name")

			if _, err := dropColumn.Exec(ctx); err != nil {
				return errors.Wrapf(err, "failed to drop last_name columns from login")
			}
		}

		return nil
	})
}
