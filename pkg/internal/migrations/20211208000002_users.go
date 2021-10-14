package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TableModel struct{}
		createTable := db.NewCreateTable().
			Model(&TableModel{}).
			ModelTableExpr("users")

		db.Dialect().IdentQuote()
		switch db.Dialect().Name() {
		case dialect.PG:
			createTable.ColumnExpr(`user_id BIGSERIAL NOT NULL`)
			createTable.ColumnExpr(`first_name TEXT NOT NULL`)
			createTable.ColumnExpr(`last_name TEXT`)
			createTable.ColumnExpr(`stripe_customer_id TEXT`)
		case dialect.MySQL:
			createTable.ColumnExpr(`user_id BIGINT NOT NULL AUTO_INCREMENT`)
			createTable.ColumnExpr(`first_name VARCHAR(250) NOT NULL`)
			createTable.ColumnExpr(`last_name VARCHAR(250)`)
			createTable.ColumnExpr(`stripe_customer_id VARCHAR(250)`)
		case dialect.SQLite:
			createTable.ColumnExpr(`user_id BIGINTEGER NOT NULL`)
			createTable.ColumnExpr(`first_name TEXT NOT NULL`)
			createTable.ColumnExpr(`last_name TEXT`)
			createTable.ColumnExpr(`stripe_customer_id TEXT`)
		}

		createTable.ColumnExpr(`login_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`account_id BIGINT NOT NULL`)

		createTable.ColumnExpr(`CONSTRAINT pk_users PRIMARY KEY (user_id)`)
		createTable.ColumnExpr(`CONSTRAINT uq_users_login_id_account_id UNIQUE (login_id, account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_users_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_users_logins_login_id FOREIGN KEY (login_id) REFERENCES logins (login_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("users")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
