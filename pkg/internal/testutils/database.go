package testutils

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/config"
	monetrMigrations "github.com/monetr/monetr/pkg/internal/migrations"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"

	_ "github.com/go-sql-driver/mysql"
)

var (
	_ pg.QueryHook = &queryHook{}
)

type queryHook struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func (q *queryHook) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	queryId := gofakeit.UUID()[0:8]
	if event.Stash != nil {
		event.Stash["queryId"] = queryId
	} else {
		event.Stash = map[interface{}]interface{}{
			"queryId": queryId,
		}
	}

	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}

	q.log.WithContext(ctx).WithField("queryId", queryId).Trace(string(query))

	return ctx, nil
}

func (q *queryHook) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	if q.stats != nil {
		q.stats.Queries.With(prometheus.Labels{}).Inc()
	}

	if event.Err != nil {
		log := q.log.WithContext(ctx)
		if event.Stash != nil {
			if queryId, ok := event.Stash["queryId"].(string); ok {
				log = log.WithField("queryId", queryId)
			}
		}
		log.WithError(event.Err).Warn("query failed")
	}

	return nil
}

var testDatabases struct {
	lock      sync.Mutex
	databases map[string]*bun.DB
}

func init() {
	testDatabases = struct {
		lock      sync.Mutex
		databases map[string]*bun.DB
	}{
		lock:      sync.Mutex{},
		databases: map[string]*bun.DB{},
	}
}

func getEnvDefault(envName, defaultValue string) string {
	if value := os.Getenv(envName); value != "" {
		return value
	}

	return defaultValue
}

func getDatabaseName(t *testing.T) string {
	hash := md5.Sum([]byte(t.Name()))
	return fmt.Sprintf("%X", hash[:8])
}

func getTestingDatabase(t *testing.T, engine config.DatabaseEngine) *bun.DB {
	testDatabases.lock.Lock()
	defer testDatabases.lock.Unlock()

	if db, ok := testDatabases.databases[t.Name()]; ok {
		return db
	}

	dbname := getDatabaseName(t)
	var db *bun.DB
	switch engine {
	case config.PostgreSQLDatabaseEngine:
		{
			dsn := "postgres://monetr:@localhost:5432/monetr?sslmode=disable"
			// dsn := "unix://user:pass@dbname/var/run/postgresql/.s.PGSQL.5432"
			tempDb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

			_, err := tempDb.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbname))
			require.NoError(t, err, "must drop database if it exists")
			_, err = tempDb.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbname))
			require.NoError(t, err, "must create database")

			t.Cleanup(func() {
				_, err = tempDb.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbname))
				require.NoError(t, err, "must drop database if it exists")
				require.NoError(t, tempDb.Close(), "must close PostgreSQL connection successfully")
			})
		}

		dsn := fmt.Sprintf("postgres://monetr:@localhost:5432/%s?sslmode=disable", dbname)
		postgresDb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
		t.Cleanup(func() {
			require.NoError(t, postgresDb.Close(), "must close PostgreSQL connection successfully")
		})

		db = bun.NewDB(postgresDb, pgdialect.New())
	case config.MySQLDatabaseEngine:
		{ // Create a test database and cleanup afterwards.
			tempDb, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/monetr")
			require.NoError(t, err, "must be able to MySQL database")

			_, err = tempDb.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname))
			require.NoError(t, err, "must drop database if it exists")
			_, err = tempDb.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
			require.NoError(t, err, "must create database")
			t.Cleanup(func() {
				_, err = tempDb.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname))
				require.NoError(t, err, "must drop database if it exists")
				require.NoError(t, tempDb.Close(), "must close MySQL connection successfully")
			})
		}

		mysqlDb, err := sql.Open("mysql", fmt.Sprintf("root:password@tcp(localhost:3306)/%s", dbname))
		require.NoError(t, err, "must be able to MySQL database")
		t.Cleanup(func() {
			require.NoError(t, mysqlDb.Close(), "must close MySQL connection successfully")
		})

		db = bun.NewDB(mysqlDb, mysqldialect.New())
	case config.SQLiteDatabaseEngine:
		sqliteDb, err := sql.Open(sqliteshim.DriverName(), "file::memory:?cache=shared")
		require.NoError(t, err, "must be able to create in-memory SQLite database")
		sqliteDb.SetMaxIdleConns(1000)
		sqliteDb.SetConnMaxLifetime(0)
		t.Cleanup(func() {
			require.NoError(t, sqliteDb.Close(), "must close in-memory SQLite database successfully")
		})

		db = bun.NewDB(sqliteDb, sqlitedialect.New())
	}

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(true),
		bundebug.WithVerbose(true),
		bundebug.FromEnv(""),
	))

	testDatabases.databases[t.Name()] = db

	t.Cleanup(func() {
		testDatabases.lock.Lock()
		defer testDatabases.lock.Unlock()
		delete(testDatabases.databases, t.Name())
	})

	return db
}

func ForEachDatabase(t *testing.T, innerTest func(ctx context.Context, t *testing.T, db *bun.DB)) {
	databases := []config.DatabaseEngine{
		config.PostgreSQLDatabaseEngine,
		// config.MySQLDatabaseEngine,
		config.SQLiteDatabaseEngine,
	}
	for _, engine := range databases {
		t.Run(engine.String(), func(innerT *testing.T) {
			ctx := context.Background()

			database := getTestingDatabase(innerT, engine)
			{
				migrator := migrate.NewMigrator(database, monetrMigrations.Migrations)
				require.NotNil(innerT, migrator, "database migrations must not be nil from bun")
				require.NoError(innerT, migrator.Init(ctx), "must init db")
				status, err := migrator.MigrationsWithStatus(ctx)
				require.NoError(innerT, err, "failed")
				require.NotNil(innerT, status)

				{ // Perform the migrations
					group, err := migrator.Migrate(ctx)
					require.NoError(innerT, err, "failed to migrate")
					require.NotNil(innerT, group, "migration group must not be nil")
				}

				// Cleanup the migrations once we are done.
				innerT.Cleanup(func() {
					group, err := migrator.Rollback(ctx)
					require.NoError(innerT, err, "must be able to rollback all migrations")
					require.NotNil(innerT, group, "migration group must not be nil")
				})
			}

			innerTest(ctx, innerT, database)
		})
	}
}

func ForEachDatabaseUnMigrated(t *testing.T, innerTest func(ctx context.Context, t *testing.T, db *bun.DB)) {
	databases := []config.DatabaseEngine{
		config.PostgreSQLDatabaseEngine,
		// config.MySQLDatabaseEngine,
		config.SQLiteDatabaseEngine,
	}
	for _, engine := range databases {
		t.Run(engine.String(), func(innerT *testing.T) {
			innerTest(context.Background(), innerT, getTestingDatabase(innerT, engine))
		})
	}
}

func GetDatabase(t *testing.T, engine config.DatabaseEngine) *bun.DB {
	return getTestingDatabase(t, engine)
}

func GetTestDatabase(t *testing.T) *bun.DB {
	testDatabases.lock.Lock()
	defer testDatabases.lock.Unlock()

	if db, ok := testDatabases.databases[t.Name()]; ok {
		return db
	}

	panic("database must be initialized for test first!")
}
