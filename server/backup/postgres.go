package backup

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/monetr/monetr/server/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type postgresBackupSource struct {
	log       *logrus.Entry
	conf      config.Configuration
	chunkSize int
}

func newPostgresBackupSource(
	log *logrus.Entry,
	conf config.Configuration,
) *postgresBackupSource {
	return &postgresBackupSource{
		log:  log,
		conf: conf,
	}
}

func (s *postgresBackupSource) start(
	ctx context.Context,
	snapshotId string,
	writer *tar.Writer,
) error {
	log := s.log.WithContext(ctx)

	dumpCommand := exec.CommandContext(
		ctx,
		"pg_dump",
		"-h", s.conf.PostgreSQL.Address,
		"-U", s.conf.PostgreSQL.Username,
		"-d", s.conf.PostgreSQL.Database,
		"-p", strconv.FormatInt(int64(s.conf.PostgreSQL.Port), 10),
		// By using snapshots our data should be consistent between different
		// processes.
		fmt.Sprintf("--snapshot=%s", snapshotId),
		"--if-exists", // Make it easier to restore
		"--no-privileges",
		"--no-owner",
		"--clean",            // Restoring should always overwrite the destination
		"--no-tablespaces",   // Tablespaces will vary by deployment
		"--no-subscriptions", // So will subscriptions and publications
		"--no-publications",
		"--disable-triggers",        // There _shouldn't_ be any triggers anyway?
		"--load-via-partition-root", // I believe this will fix any partition differences.
	)
	dumpCommand.Env = append(dumpCommand.Env,
		fmt.Sprintf("PGPASSWORD=%s", s.conf.PostgreSQL.Password),
	)

	stdout, err := dumpCommand.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get STDOUT pipe from pg_dump")
	}

	log.Info("starting PostgreSQL backup")
	log.Tracef("command: %s", dumpCommand.String())

	if err := dumpCommand.Start(); err != nil {
		return errors.Wrap(err, "failed to start pg_dump")
	}

	// Because we don't know how big the pg_dump output will actually be until it
	// has finished running, we are going to write the result in chunks to the tar
	// writer. This might cause some SQL statements to be clobbered in the middle
	// of the statement, but its not a problem because the restore process will
	// merge these chunks when it rewrites them.
	part := 0
	chunk := make([]byte, s.chunkSize)
	for {
		// Read our chunk from the STDOUT stream of pg_dump.
		n, err := stdout.Read(chunk)
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "failed to read chunk from pg_dump")
		}

		// If there wasn't a problem, and the chunk is not empty. Then we need to
		// write this chunk to the tar file.
		if n == 0 {
			break
		}

		log.WithField("bytes", n).Debug("data exported from PostgreSQL")

		// Write the header for this chunk.
		header := &tar.Header{
			Name: fmt.Sprintf("data/database/%08d.bin", part),
			Mode: 0600,
			// The size will be the number of bytes read, for most chunks this
			// _should_ be the same as the chunk size. But at the very end it is
			// unlikely that it will match up.
			Size: int64(n),
		}
		if err := writer.WriteHeader(header); err != nil {
			return errors.Wrap(err, "failed to write tar header for pg_dump chunk")
		}

		log.WithFields(logrus.Fields{
			"bytes": n,
			"name":  header.Name,
			"mode":  header.Mode,
		}).Trace("writing PostgreSQL export chunk")

		// Write the chunk dynamically sized to the tar file.
		if w, err := writer.Write(chunk[:n]); err != nil {
			return errors.Wrap(err, "failed to write chunk to tar for pg_dump")
		} else if w != n {
			return errors.Errorf("write mismatch, expected: %d bytes, got %d bytes", n, w)
		}

		// Make sure we increment the part we are on.
		part++
	}

	// Wait for the command to complete!
	if err := dumpCommand.Wait(); err != nil {
		return errors.Wrap(err, "failed to execute pg_dump")
	}

	log.Info("PostgreSQL data export completed")

	return nil
}
