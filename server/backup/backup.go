package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BackupManager struct {
	log         *logrus.Entry
	conf        config.Configuration
	db          pg.DBI
	kms         secrets.KeyManagement
	objects     storage.Storage
	destination BackupDestination
}

func NewBackupManager(
	log *logrus.Entry,
	configuration config.Configuration,
	db pg.DBI,
	kms secrets.KeyManagement,
	objects storage.Storage,
	destination BackupDestination,
) *BackupManager {
	return &BackupManager{
		log:         log,
		conf:        configuration,
		db:          db,
		kms:         kms,
		objects:     objects,
		destination: destination,
	}
}

func (b *BackupManager) Backup(ctx context.Context) error {
	log := b.log.WithContext(ctx)
	// Create a snapshot in PostgreSQL, this way we can maintain a consistent
	// database state over a long period of time depending on the amount of time
	// the backup actually takes.
	var snapshot struct {
		SnapshotID string `pg:"pg_export_snapshot"`
	}
	txn, err := b.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer txn.Rollback()
	if _, err := txn.QueryOneContext(
		ctx, &snapshot, `SELECT pg_export_snapshot();`,
	); err != nil {
		return errors.Wrap(err, "failed to create PostgreSQL snapshot for backup")
	}

	if snapshot.SnapshotID == "" {
		return errors.New("PostgreSQL did not create a snapshot for backup")
	}

	postgresBackup := newPostgresBackupSource(b.log, b.conf, txn)

	reader, writer := io.Pipe()
	gzipWriter := gzip.NewWriter(writer)
	tarWriter := tar.NewWriter(gzipWriter)

	// TODO before writing any data to the tarball, we need to persist information
	// about the monetr instance like the version, timestamp of the backup etc.

	// Need to do this async somehow
	// Backup the PostgreSQL database first.
	if err := postgresBackup.start(ctx, tarWriter); err != nil {
		log.WithError(err).Fatal("failed to backup database")
		return err
	}

	// As we read data from our data stores, write that data in a stream to the
	// destination.
	if err := b.destination.Start(ctx, reader); err != nil {
		return err
	}

	return nil
}
