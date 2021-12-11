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
			ModelTableExpr("accounts")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "account_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:         "timezone",
				Type:         TextOrVarChar(100),
				Flags:        NotNull,
				DefaultValue: "UTC",
			},
			ColumnDefinition{
				Name: "stripe_customer_id",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "stripe_subscription_id",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "subscription_active_until",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "stripe_webhook_latest_timestamp",
				Type: Timestamp,
			},
		)

		createTable.ColumnExpr(`CONSTRAINT pk_accounts PRIMARY KEY (account_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("accounts")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
