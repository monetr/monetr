package main

import (
	"flag"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
)

var (
	address    = flag.String("address", "", "PostgreSQL address")
	port       = flag.Int("port", 5432, "PostgreSQL port")
	username   = flag.String("user", "postgres", "PostgreSQL username")
	database   = flag.String("db", "postgres", "PostgreSQL database")
	password   = flag.String("password", "", "PostgreSQL password")
	dryRun     = flag.Bool("dry-run", true, "Simply print the schema to STDOUT rather than applying it to the database")
	dropTables = flag.Bool("drop", false, "add DROP TABLE before CREATE statements")
)

var (
	createExtensions = []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE EXTENSION IF NOT EXISTS "citext":`,
	}
)

func main() {
	flag.Parse()

	var db *pg.DB
	if *dryRun {
		db = pg.Connect(&pg.Options{})
	} else {
		db = pg.Connect(&pg.Options{
			Addr:            fmt.Sprintf("%s:%d", *address, *port),
			User:            *username,
			Password:        *password,
			Database:        *database,
			ApplicationName: "harder - schemagen",
		})
	}

	for _, extension := range createExtensions {
		fmt.Println(extension + "\n")
		if !*dryRun {
			if _, err := db.Exec(extension); err != nil {
				panic(err)
			}
		}
	}

	for _, model := range models.AllModels {
		if *dropTables {
			query := orm.NewDropTableQuery(db.Model(model), &orm.DropTableOptions{
				IfExists: true,
				Cascade:  true,
			})
			fmt.Println(query.String() + ";\n")
			if !*dryRun {
				if _, err := db.Exec(query); err != nil {
					panic(err)
				}
			}
		}

		query := orm.NewCreateTableQuery(db.Model(model), &orm.CreateTableOptions{
			Varchar:       0,
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})
		fmt.Println(query.String() + ";\n")
		if !*dryRun {
			if _, err := db.Exec(query); err != nil {
				panic(err)
			}
		}
	}
}
