package migrations

import (
	"context"
	"log/slog"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// MonetrMigrationsManager owns the lifecycle of monetr's schema. It picks up
// the embedded SQL files at construction time and exposes a small surface for
// reading the current version and advancing it.
type MonetrMigrationsManager struct {
	log           *slog.Logger
	db            *pg.DB
	exec          Executor
	files         []migrationFile
	latestVersion int64
}

// NewMigrationsManager wires up a manager against db. The embedded migration
// files are parsed once at package load (see embeddedMigrationFiles); here we
// just create the schema_migrations tracking table if it isn't already there
// and seed rows forward from the legacy gopg_migrations table if we find one.
// It does not actually run any migrations, call Up for that.
func NewMigrationsManager(ctx context.Context, log *slog.Logger, db *pg.DB) (*MonetrMigrationsManager, error) {
	files := embeddedMigrationFiles

	var latest int64
	for _, f := range files {
		if f.Version > latest {
			latest = f.Version
		}
	}

	// ensureSchemaTable grabs the advisory lock, which is session scoped, so it
	// has to run on a single pinned connection rather than the pool. We only need
	// the pin for setup; CurrentVersion and friends are fine on the pool.
	conn := db.Conn()
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.WarnContext(ctx, "failed to release migration setup connection", "err", closeErr)
		}
	}()
	if err := ensureSchemaTable(ctx, log, newPgExecutor(conn)); err != nil {
		return nil, err
	}

	return &MonetrMigrationsManager{
		log:           log,
		db:            db,
		exec:          newPgExecutor(db),
		files:         files,
		latestVersion: latest,
	}, nil
}

// CurrentVersion returns the highest version recorded in schema_migrations,
// or 0 if no migrations have been applied yet.
func (m *MonetrMigrationsManager) CurrentVersion(ctx context.Context) (int64, error) {
	var v int64
	err := m.exec.Get(ctx, &v,
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations",
	)
	return v, errors.Wrap(err, "failed to get current database version")
}

// LatestVersion is the highest version present in the embedded migration files.
// It's computed once at construction time and doesn't query the database.
func (m *MonetrMigrationsManager) LatestVersion() int64 {
	return m.latestVersion
}

// Up applies every pending migration in ascending version order. Pins a single
// backend connection for the run so the advisory lock actually serves its
// purpose, and returns the version before and after.
func (m *MonetrMigrationsManager) Up(ctx context.Context) (oldVersion, newVersion int64, err error) {
	conn := m.db.Conn()
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			m.log.WarnContext(ctx, "failed to release migration connection", "err", closeErr)
		}
	}()

	pinned := newPgExecutor(conn)
	return applyMigrations(ctx, m.log, pinned, embeddedMigrations, m.files)
}

// RunMigrations is the auto-migrate entry point used at server startup and from
// the test harness when it sets up isolated databases. Errors are logged and
// swallowed, the caller doesn't get to react.
func RunMigrations(ctx context.Context, log *slog.Logger, db *pg.DB) {
	m, err := NewMigrationsManager(ctx, log, db)
	if err != nil {
		log.ErrorContext(ctx, "failed to initialize migration manager", "err", err)
		return
	}

	currentVersion, err := m.CurrentVersion(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to get current database version", "err", err)
		return
	}
	log.InfoContext(ctx, "current database version", "version", currentVersion)

	oldVersion, newVersion, err := m.Up(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to run migrations", "err", err)
		return
	}

	if oldVersion == newVersion {
		log.InfoContext(ctx, "no database updates")
	} else {
		log.InfoContext(ctx, "database upgraded", "from", oldVersion, "to", newVersion)
	}
}
