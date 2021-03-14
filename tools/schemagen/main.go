package main

import (
	"flag"
	"fmt"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/schemagen"
	"math"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
)

var (
	address      = flag.String("address", "", "PostgreSQL address")
	port         = flag.Int("port", 5432, "PostgreSQL port")
	username     = flag.String("user", "postgres", "PostgreSQL username")
	database     = flag.String("db", "postgres", "PostgreSQL database")
	password     = flag.String("password", "", "PostgreSQL password")
	dryRun       = flag.Bool("dry-run", true, "Simply print the schema to STDOUT rather than applying it to the database")
	dropTables   = flag.Bool("drop", false, "add DROP TABLE before CREATE statements")
	printChanges = flag.Bool("print", true, "print schema changes to STDOUT")
)

var (
	createExtensions = []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE EXTENSION IF NOT EXISTS "citext";`,
	}
)

func printMaybe(str string) {
	if *printChanges {
		fmt.Println(str)
	}
}

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
		printMaybe(extension + "\n")
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
			printMaybe(query.String() + ";\n")
			if !*dryRun {
				if _, err := db.Exec(query); err != nil {
					panic(err)
				}
			}
		}

		options := orm.CreateTableOptions{
			Varchar:       0,
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		}

		query := schemagen.NewCreateTableQuery(model, options)
		printMaybe(query.String() + "\n")
		if !*dryRun {
			if _, err := db.Exec(orm.NewCreateTableQuery(db.Model(model), &options)); err != nil {
				panic(err)
			}
		}
	}

	login := models.Login{
		LoginId:      math.MaxUint64,
		Email:        "support@harderthanitneedstobe.com",
		PasswordHash: string([]byte{0}), // Will never be valid.
	}
	{
		query := orm.NewInsertQuery(db.Model(&login))
		printMaybe(query.String() + ";")

		if !*dryRun {
			if _, err := db.Model(&login).Insert(&login); err != nil {
				panic(err)
			}
		}
	}

	account := models.Account{
		AccountId: math.MaxUint64,
		Timezone:  "UTC",
	}
	{
		query := orm.NewInsertQuery(db.Model(&account))
		printMaybe(query.String() + ";")

		if !*dryRun {
			if _, err := db.Model(&account).Insert(&account); err != nil {
				panic(err)
			}
		}
	}

	user := models.User{
		UserId:    math.MaxUint64,
		LoginId:   math.MaxUint64,
		AccountId: math.MaxUint64,
		FirstName: "System",
		LastName:  "Bot",
	}
	{
		query := orm.NewInsertQuery(db.Model(&user))
		printMaybe(query.String() + ";")

		if !*dryRun {
			if _, err := db.Model(&user).Insert(&user); err != nil {
				panic(err)
			}
		}
	}

}
