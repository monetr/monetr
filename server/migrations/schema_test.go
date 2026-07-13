package migrations

import (
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils/testlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests live in-package because they need to touch unexported helpers
// like applyMigrations and newPgExecutor. That's also why the database setup
// is duplicated here instead of pulled in from server/internal/testutils,
// testutils itself imports server/migrations and would form a cycle. The
// log helpers do come from testutils/testlog, which sits in its own sub-
// package precisely to keep itself cycle-free.

func testPgOptions(_ *testing.T) *pg.Options {
	// Precedence runs monetr's own vars first, then the POSTGRES_* names our
	// compose files set, then the standard libpq PG* vars that psql itself
	// honors (https://www.postgresql.org/docs/current/libpq-envars.html), then
	// a sensible default. That last libpq tier means a shell already pointed at
	// a database via psql can run these tests without any extra setup.
	return &pg.Options{
		Network: "tcp",
		Addr: net.JoinHostPort(
			myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_ADDRESS"), os.Getenv("POSTGRES_HOST"), os.Getenv("PGHOST"), "localhost"),
			myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_PORT"), os.Getenv("POSTGRES_PORT"), os.Getenv("PGPORT"), "5432"),
		),
		User:            myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_USERNAME"), os.Getenv("POSTGRES_USER"), os.Getenv("PGUSER"), "postgres"),
		Password:        myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_PASSWORD"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("PGPASSWORD")),
		Database:        myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_DATABASE"), os.Getenv("POSTGRES_DB"), os.Getenv("PGDATABASE"), "postgres"),
		ApplicationName: "monetr - migrations - tests",
	}
}

// newCleanDatabase spins up a fresh Postgres database for the test with
// nothing in it yet, no migrations applied. Cleanup drops the database when
// the test ends.
func newCleanDatabase(t *testing.T) *pg.DB {
	parentOpts := testPgOptions(t)
	parent := pg.Connect(parentOpts)
	require.NoError(t, parent.Ping(t.Context()))

	// Postgres truncates identifiers at 63 bytes. An FNV-1a 64-bit hash is 16
	// hex chars, so monetr_test_<hash> stays well under the cap while staying
	// unique per test name. No need for a crypto hash here, this is just a name.
	h := fnv.New64a()
	_, _ = h.Write([]byte(t.Name()))
	dbName := fmt.Sprintf("monetr_test_%x", h.Sum64())
	_, err := parent.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, dbName))
	require.NoError(t, err)
	_, err = parent.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbName))
	require.NoError(t, err)

	childOpts := *parentOpts
	childOpts.Database = dbName
	child := pg.Connect(&childOpts)
	require.NoError(t, child.Ping(t.Context()))

	t.Cleanup(func() {
		_ = child.Close()
		_, _ = parent.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, dbName))
		_ = parent.Close()
	})

	return child
}

func readSchemaVersions(t *testing.T, db *pg.DB) []int64 {
	var raw []int64
	_, err := db.QueryContext(
		t.Context(),
		&raw,
		`SELECT version FROM schema_migrations ORDER BY version`,
	)
	require.NoError(t, err)
	return raw
}

func TestApplyMigrations_TxRollbackOnFailure(t *testing.T) {
	db := newCleanDatabase(t)
	log := testlog.GetLog(t)

	// ensureSchemaTable takes the session-scoped advisory lock, so it has to
	// share the pinned connection with applyMigrations rather than borrow a
	// pool connection that might never release the lock.
	conn := db.Conn()
	defer conn.Close()
	pinned := newPgExecutor(conn)
	require.NoError(t, ensureSchemaTable(t.Context(), log, pinned))

	fsys := fstest.MapFS{
		"schema/2030010100_BadTx.tx.up.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE rollback_marker (id INT);
SELECT this_function_does_not_exist();`),
		},
	}
	files, err := discoverMigrations(fsys)
	require.NoError(t, err)

	_, _, err = applyMigrations(t.Context(), log, pinned, fsys, files)
	require.Error(t, err)

	var exists bool
	_, err = db.QueryOneContext(t.Context(), pg.Scan(&exists),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_schema = current_schema() AND table_name = 'rollback_marker')`,
	)
	require.NoError(t, err)
	assert.False(t, exists, "tx-wrapped failure must roll back the CREATE TABLE")

	var count int
	_, err = db.QueryOneContext(t.Context(), pg.Scan(&count),
		`SELECT COUNT(*) FROM schema_migrations WHERE version = 2030010100`,
	)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "failed tx must not record a version")
}

func TestApplyMigrations_NonTxApplied(t *testing.T) {
	db := newCleanDatabase(t)
	log := testlog.GetLog(t)

	// ensureSchemaTable takes the session-scoped advisory lock, so it has to
	// share the pinned connection with applyMigrations rather than borrow a
	// pool connection that might never release the lock.
	conn := db.Conn()
	defer conn.Close()
	pinned := newPgExecutor(conn)
	require.NoError(t, ensureSchemaTable(t.Context(), log, pinned))

	fsys := fstest.MapFS{
		"schema/2030010200_NonTx.up.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE nontx_marker (id INT);`),
		},
	}
	files, err := discoverMigrations(fsys)
	require.NoError(t, err)

	oldV, newV, err := applyMigrations(t.Context(), log, pinned, fsys, files)
	require.NoError(t, err)
	assert.Equal(t, int64(0), oldV)
	assert.Equal(t, int64(2030010200), newV)

	var exists bool
	_, err = db.QueryOneContext(t.Context(), pg.Scan(&exists),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_schema = current_schema() AND table_name = 'nontx_marker')`,
	)
	require.NoError(t, err)
	assert.True(t, exists, "non-tx migration must apply its body")

	var count int
	_, err = db.QueryOneContext(t.Context(), pg.Scan(&count),
		`SELECT COUNT(*) FROM schema_migrations WHERE version = 2030010200`,
	)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "non-tx migration must record its version")
}

func TestApplyMigrations_GapWarn(t *testing.T) {
	db := newCleanDatabase(t)
	log, hook := testlog.GetTestLog(t)

	// ensureSchemaTable takes the session-scoped advisory lock, so it has to
	// share the pinned connection with applyMigrations rather than borrow a
	// pool connection that might never release the lock.
	conn := db.Conn()
	defer conn.Close()
	pinned := newPgExecutor(conn)
	require.NoError(t, ensureSchemaTable(t.Context(), log, pinned))

	_, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (99999999999)`)
	require.NoError(t, err)

	ups, err := discoverMigrations(embeddedMigrations)
	require.NoError(t, err)

	oldV, newV, err := applyMigrations(t.Context(), log, pinned, embeddedMigrations, ups)
	require.NoError(t, err)
	assert.Equal(t, int64(99999999999), oldV)
	assert.Equal(t, int64(99999999999), newV)

	testlog.MustHaveLogMessage(t, hook,
		"migration file present but skipped because its version is at or below the currently applied max")
}

func TestNewMigrationsManager_SeedFromGopgMigrations(t *testing.T) {
	db := newCleanDatabase(t)
	log := testlog.GetLog(t)

	_, err := db.Exec(`CREATE TABLE gopg_migrations (
        id         serial,
        version    bigint,
        created_at timestamptz
    )`)
	require.NoError(t, err)

	originalCreated := time.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)
	_, err = db.Exec(
		`INSERT INTO gopg_migrations (version, created_at) VALUES
            (2021041100, ?),
            (2021050999, ?),
            (2023060100, NULL)`,
		originalCreated, originalCreated,
	)
	require.NoError(t, err)

	m, err := NewMigrationsManager(t.Context(), log, db)
	require.NoError(t, err)
	require.NotNil(t, m)

	type row struct {
		Version   int64     `pg:"version"`
		AppliedAt time.Time `pg:"applied_at"`
	}
	var got []row
	_, err = db.QueryContext(t.Context(), &got,
		`SELECT version, applied_at FROM schema_migrations
         WHERE version IN (2021041100, 2021050999, 2023060100)
         ORDER BY version`,
	)
	require.NoError(t, err)

	require.Len(t, got, 2, "2021050999 (deleted test migration) must be filtered out")
	assert.Equal(t, int64(2021041100), got[0].Version)
	assert.True(t, originalCreated.Equal(got[0].AppliedAt),
		"seeded applied_at must equal original created_at: got %s want %s",
		got[0].AppliedAt, originalCreated,
	)
	assert.Equal(t, int64(2023060100), got[1].Version)
	assert.False(t, got[1].AppliedAt.IsZero(),
		"null created_at must fall back to NOW(), not zero",
	)

	var still bool
	_, err = db.QueryOneContext(t.Context(), pg.Scan(&still), gopgExistsSQL)
	require.NoError(t, err)
	assert.True(t, still, "gopg_migrations must remain after seed")

	_, err = NewMigrationsManager(t.Context(), log, db)
	require.NoError(t, err)

	var rowCount int
	_, err = db.QueryOneContext(t.Context(), pg.Scan(&rowCount),
		`SELECT COUNT(*) FROM schema_migrations
         WHERE version IN (2021041100, 2021050999, 2023060100)`,
	)
	require.NoError(t, err)
	assert.Equal(t, 2, rowCount, "second seed call must not duplicate rows")
}

func TestUp_ConcurrentSafe(t *testing.T) {
	db := newCleanDatabase(t)
	log := testlog.GetLog(t)

	m1, err := NewMigrationsManager(t.Context(), log, db)
	require.NoError(t, err)
	m2, err := NewMigrationsManager(t.Context(), log, db)
	require.NoError(t, err)

	var wg sync.WaitGroup
	var err1, err2 error
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _, err1 = m1.Up(t.Context())
	}()
	go func() {
		defer wg.Done()
		_, _, err2 = m2.Up(t.Context())
	}()
	wg.Wait()

	require.NoError(t, err1)
	require.NoError(t, err2)

	versions := readSchemaVersions(t, db)
	ups, err := discoverMigrations(embeddedMigrations)
	require.NoError(t, err)

	require.Equal(t, len(ups), len(versions),
		"concurrent runs must not duplicate or skip rows")
	assert.Equal(t, ups[len(ups)-1].Version, versions[len(versions)-1])
}

func TestUp_FreshFullEmbed(t *testing.T) {
	db := newCleanDatabase(t)
	log := testlog.GetLog(t)

	m, err := NewMigrationsManager(t.Context(), log, db)
	require.NoError(t, err)

	oldV, newV, err := m.Up(t.Context())
	require.NoError(t, err)
	assert.Equal(t, int64(0), oldV)
	assert.Equal(t, int64(2026071201), newV)

	versions := readSchemaVersions(t, db)
	ups, err := discoverMigrations(embeddedMigrations)
	require.NoError(t, err)
	assert.Equal(t, len(ups), len(versions))

	// A second Up() against an already-current database must be a no-op.
	oldV2, newV2, err := m.Up(t.Context())
	require.NoError(t, err)
	assert.Equal(t, newV, oldV2)
	assert.Equal(t, newV, newV2)

	latest := m.LatestVersion()
	current, err := m.CurrentVersion(t.Context())
	require.NoError(t, err)
	assert.Equal(t, latest, current)
	assert.Equal(t, int64(2026071201), latest)
}
