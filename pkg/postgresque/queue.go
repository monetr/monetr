package postgresque

import (
	"context"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/postgresque/queue_migrations"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const MigrationsTableName = "gopg_queue_migrations"

type Queue struct {
	log *logrus.Entry
	db  *pg.DB
}

func NewPostgresQueue(log *logrus.Entry, db *pg.DB) *Queue {
	return &Queue{
		log: log,
		db:  db,
	}
}

func (q *Queue) SetupQueue(ctx context.Context) error {
	log := q.log.WithContext(ctx)
	collection := migrations.NewCollection(queue_migrations.PostgresQueueMigrations...)
	collection.SetTableName(MigrationsTableName)
	if _, _, err := collection.Run(q.db, "init"); err != nil {
		return errors.Wrap(err, "failed to initialize postgres queue schema migrations")
	}

	log.WithField("migrations", len(collection.Migrations())).Trace("found postgres queue migration(s)")

	currentVersion, err := collection.Version(q.db)
	if err != nil {
		return errors.Wrap(err, "failed to determine current postgres queue schema version")
	}

	log.WithField("currentVersion", currentVersion).Trace("current postgres queue schema version")

	oldVersion, newVersion, err := collection.Run(q.db, "up")
	if err != nil {
		return errors.Wrap(err, "failed to update postgres queue schema")
	}

	if oldVersion == newVersion {
		log.Debug("no schema changes applied for postgres queue")
	} else {
		log.WithFields(logrus.Fields{
			"oldVersion": oldVersion,
			"newVersion": newVersion,
		}).Info("updated postgres queue schema")
	}

	return nil
}

func (q *Queue) Enqueue(ctx context.Context, name string, arguments interface{}) (id string, err error) {
	return "", nil
}
