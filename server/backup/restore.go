package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RestoreManager struct {
	log *logrus.Entry
}

func (r *RestoreManager) Restore(ctx context.Context) error {
	buffer := bytes.NewReader([]byte(""))
	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return errors.Wrap(err, "failed to create gzip reader")
	}
	tarReader := tar.NewReader(gzipReader)

	for {
		next, err := tarReader.Next()
		if err != nil {
			return errors.Wrap(err, "failed to read next file in tar")
		}
		fmt.Println(next)
	}
}
