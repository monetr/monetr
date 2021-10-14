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
			ModelTableExpr("funding_schedules")

		db.Dialect().IdentQuote()
		switch db.Dialect().Name() {
		case dialect.PG:
			createTable.ColumnExpr(`funding_schedule_id BIGSERIAL NOT NULL`)
		case dialect.MySQL:
			createTable.ColumnExpr(`funding_schedule_id BIGINT NOT NULL AUTO_INCREMENT`)
		case dialect.SQLite:
			createTable.ColumnExpr(`funding_schedule_id BIGINTEGER NOT NULL`)
		}

		createTable.ColumnExpr(`account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`bank_account_id BIGINT NOT NULL`)
		createTable.ColumnExpr(`name VARCHAR(250) NOT NULL`)
		createTable.ColumnExpr(`description TEXT`)
		createTable.ColumnExpr(`rule TEXT NOT NULL`)
		createTable.ColumnExpr(`last_occurrence TIMESTAMP`)
		createTable.ColumnExpr(`next_occurrence TIMESTAMP NOT NULL`)

		createTable.ColumnExpr(`CONSTRAINT pk_funding_schedules PRIMARY KEY (funding_schedule_id, account_id, bank_account_id)`)
		createTable.ColumnExpr(`CONSTRAINT uq_funding_schedules_bank_account_id_name UNIQUE (bank_account_id, name)`)
		createTable.ColumnExpr(`CONSTRAINT fk_funding_schedules_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_funding_schedules_bank_accounts_bank_account_id_account_id FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts (bank_account_id, account_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("funding_schedules")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
