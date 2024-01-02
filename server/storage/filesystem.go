package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type filesystemStorage struct {
	log           *logrus.Entry
	baseDirectory string
}

func NewFilesystemStorage(
	log *logrus.Entry,
	baseDirectory string,
) Storage {
	return &filesystemStorage{
		log:           log,
		baseDirectory: baseDirectory,
	}
}

func (f *filesystemStorage) Store(
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

	uri = fmt.Sprintf("file:///%s", key)
	filePath := path.Join(f.baseDirectory, key)
	directory := path.Dir(filePath)

	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create destination directory")
	}

	log := f.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"destination": uri,
		})

	span.SetData("destination", uri)

	log.Debug("writing file to filesystem")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return "", errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	_, err = io.Copy(file, buf)
	if err != nil {
		return "", errors.Wrap(err, "failed to copy buffer to file")
	}
	if err := file.Sync(); err != nil {
		return "", errors.Wrap(err, "failed to fsync file")
	}

	return uri, nil
}

func (f *filesystemStorage) Read(
	ctx context.Context,
	uri string,
) (buf io.ReadCloser, contentType ContentType, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	url, err := url.Parse(uri)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse file uri")
	}
	filePath := path.Join(f.baseDirectory, url.Path)

	contentType, err = getContentTypeByPath(filePath)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to determine content type")
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to open file")
	}

	return file, contentType, nil
}
