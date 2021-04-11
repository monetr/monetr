package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
	"net/http"
)

func RunMigrations(log *logrus.Entry, db migrations.DB) {
	collection := migrations.NewCollection()
	collection.DiscoverSQLMigrationsFromFilesystem(http.FS(things), "schema")

	if _, _, err := collection.Run(db, "init"); err != nil {
		log.Fatalf("failed to init schema migrations: %+v", err)
		return
	}

	currentVersion, err := collection.Version(db)
	if err != nil {
		log.Fatalf("failed to get database version: %+v", err)
		return
	}

	log.Infof("current database version is %d", currentVersion)

	oldVersion, newVersion, err := collection.Run(db, "up")
	if err != nil {
		log.Fatalf("failed to run migrations: %+v", err)
		return
	}

	if oldVersion == newVersion {
		log.Info("no database updates")
	} else {
		log.Infof("database upgraded from %d to %d", oldVersion, newVersion)
	}
}
