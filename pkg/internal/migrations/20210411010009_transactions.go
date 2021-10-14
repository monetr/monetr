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
			ModelTableExpr("transactions")

		db.Dialect().IdentQuote()
		switch db.Dialect().Name() {
		case dialect.PG:
			createTable.ColumnExpr(`transaction_id BIGSERIAL NOT NULL`)
			createTable.ColumnExpr(`plaid_transaction_id TEXT`)
			createTable.ColumnExpr(`categories TEXT[]`)
			createTable.ColumnExpr(`original_categories TEXT[]`)
		case dialect.MySQL:
			createTable.ColumnExpr(`transaction_id BIGINT NOT NULL AUTO_INCREMENT`)
			createTable.ColumnExpr(`plaid_transaction_id VARCHAR(250)`)
			createTable.ColumnExpr(`categories TEXT`)
			createTable.ColumnExpr(`original_categories TEXT`)
		case dialect.SQLite:
			createTable.ColumnExpr(`transaction_id BIGINTEGER NOT NULL`)
			createTable.ColumnExpr(`plaid_transaction_id TEXT`)
			createTable.ColumnExpr(`categories BLOB`)
			createTable.ColumnExpr(`original_categories BLOB`)
		}

		createTable.ColumnExpr(`account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`bank_account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`amount BIGINT NOT NULL`)
		createTable.ColumnExpr(`spending_id BIGINT`)
		createTable.ColumnExpr(`spending_amount BIGINT`)
		createTable.ColumnExpr(`date DATE NOT NULL`)
		createTable.ColumnExpr(`authorized_date DATE`)
		createTable.ColumnExpr(`name TEXT`)
		createTable.ColumnExpr(`original_name TEXT NOT NULL`)
		createTable.ColumnExpr(`merchant_name TEXT`)
		createTable.ColumnExpr(`original_merchant_name TEXT`)
		createTable.ColumnExpr(`is_pending BOOLEAN NOT NULL`)
		createTable.ColumnExpr(`created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`)

		createTable.ColumnExpr(`CONSTRAINT pk_transactions PRIMARY KEY (transaction_id, account_id, bank_account_id)`)
		createTable.ColumnExpr(`CONSTRAINT uq_transactions_bank_account_id_plaid_transaction_id UNIQUE (bank_account_id, plaid_transaction_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_transactions_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_transactions_bank_accounts_bank_account_id_account_id FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts (bank_account_id, account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_transactions_spending_spending_id_account_id_bank_account_id FOREIGN KEY (spending_id, account_id, bank_account_id) REFERENCES spending (spending_id, account_id, bank_account_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("transactions")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
