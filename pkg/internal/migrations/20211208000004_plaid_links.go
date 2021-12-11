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
			ModelTableExpr("plaid_links")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "plaid_link_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:  "item_id",
				Type:  TextOrVarChar(250),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "products",
				Type: Array{
					DataType: TextOrVarChar(250),
				},
			},
			ColumnDefinition{
				Name: "webhook_url",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "institution_id",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "institution_name",
				Type: TextOrVarChar(250),
			},
		)

		createTable.ColumnExpr(`CONSTRAINT pk_plaid_links PRIMARY KEY (plaid_link_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("plaid_links")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
