package migrations

import (
	"context"
	"io/fs"
	"log/slog"
	"path"
	"regexp"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

// advisoryLockKey is the session-scoped Postgres advisory lock we hold while
// applying migrations. Stops two monetr processes (rolling restart, parallel
// test setups) from racing each other on the same database. The value itself is
// arbitrary, it just has to be unique within monetr.
const advisoryLockKey int64 = 20180110

const schemaDir = "schema"

// migrationFilenameRegex matches the four shapes of migration filenames we
// support:
//
//	YYYYMMDDNN_Name.up.sql
//	YYYYMMDDNN_Name.tx.up.sql
//	YYYYMMDDNN_Name.down.sql
//	YYYYMMDDNN_Name.tx.down.sql
//
// The version prefix is a fixed ten digits: an eight-digit date plus a
// two-digit per-day sequence. Pinning the width keeps a fat-fingered prefix
// from quietly parsing as some wildly different version, which has tripped us
// up before. Names may contain underscores (2021041102_Balances_View is real)
// but not dots, since we use dots to mark the tx/direction suffix.
var migrationFilenameRegex = regexp.MustCompile(`^(\d{10})_([^.]+)\.(tx\.)?(up|down)\.sql$`)

type migrationFile struct {
	Version       int64
	Name          string
	Transactional bool
	Direction     string
	Filename      string
}

func parseMigrationFilename(name string) (migrationFile, error) {
	// A successful match has exactly five elements: the whole string plus the
	// four capture groups. Checking the length rather than just != nil keeps the
	// match[1..4] indexing below safe even if someone later adds or drops a group
	// in the pattern.
	match := migrationFilenameRegex.FindStringSubmatch(name)
	if len(match) != 5 {
		return migrationFile{}, errors.Errorf("unrecognized migration filename %q", name)
	}

	version, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return migrationFile{}, errors.Wrapf(err, "invalid version in filename %q", name)
	}

	return migrationFile{
		Version:       version,
		Name:          match[2],
		Transactional: match[3] == "tx.",
		Direction:     match[4],
		Filename:      name,
	}, nil
}

// discoverMigrations walks fsys/schema, parses every filename, and returns the
// up-files sorted ascending by version. Down files are validated for shape but
// stripped out, since the runtime never invokes them.
func discoverMigrations(fsys fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(fsys, schemaDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read embedded migrations directory")
	}

	ups := make([]migrationFile, 0, len(entries))
	seen := make(map[int64]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, errors.Errorf("unexpected directory in migrations: %q", entry.Name())
		}
		mf, err := parseMigrationFilename(entry.Name())
		if err != nil {
			return nil, err
		}
		if mf.Direction != "up" {
			continue
		}
		if existing, ok := seen[mf.Version]; ok {
			return nil, errors.Errorf(
				"duplicate migration version %d in both %q and %q",
				mf.Version, existing, mf.Filename,
			)
		}
		seen[mf.Version] = mf.Filename
		ups = append(ups, mf)
	}

	sort.Slice(ups, func(i, j int) bool {
		return ups[i].Version < ups[j].Version
	})

	return ups, nil
}

// embeddedMigrationFiles is the parsed, sorted list of up-migrations baked into
// the binary. We resolve it once at package load via mustDiscoverMigrations so
// a malformed embedded filename blows up the moment the package is loaded (at
// build/test time, or worst case the very first instruction of a server boot)
// rather than lying in wait until someone actually tries to migrate.
var embeddedMigrationFiles = mustDiscoverMigrations(embeddedMigrations)

// mustDiscoverMigrations is discoverMigrations for the embedded set, where a
// parse failure is a programming mistake compiled into the build and there is
// nothing sensible to do but refuse to start.
func mustDiscoverMigrations(fsys fs.FS) []migrationFile {
	files, err := discoverMigrations(fsys)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse embedded migrations"))
	}
	return files
}

// applyMigrations runs every file in files whose version is greater than the
// current MAX(version) in schema_migrations, ascending. The advisory lock is
// held for the whole run so concurrent migrators (rolling restarts, parallel
// test DBs) just take turns.
//
// exec MUST be pinned to a single backend connection or the advisory lock won't
// gate anything past the lock acquisition itself.
func applyMigrations(
	ctx context.Context,
	log *slog.Logger,
	exec Executor,
	fsys fs.FS,
	files []migrationFile,
) (oldVersion, newVersion int64, err error) {
	if err := exec.Exec(ctx, "SELECT pg_advisory_lock(?)", advisoryLockKey); err != nil {
		return 0, 0, errors.Wrap(err, "failed to acquire migration advisory lock")
	}
	defer func() {
		if unlockErr := exec.Exec(ctx, "SELECT pg_advisory_unlock(?)", advisoryLockKey); unlockErr != nil {
			log.WarnContext(ctx, "failed to release migration advisory lock", "err", unlockErr)
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
				log.WarnContext(ctx,
					"migration file present but skipped because its version is at or below the currently applied max",
					"version", f.Version,
					"name", f.Name,
					"current", oldVersion,
				)
			}
			continue
		}

		log.InfoContext(ctx,
			"applying migration",
			"version", f.Version,
			"name", f.Name,
			"tx", f.Transactional,
		)

		content, err := fs.ReadFile(fsys, path.Join(schemaDir, f.Filename))
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
func ensureSchemaTable(ctx context.Context, log *slog.Logger, exec Executor) error {
	if err := exec.Exec(ctx, "SELECT pg_advisory_lock(?)", advisoryLockKey); err != nil {
		return errors.Wrap(err, "failed to acquire migration advisory lock")
	}
	defer func() {
		if unlockErr := exec.Exec(ctx, "SELECT pg_advisory_unlock(?)", advisoryLockKey); unlockErr != nil {
			log.WarnContext(ctx, "failed to release migration advisory lock", "err", unlockErr)
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

	log.InfoContext(ctx, "legacy gopg_migrations table found, seeding schema_migrations from it")
	if err := exec.Exec(ctx, seedFromGopgSQL); err != nil {
		return errors.Wrap(err, "failed to seed schema_migrations from gopg_migrations")
	}
	return nil
}
