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

// BackupDestination represents an implementation of the destination of a monetr
// backup. This could be a local filesystem or an object store. But it must
// consume from a reader that is passed to it. The backup destination
// implementation cannot know or rely on knowing how large the backup will be.
// But must be able to read from the stream of data to it in chunks.
type BackupDestination interface {
	// Start begins the backup writer, it consumes a reader and should read from
	// that reader in a chunk size that is up to the implementation. As it reads
	// chunks from the reader it should flush that data to whatever storage system
	// this destination represents.
	Start(ctx context.Context, reader io.Reader) error
	// Close cleans up any temporary files from the backup destination
	// implementation and finalizes anything needed for the backup. If the backup
	// is incomplete or was canceled then this should not remove any files written
	// (other than temp files).
	Close() error
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

	postgresBackup := newPostgresBackupSource(b.log, b.conf)

	reader, writer := io.Pipe()
	gzipWriter := gzip.NewWriter(writer)
	tarWriter := tar.NewWriter(gzipWriter)

	// TODO before writing any data to the tarball, we need to persist information
	// about the monetr instance like the version, timestamp of the backup etc.

	go func() {
		defer tarWriter.Close()
		defer gzipWriter.Close()
		defer writer.Close()

		// Backup the PostgreSQL database first.
		if err := postgresBackup.start(ctx, snapshot.SnapshotID, tarWriter); err != nil {
			log.WithError(err).Fatal("failed to backup database")
			return
		}
	}()

	// As we read data from our data stores, write that data in a stream to the
	// destination.
	if err := b.destination.Start(ctx, reader); err != nil {
		return err
	}

	return nil
}
