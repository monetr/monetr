package backup

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type postgresBackupSource struct {
	log       *logrus.Entry
	db        pg.DBI
	conf      config.Configuration
	chunkSize int
}

func newPostgresBackupSource(
	log *logrus.Entry,
	conf config.Configuration,
	db pg.DBI,
) *postgresBackupSource {
	return &postgresBackupSource{
		log:       log,
		conf:      conf,
		db:        db,
		chunkSize: conf.Backup.ChunkSize,
	}
}

func (s *postgresBackupSource) getTableOrder(ctx context.Context) ([]string, error) {
	query := `
		WITH RECURSIVE dependency_tree AS (
				-- Get all tables without foreign key dependencies (they can be created first)
				SELECT
						c.relname AS table_name,
						0 AS level
				FROM
						pg_class c
				LEFT JOIN
						pg_constraint p ON c.oid = p.conrelid AND p.contype = 'f'
				WHERE
						p.confrelid IS NULL
						AND c.relkind = 'r'
						AND c.relname NOT LIKE 'pg_%'
						AND c.relname NOT LIKE 'sql_%'
				
				UNION ALL
				
				-- Recursively get tables with dependencies
				SELECT
						c.relname AS table_name,
						dt.level + 1 AS level
				FROM
						pg_constraint p
				JOIN
						pg_class c ON c.oid = p.conrelid
				JOIN
						dependency_tree dt ON dt.table_name = p.confrelid::regclass::text
				WHERE
						p.contype = 'f'
						AND c.relkind = 'r'
		)
		SELECT
				table_name
		FROM
				dependency_tree
		GROUP BY 
			table_name
		ORDER BY
				MAX(level) ASC, table_name ASC;
	`

	var names []struct {
		TableName string `pg:"table_name"`
	}
	if _, err := s.db.QueryContext(ctx, &names, query); err != nil {
		return nil, errors.Wrap(err, "failed to find tables to backup")
	}

	tables := make([]string, len(names))
	for i := range names {
		tables[i] = names[i].TableName
	}
	return tables, nil
}

func (s *postgresBackupSource) start(
	ctx context.Context,
	snapshotId string,
	writer *tar.Writer,
) error {
	log := s.log.WithContext(ctx)

	tables, err := s.getTableOrder(ctx)
	if err != nil {
		panic(err)
	}

	for i := range tables {
		tablename := tables[i]

		pipeReader, pipeWriter := io.Pipe()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer pipeWriter.Close()
			result, err := s.db.CopyTo(pipeWriter, fmt.Sprintf("COPY %s TO STDOUT", tablename))
			if err != nil {
				panic(err)
			}
			fmt.Println("RESULT", result.RowsReturned(), result.RowsAffected())
			pipeWriter.Close()
		}()

		part := 0
		chunk := make([]byte, s.chunkSize)
		for {
			// Read our chunk from the STDOUT stream of pg_dump.
			n, err := pipeReader.Read(chunk)
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
				Name: fmt.Sprintf("data/database/%s/%08d.bin", tablename, part),
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
		wg.Wait()
	}

	log.Info("PostgreSQL data export completed")

	return nil
}
