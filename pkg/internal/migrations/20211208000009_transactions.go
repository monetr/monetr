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
			ModelTableExpr("transactions")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "transaction_id",
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
				Name: "plaid_transaction_id",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:  "amount",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "spending_id",
				Type: BigInteger,
			},
			ColumnDefinition{
				Name: "spending_amount",
				Type: BigInteger,
			},
			ColumnDefinition{
				Name: "categories",
				Type: Array{
					DataType: TextOrVarChar(500),
				},
			},
			ColumnDefinition{
				Name: "original_categories",
				Type: Array{
					DataType: TextOrVarChar(500),
				},
			},
			ColumnDefinition{
				Name: "name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:  "original_name",
				Type:  TextOrVarChar(250),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "merchant_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "original_merchant_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:  "is_pending",
				Type:  Boolean,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:         "created_at",
				Type:         Timestamp,
				Flags:        NotNull, // TODO Default?
				DefaultValue: nil,
			},
			ColumnDefinition{
				Name: "pending_plaid_transaction_id",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:  "date",
				Type:  Timestamp,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "authorized_date",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "column_name",
				Type: TextOrVarChar(250),
			},
		)

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
