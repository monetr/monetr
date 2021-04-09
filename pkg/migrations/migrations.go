package migrations

import (
	"fmt"
	"github.com/go-pg/migrations/v8"
	"log"
	"net/http"
)

func RunMigrations(db migrations.DB) {
	collection := migrations.NewCollection()
	collection.DiscoverSQLMigrationsFromFilesystem(http.FS(things), "schema")

	if _, _, err := collection.Run(db, "init"); err != nil {
		log.Fatalf("failed to init schema migrations: %+v", err)
		return
	}

	currentVersion, err := collection.Version(db)
	if err != nil {
		panic(err)
	}

	fmt.Printf("current database version is %d\n", currentVersion)
}
