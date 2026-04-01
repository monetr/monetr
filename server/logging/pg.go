package logging

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type pgLogger struct {
	log *slog.Logger
}

func (l *pgLogger) Printf(ctx context.Context, format string, v ...any) {
	// I'm making an assumption here that go-pg is only going to log something if
	// there is a problem, generally its a pretty quiet library.
	l.log.WarnContext(ctx, fmt.Sprintf(format, v...), "logger", "go-pg")
}

func NewPGLogger(log *slog.Logger) *pgLogger {
	return &pgLogger{log}
}

var (
	_ pg.QueryHook = &PostgresHooks{}
)

type PostgresHooks struct {
	log   *slog.Logger
	stats *metrics.Stats
}

func NewPostgresHooks(log *slog.Logger, stats *metrics.Stats) pg.QueryHook {
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
	cleanedQuery := strings.TrimSpace(strings.ToLower(string(query)))
	if cleanedQuery != "select 1" && !strings.HasSuffix(cleanedQuery, "/* no log */") {
		h.log.Log(ctx, LevelTrace, strings.TrimSpace(string(query)))
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

	return nil
}
