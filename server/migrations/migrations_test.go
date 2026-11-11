package migrations

import (
	"testing"

	"github.com/monetr/monetr/server/internal/testutils/testlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations(t *testing.T) {
	t.Run("clean database", func(t *testing.T) {
		db := newCleanDatabase(t)
		log := testlog.GetLog(t)

		require.NoError(t, RunMigrations(t.Context(), log, db), "migrating a clean database should succeed")

		m, err := NewMigrationsManager(t.Context(), log, db)
		require.NoError(t, err)
		current, err := m.CurrentVersion(t.Context())
		require.NoError(t, err)
		assert.Equal(t, m.LatestVersion(), current, "database should be at the latest version")
	})

	t.Run("failure is returned", func(t *testing.T) {
		db := connectTestDatabase(testPgOptions(t)...)
		require.NoError(t, db.Close())
		log := testlog.GetLog(t)

		err := RunMigrations(t.Context(), log, db)
		assert.Error(t, err, "migrations against an unusable database must return an error")
	})
}
