package testutils

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"sort"
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

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	logger.Formatter = &logrus.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              true,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          true,
		FullTimestamp:             false,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc: func(input []string) {
			keys := make([]string, 0, len(input)-1)
			for _, key := range input {
				if key == "msg" {
					continue
				}

				keys = append(keys, key)
			}
			sort.Strings(keys)
			keys = append(keys, "msg")
			copy(input, keys)
		},
		DisableLevelTruncation: false,
		PadLevelText:           false,
		QuoteEmptyFields:       false,
		FieldMap:               nil,
		CallerPrettyfier:       nil,
	}
	log := logrus.NewEntry(logger)
	db.AddQueryHook(&queryHook{
		log: log.WithField("test", t.Name()),
	})

	t.Cleanup(func() {
		require.NoError(t, db.Close(), "must close database connection")
	})

	return db
}
