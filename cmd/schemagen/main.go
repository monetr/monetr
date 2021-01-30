package main

import (
	"flag"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
)

var (
	dropTables = flag.Bool("drop", false, "add DROP TABLE before CREATE statements")
)

func main() {
	flag.Parse()

	db := pg.Connect(&pg.Options{})

	for _, model := range models.AllModels {
		if *dropTables {
			query := orm.NewDropTableQuery(db.Model(model), &orm.DropTableOptions{
				IfExists: true,
				Cascade:  true,
			})
			fmt.Println(query.String() + ";")
		}

		query := orm.NewCreateTableQuery(db.Model(model), &orm.CreateTableOptions{
			Varchar:       0,
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})
		fmt.Println(query.String() + ";")
	}
}
