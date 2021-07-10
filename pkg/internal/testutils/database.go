package testutils

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var (
	_ pg.QueryHook = &queryHook{}
)

type queryHook struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func (q *queryHook) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}

	q.log.Trace(string(query))

	return ctx, nil
}

func (q *queryHook) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	if q.stats != nil {
		q.stats.Queries.With(prometheus.Labels{}).Inc()
	}

	return nil
}

func GetPgDatabaseTxn(t *testing.T) *pg.Tx {
	db := GetPgDatabase(t)

	txn, err := db.Begin()
	require.NoError(t, err, "must begin transaction")

	t.Cleanup(func() {
		require.NoError(t, txn.Rollback(), "must rollback database transaction")
	})

	return txn
}

func GetPgDatabase(t *testing.T) *pg.DB {
	options := &pg.Options{
		Network:         "tcp",
		Addr:            os.Getenv("POSTGRES_HOST") + ":5432",
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		Database:        os.Getenv("POSTGRES_DB"),
		ApplicationName: "harder - api - tests",
	}
	db := pg.Connect(options)

	require.NoError(t, db.Ping(context.Background()), "must ping database")

	log := GetLog(t)

	db.AddQueryHook(&queryHook{
		log: log.WithField("test", t.Name()),
	})

	t.Cleanup(func() {
		require.NoError(t, db.Close(), "must close database connection")
	})

	return db
}
