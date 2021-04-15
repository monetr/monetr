package main

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/cache"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/controller"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/monetrapp/rest-api/pkg/migrations"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	stats := metrics.NewStats()
	stats.Listen(":9000")
	defer stats.Close()

	configuration := config.LoadConfiguration(nil)

	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.TraceLevel)
	log := logrus.NewEntry(logger)

	pgOptions := &pg.Options{
		Addr: fmt.Sprintf("%s:%d",
			configuration.PostgreSQL.Address,
			configuration.PostgreSQL.Port,
		),
		User:            configuration.PostgreSQL.Username,
		Password:        configuration.PostgreSQL.Password,
		Database:        configuration.PostgreSQL.Database,
		ApplicationName: "harder - api",
	}

	db := pg.Connect(pgOptions)

	migrations.RunMigrations(log, db)

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
