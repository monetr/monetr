package main

import (
	"context"
	"fmt"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/application"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/cache"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/jobs"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
)

var (
	_ pg.QueryHook = &hooks{}
)

type hooks struct {
	log *logrus.Entry
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
	return nil
}

func main() {
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
		log: log,
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

	job := jobs.NewJobManager(log, redisController.Pool(), db, p)
	defer job.Close()

	c := controller.NewController(configuration, db, job, p)

	app := application.NewApp(configuration, c)

	app.Listen(":4000")
}
