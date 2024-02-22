package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"

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

func NewGCSStorageBackend(log *logrus.Entry, bucket string, client *storage.Client) Storage {
	return &gcsStorage{
		log:    log,
		bucket: bucket,
		client: client,
	}
}

func (s *gcsStorage) Store(
	ctx context.Context,
	buf io.ReadSeekCloser,
	info FileInfo,
) (uri string, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key, err := getStorePath(info)
	if err != nil {
		return "", err
	}
	uri = fmt.Sprintf("gs://%s/%s", s.bucket, key)

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

	writer.Attrs().ContentType = string(info.ContentType)
	writer.ContentType = string(info.ContentType)

	return uri, errors.Wrap(writer.Close(), "failed to store file in gcs")
}

func (s *gcsStorage) Read(
	ctx context.Context,
	uri string,
) (buf io.ReadCloser, contentType ContentType, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	url, err := url.Parse(uri)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse file uri")
	}

	// TODO Make sure the context here does not pass a bad timeout to the actual reader object, if it does it could cause
	// the reader itself to expire before we would want it to.
	reader, err := s.client.
		Bucket(url.Host).
		Object(url.Path).
		NewReader(span.Context())
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create object reader for google cloud storage")
	}

	return reader, ContentType(reader.Attrs.ContentType), nil
}
