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
			ModelTableExpr("users")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "user_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:  "login_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:         "account_id",
				Type:         BigInteger,
				Flags:        NotNull,
				DefaultValue: nil,
			},
			ColumnDefinition{
				Name:  "first_name",
				Type:  TextOrVarChar(250),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "last_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "stripe_customer_id",
				Type: TextOrVarChar(250),
			},
		)

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
