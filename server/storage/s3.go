package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/monetr/monetr/server/crumbs"
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
		session: &s3.S3{},
	}
}

func (s *s3Storage) Store(
	ctx context.Context,
	buf io.ReadSeekCloser,
	contentType ContentType,
) (uri string, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key, err := getStorePath(contentType)
	if err != nil {
		return "", err
	}
	uri = fmt.Sprintf("s3://%s/%s", s.bucket, key)

	log := s.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"destination": uri,
		})

	span.SetData("destination", uri)

	log.Debug("uploading file to S3")

	_, err = s.session.PutObjectWithContext(
		span.Context(),
		&s3.PutObjectInput{
			Body:        buf,
			Bucket:      &s.bucket,
			Key:         &key,
			ContentType: aws.String(string(contentType)),
			Metadata:    map[string]*string{},
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to store file in s3")
	}

	return uri, nil
}

func (s *s3Storage) Read(
	ctx context.Context,
	uri string,
) (buf io.ReadCloser, contentType ContentType, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	url, err := url.Parse(uri)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse file uri")
	}

	result, err := s.session.GetObjectWithContext(
		span.Context(),
		&s3.GetObjectInput{
			Bucket: aws.String(url.Host),
			Key:    aws.String(url.Path),
		},
	)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to retrieve object from s3")
	}

	return result.Body, ContentType(*result.ContentType), nil
}
