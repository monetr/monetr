package testutils

import (
	"context"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

var (
	_ pg.QueryHook  = &pgQueryHook{}
	_ bun.QueryHook = &bunQueryHook{}
)

type pgQueryHook struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func (q *pgQueryHook) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
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

	q.log.WithField("queryId", queryId).Trace(string(query))

	return ctx, nil
}

func (q *pgQueryHook) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	if q.stats != nil {
		q.stats.Queries.With(prometheus.Labels{}).Inc()
	}

	if event.Err != nil {
		log := q.log
		if event.Stash != nil {
			if queryId, ok := event.Stash["queryId"].(string); ok {
				log = log.WithField("queryId", queryId)
			}
		}
		log.WithError(event.Err).Warn("query failed")
	}

	return nil
}

type bunQueryHook struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func (q *bunQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	queryId := gofakeit.UUID()[0:8]
	if event.Stash != nil {
		event.Stash["queryId"] = queryId
	} else {
		event.Stash = map[interface{}]interface{}{
			"queryId": queryId,
		}
	}

	q.log.WithField("queryId", queryId).Trace(event.Query)

	return ctx
}

func (q *bunQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if q.stats != nil {
		q.stats.Queries.With(prometheus.Labels{}).Inc()
	}

	if event.Err != nil {
		log := q.log
		if event.Stash != nil {
			if queryId, ok := event.Stash["queryId"].(string); ok {
				log = log.WithField("queryId", queryId)
			}
		}
		log.WithError(event.Err).Warn("query failed")
	}
}
