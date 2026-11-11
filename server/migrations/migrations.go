package migrations

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// MigrationsManager owns the lifecycle of monetr's schema for a single
// database engine. Each engine gets its own implementation since the way we
// lock, track versions, and run migration bodies is going to differ between
// them.
type MigrationsManager interface {
	CurrentVersion(ctx context.Context) (int64, error)
	LatestVersion() int64
	Up(ctx context.Context) (oldVersion, newVersion int64, err error)
}

// NewMigrationsManager wires up a manager against db, picking the
// implementation based on the dialect db was opened with. It does not
// actually run any migrations, call Up for that.
func NewMigrationsManager(
	ctx context.Context,
	log *slog.Logger,
	db *bun.DB,
) (MigrationsManager, error) {
	switch db.Dialect().Name() {
	case dialect.PG:
		return newPGMigrationsManager(ctx, log, db)
	default:
		return nil, errors.Errorf(
			"migrations are not supported for database dialect %q",
			db.Dialect().Name(),
		)
	}
}

// RunMigrations is the auto-migrate entry point used at server startup and from
// the test harness when it sets up isolated databases. Any failure is returned
// so the caller can refuse to start against a database whose schema is not
// where monetr expects it to be.
func RunMigrations(ctx context.Context, log *slog.Logger, db *bun.DB) error {
	m, err := NewMigrationsManager(ctx, log, db)
	if err != nil {
		return errors.Wrap(err, "failed to initialize migration manager")
	}

	currentVersion, err := m.CurrentVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get current database version")
	}
	log.InfoContext(ctx, "current database version", "version", currentVersion)

	oldVersion, newVersion, err := m.Up(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	if oldVersion == newVersion {
		log.InfoContext(ctx, "no database updates")
	} else {
		log.InfoContext(ctx, "database upgraded", "from", oldVersion, "to", newVersion)
	}

	return nil
}
