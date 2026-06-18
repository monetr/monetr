package backup

import (
	"context"
	"io"
	"strings"

	"github.com/monetr/monetr/server/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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

func NewBackupDestination(log *logrus.Entry, configuration config.Configuration) (BackupDestination, error) {
	switch strings.ToLower(configuration.Backup.Kind) {
	case "s3":
		return nil, nil
	case "filesystem":
		return &filesystemBackupDestination{
			log:       log,
			chunkSize: configuration.Backup.ChunkSize,
			path:      configuration.Backup.Prefix,
		}, nil
	default:
		return nil, errors.Errorf("invalid backup kind specified: %s", configuration.Backup.Kind)
	}
}
