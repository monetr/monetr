package main

import (
	"fmt"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/application"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
)

func main() {
	configuration := config.LoadConfiguration()
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

	c := controller.NewController(configuration, db)

	app := application.NewApp(configuration, c)

	app.Listen(":4000")
}
