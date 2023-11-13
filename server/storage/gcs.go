package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type gcsStorage struct {
	log    *logrus.Entry
	bucket string
	client *storage.Client
}

func (s *gcsStorage) Store(ctx context.Context, buf io.ReadSeekCloser) (uri string, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key := getStorePath()
	uri = fmt.Sprintf("gcs://%s/%s", s.bucket, key)

	log := s.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"destination": uri,
		})

	span.SetData("destination", uri)

	log.Debug("uploading file to Google Cloud Storage")

	writer := s.client.Bucket(s.bucket).Object(key).NewWriter(span.Context())
	if _, err := io.Copy(writer, buf); err != nil {
		return "", errors.Wrap(err, "failed to write buffer to gcs writer")
	}

	return uri, errors.Wrap(writer.Close(), "failed to store file in gcs")
}
