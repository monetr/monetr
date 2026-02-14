package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type s3Storage struct {
	log     *logrus.Entry
	bucket  string
	session *s3.S3
}

func NewS3StorageBackend(
	log *logrus.Entry,
	bucket string,
	s3client *s3.S3,
) Storage {
	return &s3Storage{
		log:     log,
		bucket:  bucket,
		session: s3client,
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

	log := s.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"uri": key,
		})

	span.SetData("destination", key)

	log.Debug("uploading file to S3")

	_, err = s.session.PutObjectWithContext(
		span.Context(),
		&s3.PutObjectInput{
			Body:        buf,
			Bucket:      &s.bucket,
			Key:         &key,
			ContentType: aws.String(string(file.ContentType)),
			Metadata: map[string]*string{
				"fileId":    aws.String(file.FileId.String()),
				"accountId": aws.String(file.AccountId.String()),
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

	result, err := s.session.GetObjectWithContext(
		span.Context(),
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
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

	result, err := s.session.HeadObjectWithContext(
		span.Context(),
		&s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
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

	result, err := s.session.DeleteObjectWithContext(
		span.Context(),
		&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete file from s3")
	}

	s.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"uri": key,
		"s3DeleteResult": logrus.Fields{
			"deleteMarker":   result.DeleteMarker,
			"versionId":      result.VersionId,
			"requestCharged": result.RequestCharged,
		},
	}).Debug("file was removed from storage")

	return nil
}
