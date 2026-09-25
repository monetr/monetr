package logging

import (
	"context"
	"strings"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uptrace/bun"
)

var (
	_ bun.QueryHook = &PostgresHooks{}
)

type PostgresHooks struct {
	log   *slog.Logger
	stats *metrics.Stats
}

func NewPostgresHooks(log *slog.Logger, stats *metrics.Stats) bun.QueryHook {
	return &PostgresHooks{
		log:   log,
		stats: stats,
	}
}

func (h *PostgresHooks) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	query := event.Query
	cleanedQuery := strings.TrimSpace(strings.ToLower(query))
	if cleanedQuery != "select 1" && !strings.HasSuffix(cleanedQuery, "/* no log */") {
		h.log.Log(ctx, LevelTrace, strings.TrimSpace(query))
	}

	return ctx
}

func (h *PostgresHooks) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	endTime := time.Now()
	query := strings.TrimSpace(event.Query)
	query = strings.ReplaceAll(query, "\n", " ")

	// Don't do anything with health check queries.
	if strings.ToLower(query) == "select 1" {
		return
	}

	var queryType string
	switch strings.ToUpper(query) {
	case "BEGIN", "COMMIT", "ROLLBACK":
		// Do nothing we don't want to count these.
		return
	default:
		queryType = event.Operation()
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		unformattedQuery := event.QueryTemplate
		if len(unformattedQuery) > 0 {
			queryString := unformattedQuery
			queryTime := endTime.Sub(event.StartTime)

			if event.Err == nil {
				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Type:     "query",
					Category: "postgres",
					Message:  queryString,
					Data: map[string]any{
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
					Data: map[string]any{
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
			span.SetTag("db.system", "postgresql")
			span.SetTag("db.operation", queryType)

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
}
