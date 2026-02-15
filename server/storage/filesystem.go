package storage

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const filesystemPermissions = 0755

type filesystemStorage struct {
	log           *logrus.Entry
	baseDirectory string
}

func NewFilesystemStorage(
	log *logrus.Entry,
	baseDirectory string,
) (Storage, error) {
	// This will make it so that the base directory path cannot be a relative path
	// or be blank.
	if !path.IsAbs(baseDirectory) {
		return nil, errors.New("base directory for filesystem storage must be an absolute path")
	}

	// The base directory path must also exist and be accessible.
	stat, err := os.Stat(baseDirectory)
	if err != nil {
		// If the directory does not exist
		if os.IsNotExist(err) {
			// Then create it using the same permissions we would use for the files.
			if err := os.MkdirAll(baseDirectory, filesystemPermissions); err != nil {
				return nil, errors.Wrap(err, "failed to create storage directory")
			}
		} else {
			return nil, errors.Wrap(err, "must provide a valid base directory for filesystem storage")
		}
	}
	// If the path exists and is not a directory then that's a problem.
	if stat != nil && !stat.IsDir() {
		return nil, errors.New("filesystem base directory specified is not a directory, it is a file")
	}

	return &filesystemStorage{
		log:           log,
		baseDirectory: baseDirectory,
	}, nil
}

func (f *filesystemStorage) Store(
	ctx context.Context,
	buf io.ReadSeekCloser,
	file models.File,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	filePath, err := f.getFullPath(file)
	if err != nil {
		return err
	}
	directory := path.Dir(filePath)

	if err := os.MkdirAll(directory, filesystemPermissions); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	log := f.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"uri": filePath,
		})

	span.SetData("destination", filePath)

	log.Debug("writing file to filesystem")

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, filesystemPermissions)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer fd.Close()

	_, err = io.Copy(fd, buf)
	if err != nil {
		return errors.Wrap(err, "failed to copy buffer to file")
	}
	if err := fd.Sync(); err != nil {
		return errors.Wrap(err, "failed to fsync file")
	}

	return nil
}

// Read will open a file on the filesystem available to the current process as
// read only and return it as an io.ReadCloser. If the file cannot be found or
// opened then an error will be returned.
func (f *filesystemStorage) Read(
	ctx context.Context,
	file models.File,
) (buf io.ReadCloser, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	filePath, err := f.getFullPath(file)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(filePath, os.O_RDONLY, filesystemPermissions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	return fd, nil
}

func (f *filesystemStorage) Head(
	ctx context.Context,
	file models.File,
) (exists bool, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	filePath, err := f.getFullPath(file)
	if err != nil {
		return false, err
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return false, errors.Wrap(err, "failed to stat object on filesystem")
	}

	// If the path is not a directory and it exsts, then return true
	return !stat.IsDir(), nil
}

func (f *filesystemStorage) Remove(
	ctx context.Context,
	file models.File,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	filePath, err := f.getFullPath(file)
	if err != nil {
		return err
	}
	if err := os.Remove(filePath); err != nil {
		return errors.Wrap(err, "failed to remove file from filesystem")
	}

	f.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"uri": filePath,
	}).Debug("file was removed from storage")

	return nil
}

func (f *filesystemStorage) getFullPath(file models.File) (string, error) {
	storedPath, err := file.GetStorePath()
	if err != nil {
		return "", errors.Wrap(err, "failed to determine file path")
	}

	return path.Join(f.baseDirectory, storedPath), nil
}
