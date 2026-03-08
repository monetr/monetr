package storage

import (
	"context"
	"io"

	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type s3Storage struct {
	log    *slog.Logger
	bucket string
	client *s3.Client
}

func NewS3StorageBackend(
	log *slog.Logger,
	bucket string,
	s3client *s3.Client,
) Storage {
	return &s3Storage{
		log:    log,
		bucket: bucket,
		client: s3client,
	}
}

func (s *s3Storage) Store(
	ctx context.Context,
	buf io.ReadSeekCloser,
	file models.File,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key, err := file.GetStorePath()
	if err != nil {
		return err
	}

	span.SetData("destination", key)

	s.log.DebugContext(span.Context(), "uploading file to S3", "uri", key)

	contentType := string(file.ContentType)
	_, err = s.client.PutObject(
		span.Context(),
		&s3.PutObjectInput{
			Body:        buf,
			Bucket:      &s.bucket,
			Key:         &key,
			ContentType: &contentType,
			Metadata: map[string]string{
				"fileId":    file.FileId.String(),
				"accountId": file.AccountId.String(),
			},
		},
	)

	return errors.Wrap(err, "failed to store file in s3")
}

func (s *s3Storage) Read(
	ctx context.Context,
	file models.File,
) (buf io.ReadCloser, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key, err := file.GetStorePath()
	if err != nil {
		return nil, err
	}

	result, err := s.client.GetObject(
		span.Context(),
		&s3.GetObjectInput{
			Bucket: &s.bucket,
			Key:    &key,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve object from s3")
	}

	return result.Body, nil
}

func (s *s3Storage) Head(
	ctx context.Context,
	file models.File,
) (exists bool, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key, err := file.GetStorePath()
	if err != nil {
		return false, err
	}

	result, err := s.client.HeadObject(
		span.Context(),
		&s3.HeadObjectInput{
			Bucket: &s.bucket,
			Key:    &key,
		},
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to head object in s3")
	}

	return result != nil, nil
}

func (s *s3Storage) Remove(
	ctx context.Context,
	file models.File,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key, err := file.GetStorePath()
	if err != nil {
		return err
	}

	result, err := s.client.DeleteObject(
		span.Context(),
		&s3.DeleteObjectInput{
			Bucket: &s.bucket,
			Key:    &key,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete file from s3")
	}

	s.log.DebugContext(span.Context(), "file was removed from storage",
		"uri", key,
		slog.Group("s3DeleteResult",
			"deleteMarker", result.DeleteMarker,
			"versionId", result.VersionId,
			"requestCharged", string(result.RequestCharged),
		),
	)

	return nil
}
