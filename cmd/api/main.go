package main

import (
	"fmt"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/application"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/cache"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/jobs"
	"github.com/sirupsen/logrus"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
)

func main() {
	configuration := config.LoadConfiguration()

	log := logrus.NewEntry(logrus.StandardLogger())

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

	redisController, err := cache.NewRedisCache(log, configuration.Redis)
	if err != nil {
		panic(err)
	}

	job := jobs.NewJobManager(log, redisController.Pool(), db)
	defer job.Drain()

	c := controller.NewController(configuration, db)

	app := application.NewApp(configuration, c)

	app.Listen(":4000")
}
