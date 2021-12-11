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
			ModelTableExpr("jobs")

		CreateColumnDefinition(createTable,
			ColumnDefinition{
				Name:         "job_id",
				Type:         TextOrVarChar(250),
				Flags:        NotNull,
				DefaultValue: nil,
			},
			ColumnDefinition{
				Name: "account_id",
				Type: BigInteger,
			},
			ColumnDefinition{
				Name: "arguments",
				Type: JSON,
			},
			ColumnDefinition{
				Name:         "enqueued_at",
				Type:         Timestamp,
				Flags:        NotNull,
				DefaultValue: nil,
			},
			ColumnDefinition{
				Name: "started_at",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "finished_at",
				Type: Timestamp,
			},
			ColumnDefinition{
				Name: "retries",
				Type: BigInteger,
			},
		)

		createTable.ColumnExpr(`CONSTRAINT pk_jobs PRIMARY KEY (job_id)`)
		createTable.ColumnExpr(`CONSTRAINT fk_jobs_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts (account_id)`)

		_, err := createTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", createTable.GetTableName())
	}, func(ctx context.Context, db *bun.DB) error {
		dropTable := db.NewDropTable().TableExpr("jobs")

		_, err := dropTable.Exec(ctx)
		return errors.Wrapf(err, "failed to create %s table", dropTable.GetTableName())
	})
}
