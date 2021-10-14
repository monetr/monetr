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
			ModelTableExpr("bank_accounts")

		db.Dialect().IdentQuote()
		switch db.Dialect().Name() {
		case dialect.PG:
			createTable.ColumnExpr(`bank_account_id BIGSERIAL NOT NULL`)
		case dialect.MySQL:
			createTable.ColumnExpr(`bank_account_id BIGINT NOT NULL AUTO_INCREMENT`)
		case dialect.SQLite:
			createTable.ColumnExpr(`bank_account_id BIGINTEGER NOT NULL`)
		}

		createTable.ColumnExpr(`account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`link_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`plaid_bank_account_id TEXT`)
		createTable.ColumnExpr(`available_balance BIGINT NOT NULL`)
		createTable.ColumnExpr(`current_balance BIGINT NOT NULL`)
		createTable.ColumnExpr(`mask VARCHAR(10)`)
		createTable.ColumnExpr(`name TEXT NOT NULL`)
		createTable.ColumnExpr(`plaid_name TEXT`)
		createTable.ColumnExpr(`plaid_official_name TEXT`)
		createTable.ColumnExpr(`account_type TEXT`)
		createTable.ColumnExpr(`account_sub_type TEXT`)

		createTable.ColumnExpr(`CONSTRAINT pk_bank_accounts PRIMARY KEY (bank_account_id, account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_bank_accounts_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_bank_accounts_links_link_id_account_id FOREIGN KEY (link_id, account_id) REFERENCES links (link_id, account_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("bank_accounts")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
