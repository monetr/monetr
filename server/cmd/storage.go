package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// setupStorage is called when the monetr application starts running, it takes
// the configuration and sets up the appropriate storage interface based on that
// config. It also performs some basic validation of that storage configuration
// and if there is a problem it will return an error.
func setupStorage(
	log *logrus.Entry,
	configuration config.Configuration,
) (fileStorage storage.Storage, err error) {
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

		fileStorage = storage.NewS3StorageBackend(log, s3Config.Bucket, client)
	case "filesystem":
		log.Trace("setting up file storage interface using local filesystem")
		fileStorage, err = storage.NewFilesystemStorage(
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
