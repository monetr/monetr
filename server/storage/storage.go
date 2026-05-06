package storage

import (
	"context"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	monetrConfig "github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
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
	Store(ctx context.Context, buf io.Reader, file models.File) error
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
	log *slog.Logger,
	configuration monetrConfig.Configuration,
) (fileStorage Storage, err error) {
	if !configuration.Storage.Enabled {
		log.Log(context.Background(), logging.LevelTrace, "file storage is not enabled")
		return nil, nil
	}

	switch configuration.Storage.Provider {
	case "s3":
		log.Log(context.Background(), logging.LevelTrace, "setting up file storage interface using S3 protocol")
		s3Config := configuration.Storage.S3

		var configOptions []func(*config.LoadOptions) error

		configOptions = append(configOptions,
			config.WithLogger(logging.NewAWSLogger(log)),
			config.WithClientLogMode(aws.LogRetries|aws.LogDeprecatedUsage|aws.LogSigning),
		)

		if s3Config.Region != "" {
			configOptions = append(configOptions, config.WithRegion(s3Config.Region))
		}

		if s3Config.UseEnvCredentials {
			// Default config loading will pick up environment credentials automatically.
		} else if s3Config.AccessKey != nil {
			configOptions = append(configOptions, config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					*s3Config.AccessKey,
					*s3Config.SecretKey,
					"", // Not required since we aren't using temporary credentials.
				),
			))
		}

		cfg, err := config.LoadDefaultConfig(context.Background(), configOptions...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create s3 session")
		}

		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			if endpoint := s3Config.Endpoint; endpoint != nil && *endpoint != "" {
				o.BaseEndpoint = aws.String(*endpoint)
			}

			o.UsePathStyle = s3Config.ForcePathStyle
		})
		fileStorage = NewS3StorageBackend(log, s3Config.Bucket, client)
	case "filesystem":
		log.Log(context.Background(), logging.LevelTrace, "setting up file storage interface using local filesystem")
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
