package migrations

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TableModel struct{}
		createTable := db.NewCreateTable().
			Model(&TableModel{}).
			ModelTableExpr("funding_schedules")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "funding_schedule_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:  "account_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "bank_account_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "name",
				Type:  TextOrVarChar(250),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "description",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:  "rule",
				Type:  TextOrVarChar(250),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "last_occurrence",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name:  "next_occurrence",
				Type:  Timestamp,
				Flags: NotNull,
			},
		)

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
