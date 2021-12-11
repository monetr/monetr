package migrations

import (
	"context"

	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TableModel struct{}
		createTable := db.NewCreateTable().
			Model(&TableModel{}).
			ModelTableExpr("links")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "link_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:  "account_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "link_type",
				Type:  SmallInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "plaid_link_id",
				Type: BigInteger,
			},
			ColumnDefinition{
				Name: "institution_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name: "custom_institution_name",
				Type: TextOrVarChar(250),
			},
			ColumnDefinition{
				Name:         "created_at",
				Type:         Timestamp,
				Flags:        NotNull,
				DefaultValue: bun.Safe("CURRENT_TIMESTAMP"),
			},
			ColumnDefinition{
				Name:  "created_by_user_id",
				Type:  BigInteger,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:         "updated_at",
				Type:         Timestamp,
				Flags:        NotNull,
				DefaultValue: bun.Safe("CURRENT_TIMESTAMP"),
			},
			ColumnDefinition{
				Name:         "link_status",
				Type:         SmallInteger,
				Flags:        NotNull,
				DefaultValue: models.LinkStatusPending,
			},
			ColumnDefinition{
				Name: "error_code",
				Type: TextOrVarChar(100),
			},
			ColumnDefinition{
				Name: "last_successful_update",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "expiration_date",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "plaid_institution_id",
				Type: TextOrVarChar(250),
			},
		)

		createTable.ColumnExpr(`CONSTRAINT pk_links PRIMARY KEY (link_id, account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_links_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_links_plaid_links_plaid_link_id FOREIGN KEY (plaid_link_id) REFERENCES plaid_links (plaid_link_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_links_users_created_by_user_id FOREIGN KEY (created_by_user_id) REFERENCES users (user_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("links")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
