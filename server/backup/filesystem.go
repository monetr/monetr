package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type filesystemBackupDestination struct {
	log       *logrus.Entry
	chunkSize int
	path      string
}

func (d *filesystemBackupDestination) Start(ctx context.Context, reader io.Reader) error {
	fileName := fmt.Sprintf("monetr-backup-%s.tar.gz", time.Now().Format("20060102150405"))
	path := path.Join(d.path, fileName)
	log := d.log.WithContext(ctx).WithFields(logrus.Fields{
		"path": path,
	})

	log.Info("starting filesystem writer for backup destination")

	file, err := os.OpenFile(path, os.O_TRUNC|os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open/create backup destination file")
	}
	defer file.Close()

	chunk := make([]byte, d.chunkSize)
	for {
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "failed to read chunk from backup manager")
		}
		if n == 0 {
			if err == io.EOF {
				break
			}
			time.Sleep(1 * time.Second)
			continue
		}

		log.WithFields(logrus.Fields{
			"bytes": n,
		}).Debug("writing backup chunk to destination")

		w, err := file.Write(chunk[:n])
		if err != nil {
			return errors.Wrap(err, "failed to write to destination file")
		}

		if w != n {
			return errors.Errorf("chunk is %d bytes, but only %d bytes were written", n, w)
		}

		if err == io.EOF {
			break
		}
	}

	log.Info("finished writing backup file")

	return nil
}

func (d *filesystemBackupDestination) Close() error {
	// TODO Close file here instead?
	return nil
}
