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
			ModelTableExpr("spending")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "spending_id",
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
				Name:  "funding_schedule_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "spending_type",
				Type:  SmallInteger,
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
				Name:  "target_amount",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "current_amount",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "used_amount",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "recurrence_rule",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "last_recurrence",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name:  "next_recurrence",
				Type:  Timestamp,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "next_contribution_amount",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "is_behind",
				Type:  Boolean,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "is_paused",
				Type:  Boolean,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "date_created",
				Type:  Timestamp,
				Flags: NotNull,
			},
		)

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
