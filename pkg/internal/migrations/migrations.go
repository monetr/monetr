package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/monetrapp/rest-api/pkg/internal/migrations/functional"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
)

type MonetrMigrationsManager struct {
	collection *migrations.Collection
	log        *logrus.Entry
	db         migrations.DB
}

func NewMigrationsManager(log *logrus.Entry, db migrations.DB) (*MonetrMigrationsManager, error) {
	collection := migrations.NewCollection(functional.FunctionalMigrations...)
	if err := collection.DiscoverSQLMigrationsFromFilesystem(http.FS(things), "schema"); err != nil {
		return nil, errors.Wrap(err, "failed to discover embedded sql migrations")
	}

	if _, _, err := collection.Run(db, "init"); err != nil {
		return nil, errors.Wrap(err, "failed to initialize schema migrations")
	}

	return &MonetrMigrationsManager{
		collection: collection,
		log:        log,
		db:         db,
	}, nil
}

func (m *MonetrMigrationsManager) CurrentVersion() (int64, error) {
	currentVersion, err := m.collection.Version(m.db)

	return currentVersion, errors.Wrap(err, "failed to get current database version")
}

func (m *MonetrMigrationsManager) LatestVersion() (int64, error) {
	var latest int64
	for _, migration := range m.collection.Migrations() {
		if migration.Version > latest {
			latest = migration.Version
		}
	}

	return latest, nil
}

func (m *MonetrMigrationsManager) Up() (oldVersion, newVersion int64, err error) {
	oldVersion, newVersion, err = m.collection.Run(m.db, "up")
	err = errors.Wrap(err, "failed to update database")
	return
}

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
