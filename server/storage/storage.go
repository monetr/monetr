package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Storage is the interface for reading and writing files presented to monetr by
// clients. These files might be images, CSV files or OFX files or something
// else entirely. Files are stored in a random path that is returned when the
// file is written. Files can only be retrieved using their path. Files cannot
// be listed and the tree cannot be walked. Files should only be interacted with
// via this interface within monetr.
//
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=storage.go -package=mockgen -destination=../internal/mockgen/storage.go Storage
type Storage interface {
	// Store will take a buffer and a file model and store the data from the
	// buffer using the underlying storage provider. If the data is stored
	// successfully then no error is returned. An error is returned if the file is
	// not able to be stored. Depending on the implementation the file may still
	// be present in whatever storage system even if the file was not successfully
	// stored. This should be considered on a per-implementation basis as it will
	// be unique to the implementation itself.
	Store(ctx context.Context, buf io.ReadSeekCloser, file models.File) error
	// Read will take a file model and will read it from the underlying storage
	// system. If a file can be read then a buffer will be returned for that file.
	Read(ctx context.Context, file models.File) (buf io.ReadCloser, err error)
	// Remove will take a file model and will remove it from the underlying
	// storage system. If this function returns nil, then the file was removed
	// successfully.
	Remove(ctx context.Context, file models.File) error
}

// GetStorage is meant to be called by the application initially starting up, it
// is simply provided a log entry and a configuration. It will the construct the
// appropriate storage driver implementation and return it to the caller. If an
// invalid configuration is provided then this will return an error. If storage
// is not configured then this will return nil.
func GetStorage(
	log *logrus.Entry,
	configuration config.Configuration,
) (fileStorage Storage, err error) {
	if !configuration.Storage.Enabled {
		log.Trace("file storage is not enabled")
		return nil, nil
	}

	switch configuration.Storage.Provider {
	case "s3":
		log.Trace("setting up file storage interface using S3 protocol")
		s3Config := configuration.Storage.S3
		awsConfig := aws.NewConfig().WithS3ForcePathStyle(s3Config.ForcePathStyle)
		if endpoint := s3Config.Endpoint; endpoint != nil {
			awsConfig = awsConfig.WithEndpoint(*endpoint)
		}

		if useEnvCredentials := s3Config.UseEnvCredentials; useEnvCredentials {
			awsConfig = awsConfig.WithCredentials(credentials.NewEnvCredentials())
		} else if s3Config.AccessKey != nil {
			awsConfig = awsConfig.WithCredentials(credentials.NewStaticCredentials(
				*s3Config.AccessKey,
				*s3Config.SecretKey,
				"", // Not requiured since we aren't using temporary credentials.
			))
		}

		if s3Config.Region != "" {
			awsConfig = awsConfig.WithRegion(s3Config.Region)
		}

		session, err := session.NewSession(awsConfig)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create s3 session")
		}

		client := s3.New(session)

		fileStorage = NewS3StorageBackend(log, s3Config.Bucket, client)
	case "filesystem":
		log.Trace("setting up file storage interface using local filesystem")
		fileStorage, err = NewFilesystemStorage(
			log,
			configuration.Storage.Filesystem.BasePath,
		)
	default:
		return nil, errors.Errorf(
			"invalid storage provider: %s",
			configuration.Storage.Provider,
		)
	}

	return fileStorage, err
}
