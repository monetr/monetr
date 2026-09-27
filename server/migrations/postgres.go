package migrations

import (
	"context"
	"io/fs"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// advisoryLockKey is the session-scoped Postgres advisory lock we hold while
// applying migrations. Stops two monetr processes (rolling restart, parallel
// test setups) from racing each other on the same database. The value itself is
// arbitrary, it just has to be unique within monetr.
const advisoryLockKey int64 = 20180110

var (
	_ MigrationsManager = &PGMigrationsManager{}
)

// pgMigrations is the embedded schema/pg directory, rooted so the migration
// files sit at the top level of the filesystem.
var pgMigrations = mustSubMigrations(embeddedMigrations, "schema/pg")

// pgMigrationFiles is the parsed, sorted list of up-migrations baked into the
// binary. We resolve it once at package load via mustDiscoverMigrations so a
// malformed embedded filename blows up the moment the package is loaded (at
// build/test time, or worst case the very first instruction of a server boot)
// rather than lying in wait until someone actually tries to migrate.
var pgMigrationFiles = mustDiscoverMigrations(pgMigrations)

// PGMigrationsManager owns the lifecycle of monetr's schema on PostgreSQL. It
// picks up the embedded SQL files at construction time and exposes a small
// surface for reading the current version and advancing it.
type PGMigrationsManager struct {
	log           *slog.Logger
	db            *bun.DB
	exec          Executor
	files         []migrationFile
	latestVersion int64
}

// newPGMigrationsManager wires up a manager against db. The embedded migration
// files are parsed once at package load, see pgMigrationFiles. Here we
// just create the schema_migrations tracking table if it isn't already there
// and seed rows forward from the legacy gopg_migrations table if we find one.
func newPGMigrationsManager(
	ctx context.Context,
	log *slog.Logger,
	db *bun.DB,
) (*PGMigrationsManager, error) {
	files := pgMigrationFiles

	var latest int64
	for _, f := range files {
		if f.Version > latest {
			latest = f.Version
		}
	}

	m := &PGMigrationsManager{
		log:           log,
		db:            db,
		exec:          newBunExecutor(db),
		files:         files,
		latestVersion: latest,
	}

	// ensureSchemaTable grabs the advisory lock, which is session scoped, so it
	// has to run on a single pinned connection rather than the pool. We only need
	// the pin for setup; CurrentVersion and friends are fine on the pool.
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to acquire migration setup connection")
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			m.log.WarnContext(ctx, "failed to release migration setup connection", "err", closeErr)
		}
	}()
	if err := m.ensureSchemaTable(ctx, newBunExecutor(conn)); err != nil {
		return nil, err
	}

	return m, nil
}

// CurrentVersion returns the highest version recorded in schema_migrations,
// or 0 if no migrations have been applied yet.
func (m *PGMigrationsManager) CurrentVersion(ctx context.Context) (int64, error) {
	var v int64
	err := m.exec.Get(ctx, &v,
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations",
	)
	return v, errors.Wrap(err, "failed to get current database version")
}

// LatestVersion is the highest version present in the embedded migration files.
// It's computed once at construction time and doesn't query the database.
func (m *PGMigrationsManager) LatestVersion() int64 {
	return m.latestVersion
}

// Up applies every pending migration in ascending version order. Pins a single
// backend connection for the run so the advisory lock actually serves its
// purpose, and returns the version before and after.
func (m *PGMigrationsManager) Up(ctx context.Context) (oldVersion, newVersion int64, err error) {
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to acquire migration connection")
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			m.log.WarnContext(ctx, "failed to release migration connection", "err", closeErr)
		}
	}()

	pinned := newBunExecutor(conn)
	return m.applyMigrations(ctx, pinned, pgMigrations, m.files)
}

// applyMigrations runs every file in files whose version is greater than the
// current MAX(version) in schema_migrations, ascending. The advisory lock is
// held for the whole run so concurrent migrators (rolling restarts, parallel
// test DBs) just take turns.
//
// exec MUST be pinned to a single backend connection or the advisory lock won't
// gate anything past the lock acquisition itself.
func (m *PGMigrationsManager) applyMigrations(
	ctx context.Context,
	exec Executor,
	fsys fs.FS,
	files []migrationFile,
) (oldVersion, newVersion int64, err error) {
	if err := exec.Exec(ctx, "SELECT pg_advisory_lock(?)", advisoryLockKey); err != nil {
		return 0, 0, errors.Wrap(err, "failed to acquire migration advisory lock")
	}
	defer func() {
		if unlockErr := exec.Exec(ctx, "SELECT pg_advisory_unlock(?)", advisoryLockKey); unlockErr != nil {
			m.log.WarnContext(ctx, "failed to release migration advisory lock", "err", unlockErr)
		}
	}()

	applied, err := loadAppliedVersions(ctx, exec)
	if err != nil {
		return 0, 0, err
	}
	for v := range applied {
		if v > oldVersion {
			oldVersion = v
		}
	}
	newVersion = oldVersion

	// Warn (but don't error) on files whose version is at or below the current
	// max but were never applied. go-pg silently skipped these, but a maintainer
	// probably wants to know they're sitting in the tree.
	for _, f := range files {
		if f.Version <= oldVersion {
			if _, ok := applied[f.Version]; !ok {
				m.log.WarnContext(ctx,
					"migration file present but skipped because its version is at or below the currently applied max",
					"version", f.Version,
					"name", f.Name,
					"current", oldVersion,
				)
			}
			continue
		}

		m.log.InfoContext(ctx,
			"applying migration",
			"version", f.Version,
			"name", f.Name,
			"tx", f.Transactional,
		)

		content, err := fs.ReadFile(fsys, f.Filename)
		if err != nil {
			return oldVersion, newVersion, errors.Wrapf(err, "failed to read migration %q", f.Filename)
		}
		sqlText := string(content)

		if f.Transactional {
			err = exec.RunInTransaction(ctx, func(tx Executor) error {
				if err := tx.Exec(ctx, sqlText); err != nil {
					return errors.Wrap(err, "migration body failed")
				}
				if err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", f.Version); err != nil {
					return errors.Wrap(err, "failed to record migration version")
				}
				return nil
			})
		} else {
			if err = exec.Exec(ctx, sqlText); err != nil {
				err = errors.Wrap(err, "migration body failed")
			} else if err = exec.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", f.Version); err != nil {
				err = errors.Wrap(err, "failed to record non-transactional migration version after a successful body")
			}
		}
		if err != nil {
			return oldVersion, newVersion, errors.Wrapf(err, "failed to apply migration %d (%s)", f.Version, f.Name)
		}

		newVersion = f.Version
	}

	return oldVersion, newVersion, nil
}

// loadAppliedVersions returns the set of versions currently recorded in
// schema_migrations. The set is small (low hundreds), so reading the whole
// thing and stuffing it into a map is cheaper than re-querying per file.
func loadAppliedVersions(ctx context.Context, exec Executor) (map[int64]struct{}, error) {
	var versions []int64
	if err := exec.Query(ctx, &versions, "SELECT version FROM schema_migrations"); err != nil {
		return nil, errors.Wrap(err, "failed to read applied migration versions")
	}
	out := make(map[int64]struct{}, len(versions))
	for _, v := range versions {
		out[v] = struct{}{}
	}
	return out, nil
}

const schemaCreateSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT      NOT NULL PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// gopgExistsSQL is true if the legacy gopg_migrations table is in the current
// schema. We scope to current_schema() so a gopg_migrations sitting in some
// other schema (e.g. a shared public in a multi-tenant cluster) doesn't trigger
// a seed copy into the wrong place.
const gopgExistsSQL = `SELECT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = current_schema()
      AND table_name = 'gopg_migrations'
)`

// seedFromGopgSQL copies any rows we don't already have from gopg_migrations
// into schema_migrations.
//   - version > 0 is defensive; go-pg's init doesn't insert a sentinel, but a
//     hand-rolled DB might.
//   - version <> 2021050999 skips the deleted Go-side test migration, which has
//     no .sql file in the new tree.
//     TODO Remove this filter (and the rest of the gopg seed path) once we're
//     confident no production DB still has gopg_migrations.
//   - ON CONFLICT makes it idempotent across restarts and tolerant of
//     partially-seeded state.
const seedFromGopgSQL = `INSERT INTO schema_migrations (version, applied_at)
SELECT version, COALESCE(created_at, NOW())
FROM gopg_migrations
WHERE version > 0
  AND version <> 2021050999
ON CONFLICT (version) DO NOTHING`

// ensureSchemaTable creates schema_migrations if it isn't already there and, if
// the legacy gopg_migrations table is present, copies forward any versions we
// don't have yet. Safe to call repeatedly.
//
// Like applyMigrations it holds the migration advisory lock for the whole run.
// Postgres doesn't actually serialize CREATE TABLE IF NOT EXISTS, so two
// processes booting at once (rolling restart, parallel test setups) can race it
// and one side errors out, and we'd also rather not have both of them seeding
// from gopg_migrations at the same time. The lock makes them take turns
// instead. exec MUST therefore be pinned to a single backend connection or the
// lock gates nothing past acquisition.
func (m *PGMigrationsManager) ensureSchemaTable(ctx context.Context, exec Executor) error {
	if err := exec.Exec(ctx, "SELECT pg_advisory_lock(?)", advisoryLockKey); err != nil {
		return errors.Wrap(err, "failed to acquire migration advisory lock")
	}
	defer func() {
		if unlockErr := exec.Exec(ctx, "SELECT pg_advisory_unlock(?)", advisoryLockKey); unlockErr != nil {
			m.log.WarnContext(ctx, "failed to release migration advisory lock", "err", unlockErr)
		}
	}()

	if err := exec.Exec(ctx, schemaCreateSQL); err != nil {
		return errors.Wrap(err, "failed to create schema_migrations table")
	}

	var gopgExists bool
	if err := exec.Get(ctx, &gopgExists, gopgExistsSQL); err != nil {
		return errors.Wrap(err, "failed to check for legacy gopg_migrations table")
	}
	if !gopgExists {
		return nil
	}

	m.log.InfoContext(ctx, "legacy gopg_migrations table found, seeding schema_migrations from it")
	if err := exec.Exec(ctx, seedFromGopgSQL); err != nil {
		return errors.Wrap(err, "failed to seed schema_migrations from gopg_migrations")
	}
	return nil
}
