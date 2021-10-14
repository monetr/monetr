package migrations

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/monetr/monetr/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

func getDatabaseName(t *testing.T) string {
	hash := md5.Sum([]byte(t.Name()))
	return fmt.Sprintf("%X", hash[:8])
}

func getTestingDatabase(t *testing.T, engine config.DatabaseEngine) *bun.DB {
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

	return db
}

func foreachDatabase(t *testing.T, testFunc func(t *testing.T, db *bun.DB)) {
	databases := []config.DatabaseEngine{
		config.PostgreSQLDatabaseEngine,
		config.MySQLDatabaseEngine,
		config.SQLiteDatabaseEngine,
	}
	for _, engine := range databases {
		t.Run(engine.String(), func(t *testing.T) {
			testFunc(t, getTestingDatabase(t, engine))
		})
	}
}

func TestSelectOne(t *testing.T) {
	foreachDatabase(t, func(t *testing.T, db *bun.DB) {
		result, err := db.Query(`SELECT 1`)
		assert.NoError(t, err, "failed to perform basic query")
		assert.NoError(t, result.Err(), "result has error")
	})
}
