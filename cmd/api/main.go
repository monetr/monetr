package main

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/v10/orm"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/application"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/cache"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/jobs"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/metrics"
	"github.com/plaid/plaid-go/plaid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
)

var (
	_ pg.QueryHook = &hooks{}
)

type hooks struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func (h *hooks) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}
	h.log.Trace(string(query))

	return ctx, nil
}

func (h *hooks) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	var queryType string
	switch query := event.Query.(type) {
	case string:
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
	h.stats.Queries.With(prometheus.Labels{
		"stmt": queryType,
	}).Inc()
	return nil
}

func main() {
	stats := metrics.NewStats()
	stats.Listen(":9000")
	defer stats.Close()

	configuration := config.LoadConfiguration()

	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.TraceLevel)
	log := logrus.NewEntry(logger)

	db := pg.Connect(&pg.Options{
		Addr: fmt.Sprintf("%s:%d",
			configuration.PostgreSQL.Address,
			configuration.PostgreSQL.Port,
		),
		User:            configuration.PostgreSQL.Username,
		Password:        configuration.PostgreSQL.Password,
		Database:        configuration.PostgreSQL.Database,
		ApplicationName: "harder - api",
	})

	db.AddQueryHook(&hooks{
		stats: stats,
		log:   log,
	})

	redisController, err := cache.NewRedisCache(log, configuration.Redis)
	if err != nil {
		panic(err)
	}
	defer redisController.Close()

	p, err := plaid.NewClient(plaid.ClientOptions{
		ClientID:    configuration.Plaid.ClientID,
		Secret:      configuration.Plaid.ClientSecret,
		Environment: configuration.Plaid.Environment,
		HTTPClient:  http.DefaultClient,
	})
	if err != nil {
		panic(err)
	}

	job := jobs.NewJobManager(log, redisController.Pool(), db, p, stats)
	defer job.Close()

	c := controller.NewController(configuration, db, job, p, stats)

	app := application.NewApp(configuration, c)

	app.Listen(":4000")
}
