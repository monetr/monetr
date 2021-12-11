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
			ModelTableExpr("logins")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:  "login_id",
				Type:  BigInteger,
				Flags: NotNull | AutoIncrement,
			},
			ColumnDefinition{
				Name:  "email",
				Type:  TextOrVarChar(100),
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:  "password_hash",
				Type:  TextOrVarChar(500), // Is 500 enough?
				Flags: NotNull,
			},
			ColumnDefinition{
				Name: "phone_number",
				Type: TextOrVarChar(50),
			},
			ColumnDefinition{
				Name:  "is_enabled",
				Type:  Boolean,
				Flags: NotNull,
			},
			ColumnDefinition{
				Name:         "is_email_verified",
				Type:         Boolean,
				Flags:        NotNull,
				DefaultValue: false,
			},
			ColumnDefinition{
				Name:         "is_phone_verified",
				Type:         Boolean,
				Flags:        NotNull,
				DefaultValue: false,
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
				Name: "email_verified_at",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "password_reset_at",
				Type: Timestamp,
			},
		)

		createTable.ColumnExpr(`CONSTRAINT pk_logins PRIMARY KEY (login_id)`)
		createTable.ColumnExpr(`CONSTRAINT uq_logins_email UNIQUE (email)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("logins")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
