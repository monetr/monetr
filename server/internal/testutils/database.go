package testutils

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"

	"log/slog"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/migrations"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	_ bun.QueryHook = &queryHook{}
)

type queryHook struct {
	log   *slog.Logger
	stats *metrics.Stats
}

func (q *queryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	queryId := gofakeit.UUID()[0:8]
	if event.Stash != nil {
		event.Stash["queryId"] = queryId
	} else {
		event.Stash = map[any]any{
			"queryId": queryId,
		}
	}

	q.log.Log(ctx, logging.LevelTrace, event.Query, "queryId", queryId)

	return ctx
}

func (q *queryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if q.stats != nil {
		q.stats.Queries.With(prometheus.Labels{}).Inc()
	}

	if event.Err != nil {
		log := q.log
		if event.Stash != nil {
			if queryId, ok := event.Stash["queryId"].(string); ok {
				log = log.With("queryId", queryId)
			}
		}
		log.WarnContext(ctx, "query failed", "err", event.Err)
	}
}

func GetPgDatabaseTxn(t *testing.T) bun.Tx {
	db := GetPgDatabase(t)

	txn, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err, "must begin transaction")

	t.Cleanup(func() {
		require.NoError(t, txn.Rollback(), "must rollback database transaction")
	})

	return txn
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

func GetPgOptions(_ *testing.T) []pgdriver.Option {
	port := myownsanity.CoalesceStrings(
		os.Getenv("MONETR_PG_PORT"),
		os.Getenv("POSTGRES_PORT"),
		"5432",
	)

	host := myownsanity.CoalesceStrings(
		os.Getenv("MONETR_PG_ADDRESS"),
		os.Getenv("POSTGRES_HOST"),
		"localhost",
	)

	return []pgdriver.Option{
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(net.JoinHostPort(host, port)),
		// The trailing PG* / "postgres" tiers mirror go-pg's own defaulting for
		// empty values; pgdriver panics on empty options instead of defaulting.
		pgdriver.WithUser(myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_USERNAME"), os.Getenv("POSTGRES_USER"), os.Getenv("PGUSER"), "postgres")),
		pgdriver.WithPassword(myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_PASSWORD"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("PGPASSWORD"), "postgres")),
		pgdriver.WithDatabase(myownsanity.CoalesceStrings(os.Getenv("MONETR_PG_DATABASE"), os.Getenv("POSTGRES_DB"), os.Getenv("PGDATABASE"), "postgres")),
		pgdriver.WithApplicationName("monetr - api - tests"),
		pgdriver.WithReadTimeout(0),
		pgdriver.WithWriteTimeout(0),
		pgdriver.WithInsecure(true),
	}
}

// connectBun opens a bun database handle for the provided pgdriver options.
func connectBun(options ...pgdriver.Option) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(options...))
	return bun.NewDB(sqldb, pgdialect.New())
}

type DatabaseOption uint8

const (
	IsolatedDatabase DatabaseOption = 1
)

func GetBadPgDatabase(t *testing.T) *bun.DB {
	connector := pgdriver.NewConnector(GetPgOptions(t)...)
	connector.Config().Dialer = func(_ context.Context, _, _ string) (net.Conn, error) {
		return nil, errors.New("forcing a bad connection")
	}
	db := bun.NewDB(sql.OpenDB(connector), pgdialect.New())
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func GetPgDatabase(t *testing.T, databaseOptions ...DatabaseOption) *bun.DB {
	testDatabases.lock.Lock()
	defer testDatabases.lock.Unlock()

	if db, ok := testDatabases.databases[t.Name()]; ok {
		return db
	}

	options := GetPgOptions(t)
	db := connectBun(options...)

	require.NoError(t, db.PingContext(context.Background()), "must ping database")

	log := GetLog(t)

	db.AddQueryHook(&queryHook{
		log: log,
	})

	var databaseToReturn *bun.DB
	databaseToReturn = db
	if len(databaseOptions) > 0 {
		for _, option := range databaseOptions {
			switch option {
			case IsolatedDatabase:
				log.DebugContext(context.Background(), "creating isolated database for test")
				databaseName := fmt.Sprintf("%x", sha256.Sum256([]byte(t.Name())))

				_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q;`, databaseName))
				require.NoError(t, err, "must be able to drop an isolated database if it exists")

				_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE %q;`, databaseName))
				require.NoError(t, err, "must be able to create the isolated database")

				isolatedOptions := append(GetPgOptions(t), pgdriver.WithDatabase(databaseName))
				databaseToReturn = connectBun(isolatedOptions...)
				databaseToReturn.AddQueryHook(&queryHook{
					log: log,
				})

				migrations.RunMigrations(t.Context(), log, databaseToReturn)

				t.Cleanup(func() {
					require.NoError(t, databaseToReturn.Close(), "must close the isolated database once we are done")
					_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q;`, databaseName))
					require.NoError(t, err, "must be able to drop an isolated database if it exists")
					require.NoError(t, db.Close(), "must close database connection")
				})
			}
		}
	} else {
		t.Cleanup(func() {
			require.NoError(t, db.Close(), "must close database connection")
		})
	}

	testDatabases.databases[t.Name()] = databaseToReturn

	return databaseToReturn
}

func MustInsert[T any](t *testing.T, model T) T {
	db := GetPgDatabase(t)
	result, err := db.NewInsert().Model(&model).Returning("*").Exec(context.Background())
	require.NoError(t, err, "must insert data")
	affected, err := result.RowsAffected()
	require.NoError(t, err, "must read affected rows")
	require.GreaterOrEqual(t, int64(1), affected, "must insert at least one row")
	return model
}

func MustRetrieve[T any](t *testing.T, model T) T {
	var result T
	db := GetPgDatabase(t)
	err := db.NewSelect().Model(&model).WherePK().Scan(context.Background(), &result)
	require.NoError(t, err, "must retrieve data")
	return result
}

func MustEz[T any](t *testing.T, generalFunction func() (T, error)) T {
	result, err := generalFunction()
	require.NoError(t, err, "function must succeed without an error")
	return result
}

func Must[T any, A any](t *testing.T, generalFunction func(arg A) (T, error), arg A) T {
	result, err := generalFunction(arg)
	require.NoError(t, err, "function must succeed without an error")
	return result
}

func MustUnmarshalJSON(t *testing.T, data []byte, destination any) {
	err := json.Unmarshal(data, destination)
	require.NoError(t, err, "must be able to unmarshal the provided json without an error")
}

func MustDBUpdate[T any](t *testing.T, model *T) {
	db := GetPgDatabase(t) // Don't need options, DB should already be in cache
	// No RETURNING: every column was just written from the model, and models
	// like Login intentionally omit table columns (crypt) that bun would refuse
	// to discard when scanning RETURNING *.
	result, err := db.NewUpdate().Model(model).WherePK().Exec(context.Background())
	require.NoError(t, err, "must be able to update record")
	affected, err := result.RowsAffected()
	require.NoError(t, err, "must read affected rows")
	require.EqualValues(t, 1, affected, "must have updated one record")
}

func MustDBInsert[T any](t *testing.T, model *T) {
	db := GetPgDatabase(t) // Don't need options, DB should already be in cache
	result, err := db.NewInsert().Model(model).Returning("*").Exec(context.Background())
	require.NoError(t, err, "must be able to create record")
	affected, err := result.RowsAffected()
	require.NoError(t, err, "must read affected rows")
	require.EqualValues(t, 1, affected, "must have created one record")
}

func MustDBRead[T any](t *testing.T, model T) T {
	db := GetPgDatabase(t) // Don't need options, DB should already be in cache
	var result T
	err := db.NewSelect().Model(&model).WherePK().Scan(context.Background(), &result)
	require.NoError(t, err, "must be able to read updated record")
	return result
}

func MustDBNotExist[T any](t *testing.T, model T) {
	db := GetPgDatabase(t) // Don't need options, DB should already be in cache
	exists, err := db.NewSelect().Model(&model).WherePK().Exists(context.Background())
	require.NoError(t, err, "must be able to read without an error")
	require.Falsef(t, exists, "%T must not exist", model)
}

func MustDBExist[T any](t *testing.T, model T) {
	db := GetPgDatabase(t) // Don't need options, DB should already be in cache
	exists, err := db.NewSelect().Model(&model).WherePK().Exists(context.Background())
	require.NoError(t, err, "must be able to read record from the database")
	require.Truef(t, exists, "%T must exist", model)
}
