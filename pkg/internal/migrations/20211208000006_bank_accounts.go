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
			ModelTableExpr("bank_accounts")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "bank_account_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:  "account_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "link_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "plaid_account_id",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:  "available_balance",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "current_balance",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "mask",
				Type: TextOrVarChar(50),
			},
			ColumnDefinition{
				Name:  "name",
				Type:  TextOrVarChar(250),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "plaid_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "plaid_official_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "account_type",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "account_sub_type",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:         "last_updated",
				Type:         Timestamp,
				Flags:        NotNull,
				DefaultValue: bun.Safe("CURRENT_TIMESTAMP"), // TODO If this isnt in UTC this will cause issues.
			},
		)

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
