package migrations_test

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/migrations"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func TestColumnGeneration(t *testing.T) {
	t.Run("basic column", func(t *testing.T) {
		testutils.ForEachDatabaseUnMigrated(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			type TableModel struct{}
			createTable := db.NewCreateTable().
				Model(&TableModel{}).
				ModelTableExpr(`?`, bun.Ident("logins"))

			migrations.CreateColumnDefinition(
				createTable,
				"login_id",
				migrations.BigInteger,
				migrations.AutoIncrement|migrations.NotNull|migrations.PrimaryKey,
			)
			migrations.CreateColumnDefinition(
				createTable,
				"email",
				migrations.TextOrVarChar(250),
				migrations.NotNull,
			)

			query, err := createTable.AppendQuery(db.Formatter(), nil)
			assert.NoError(t, err)
			switch db.Dialect().Name() {
			case dialect.PG:
				assert.Equal(t, `CREATE TABLE "logins" ("login_id" BIGSERIAL PRIMARY KEY NOT NULL, "email" TEXT NOT NULL)`, string(query))
			case dialect.MySQL:
				assert.Equal(t, "CREATE TABLE `logins` (`login_id` BIGINT PRIMARY KEY NOT NULL AUTO_INCREMENT, `email` VARCHAR(250) NOT NULL)", string(query))
			case dialect.SQLite:
				assert.Equal(t, `CREATE TABLE "logins" ("login_id" BIGINTEGER PRIMARY KEY NOT NULL, "email" VARCHAR(250) NOT NULL)`, string(query))
			}

			_, err = createTable.Exec(ctx)
			assert.NoError(t, err, "must be able to create table")
		})
	})

	t.Run("with default", func(t *testing.T) {
		testutils.ForEachDatabaseUnMigrated(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			type TableModel struct{}
			createTable := db.NewCreateTable().
				Model(&TableModel{}).
				ModelTableExpr(`?`, bun.Ident("logins"))

			migrations.CreateColumnDefinition(
				createTable,
				"login_id",
				migrations.BigInteger,
				migrations.AutoIncrement|migrations.NotNull|migrations.PrimaryKey,
			)
			migrations.CreateColumnDefinition(
				createTable,
				"created_at",
				migrations.Timestamp,
				migrations.NotNull,
				bun.Safe("CURRENT_TIMESTAMP"),
			)

			query, err := createTable.AppendQuery(db.Formatter(), nil)
			assert.NoError(t, err)
			switch db.Dialect().Name() {
			case dialect.PG:
				assert.Equal(t, `CREATE TABLE "logins" ("login_id" BIGSERIAL PRIMARY KEY NOT NULL, "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`, string(query))
			case dialect.MySQL:
				assert.Equal(t, "CREATE TABLE `logins` (`login_id` BIGINT PRIMARY KEY NOT NULL AUTO_INCREMENT, `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)", string(query))
			case dialect.SQLite:
				assert.Equal(t, `CREATE TABLE "logins" ("login_id" BIGINTEGER PRIMARY KEY NOT NULL, "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`, string(query))
			}

			_, err = createTable.Exec(ctx)
			assert.NoError(t, err, "must be able to create table")
		})
	})

	t.Run("panics for bad defaults", func(t *testing.T) {
		testutils.ForEachDatabaseUnMigrated(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			type TableModel struct{}
			createTable := db.NewCreateTable().
				Model(&TableModel{}).
				ModelTableExpr(`?`, bun.Ident("logins"))

			migrations.CreateColumnDefinition(
				createTable,
				"login_id",
				migrations.BigInteger,
				migrations.AutoIncrement|migrations.NotNull|migrations.PrimaryKey,
			)
			assert.Panics(t, func() {
				migrations.CreateColumnDefinition(
					createTable,
					"created_at",
					migrations.Timestamp,
					migrations.NotNull,
					"First default value",
					"Second default value", // I cause a panic
				)
			})
		})
	})
}
