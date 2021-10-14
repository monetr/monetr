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
			ModelTableExpr("spending")

		db.Dialect().IdentQuote()
		switch db.Dialect().Name() {
		case dialect.PG:
			createTable.ColumnExpr(`spending_id BIGSERIAL NOT NULL`)
		case dialect.MySQL:
			createTable.ColumnExpr(`spending_id BIGINT NOT NULL AUTO_INCREMENT`)
		case dialect.SQLite:
			createTable.ColumnExpr(`spending_id BIGINTEGER NOT NULL`)
		}

		createTable.ColumnExpr(`account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`bank_account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`funding_schedule_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`spending_type SMALLINT NOT NULL`)
		createTable.ColumnExpr(`name VARCHAR(250) NOT NULL`)
		createTable.ColumnExpr(`description TEXT`)
		createTable.ColumnExpr(`target_amount BIGINT NOT NULL`)
		createTable.ColumnExpr(`current_amount BIGINT NOT NULL`)
		createTable.ColumnExpr(`used_amount BIGINT NOT NULL`)
		createTable.ColumnExpr(`recurrence_rule TEXT`)
		createTable.ColumnExpr(`last_recurrence TIMESTAMP`)
		createTable.ColumnExpr(`next_recurrence TIMESTAMP NOT NULL`)
		createTable.ColumnExpr(`next_contribution_amount BIGINT NOT NULL`)
		createTable.ColumnExpr(`is_behind BOOLEAN NOT NULL`)
		createTable.ColumnExpr(`is_paused BOOLEAN NOT NULL`)
		createTable.ColumnExpr(`date_created TIMESTAMP NOT NULL`)

		createTable.ColumnExpr(`CONSTRAINT pk_spending PRIMARY KEY (spending_id, account_id, bank_account_id)`)
		createTable.ColumnExpr(`CONSTRAINT uq_spending_bank_account_id_spending_type_name UNIQUE (bank_account_id, spending_type, name)`)
		createTable.ColumnExpr(`CONSTRAINT fk_spending_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_spending_bank_accounts_bank_account_id_account_id FOREIGN KEY (funding_schedule_id, account_id, bank_account_id) REFERENCES funding_schedules (funding_schedule_id, account_id, bank_account_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("spending")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
