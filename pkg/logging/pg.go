package logging

import (
	"context"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	_ pg.QueryHook = &PostgresHooks{}
)

type PostgresHooks struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func NewPostgresHooks(log *logrus.Entry, stats *metrics.Stats) pg.QueryHook {
	return &PostgresHooks{
		log:   log,
		stats: stats,
	}
}

func (h *PostgresHooks) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}
	if strings.TrimSpace(strings.ToLower(string(query))) != "select 1" {
		h.log.WithContext(ctx).Trace(strings.TrimSpace(string(query)))
	}

	return ctx, nil
}

func (h *PostgresHooks) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	endTime := time.Now()
	var queryType string
	switch query := event.Query.(type) {
	case string:
		query = strings.TrimSpace(query)
		query = strings.ReplaceAll(query, "\n", " ")

		// Don't do anything with health check queries.
		if strings.ToLower(query) == "select 1" {
			return nil
		}

		switch strings.ToUpper(query) {
		case "BEGIN", "COMMIT", "ROLLBACK":
			// Do nothing we don't want to count these.
			return nil
		default:
			firstSpace := strings.IndexRune(query, ' ')
			queryType = strings.ToUpper(query[:firstSpace])
		}
	case *orm.SelectQuery:
		queryType = "SELECT"
	case *orm.InsertQuery:
		queryType = "INSERT"
	case *orm.UpdateQuery:
		queryType = "UPDATE"
	case *orm.DeleteQuery:
		queryType = "DELETE"
	default:
		queryType = "UNKNOWN"
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		unformattedQuery, err := event.UnformattedQuery()
		if err == nil && len(unformattedQuery) > 0 {
			queryString := string(unformattedQuery)
			queryTime := endTime.Sub(event.StartTime)

			if event.Err == nil {
				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Type:     "query",
					Category: "postgres",
					Message:  queryString,
					Data: map[string]interface{}{
						"queryTime": queryTime.String(),
					},
					Level:     "debug",
					Timestamp: event.StartTime,
				}, nil)
			} else {
				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Type:     "query",
					Category: "postgres",
					Message:  queryString,
					Data: map[string]interface{}{
						"queryTime": queryTime.String(),
						"error":     event.Err.Error(),
					},
					Level:     "error",
					Timestamp: event.StartTime,
				}, nil)
			}

			span := sentry.StartSpan(ctx, "db.sql.query")
			span.StartTime = event.StartTime
			span.Description = queryString
			span.SetTag("query", queryType)

			if event.Err == nil {
				span.Status = sentry.SpanStatusOK
			} else {
				span.Status = sentry.SpanStatusInternalError
			}

			defer span.Finish()
		}
	}

	if h.stats != nil {
		h.stats.Queries.With(prometheus.Labels{
			"stmt": queryType,
		}).Inc()
	}

	return nil
}
