package main

import (
	"context"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

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
	case "gcs":
		log.Trace("setting up file storage interface using GCS")

		gcsConfig := configuration.Storage.GCS

		options := make([]option.ClientOption, 0)
		if gcsConfig.URL != nil && *gcsConfig.URL != "" {
			options = append(options, option.WithEndpoint(*gcsConfig.URL))
		}

		if gcsConfig.APIKey != nil && *gcsConfig.APIKey != "" {
			options = append(options, option.WithAPIKey(*gcsConfig.APIKey))
		}

		if gcsConfig.CredentialsJSON != nil && *gcsConfig.CredentialsJSON != "" {
			options = append(options, option.WithCredentialsFile(*gcsConfig.CredentialsJSON))
		}

		client, err := gcs.NewClient(context.Background(), options...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize GCS client")
		}

		{ // Test bucket permissions
			requiredPermissions := []string{
				"storage.objects.create",
				"storage.objects.delete",
				"storage.objects.get",
				"storage.objects.list",
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			permissions, err := client.Bucket(gcsConfig.Bucket).
				IAM().
				TestPermissions(ctx, requiredPermissions)
			if err != nil {
				log.WithError(err).Fatal("failed to test permissions for google cloud storage")
				return nil, errors.Wrap(err, "failed to test permissions for google cloud storage")
			}

			// TODO Actually compare the subset of permissions instead of just the
			// length.
			if len(requiredPermissions) != len(permissions) {
				log.WithFields(logrus.Fields{
					"required": requiredPermissions,
					"have":     permissions,
				}).Fatal("permission mismatch for google cloud storage")
				return nil, errors.New("permission mismatch for google cloud storage")
			}
		}

		fileStorage = storage.NewGCSStorageBackend(log, gcsConfig.Bucket, client)
	case "filesystem":
		log.Trace("setting up file storage interface using local filesystem")
		fileStorage = storage.NewFilesystemStorage(
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
