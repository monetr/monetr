package storage

import (
	"context"
	"fmt"
	"io"

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

func (s *s3Storage) Store(ctx context.Context, buf io.ReadSeekCloser) (uri string, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key := getStorePath()
	uri = fmt.Sprintf("s3://%s/%s", s.bucket, key)

	log := s.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"destination": uri,
		})

	span.SetData("destination", uri)

	log.Debug("uploading file to S3")

	_, err = s.session.PutObject(&s3.PutObjectInput{
		Body:                    buf,
		Bucket:                  &s.bucket,
		Key:                     &key,
		Metadata:                map[string]*string{},
		SSEKMSEncryptionContext: nil,
		SSEKMSKeyId:             nil,
		ServerSideEncryption:    nil,
		StorageClass:            nil,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to store file in s3")
	}

	return uri, nil
}

func (s *s3Storage) Read(ctx context.Context, uri string) (buf io.ReadCloser, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	return nil, nil
}
