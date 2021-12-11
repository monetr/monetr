package migrations_test

import (
	"context"
	"fmt"
	"testing"

	monetrMigrations "github.com/monetr/monetr/pkg/internal/migrations"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

func TestMigrate(t *testing.T) {
	testutils.ForEachDatabaseUnMigrated(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
		migrator := migrate.NewMigrator(db, monetrMigrations.Migrations)

		assert.NoError(t, migrator.Init(ctx), "must init db")

		status, err := migrator.MigrationsWithStatus(ctx)
		assert.NoError(t, err, "failed")
		assert.NotNil(t, status)

		group, err := migrator.Migrate(ctx)
		assert.NoError(t, err, "failed to migrate")
		fmt.Println("Migrate: ", group)

		group, err = migrator.Rollback(ctx)
		assert.NoError(t, err, "failed to migrate")
		fmt.Println("Rollback:", group)
	})
}

func TestTableName(t *testing.T) {
	testutils.ForEachDatabaseUnMigrated(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
		type SampleModel struct {
			bun.BaseModel `bun:"table:users"`
		}

		query := db.NewSelect().Model(&SampleModel{})
		assert.Equal(t, "users", query.GetTableName())
	})
}
