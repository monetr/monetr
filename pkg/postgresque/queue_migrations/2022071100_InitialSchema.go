package queue_migrations

import (
	"github.com/go-pg/migrations/v8"
)

func init() {
	PostgresQueueMigrations = append(PostgresQueueMigrations, &migrations.Migration{
		Version: 2022071100,
		UpTx:    false,
		Up: func(db migrations.DB) error {
			db.Exec(`
				CREATE TABLE pg_queue (
					id TEXT NOT NULL,
					name TEXT NOT NULL,
					data JSONB,
					retry_limit INT NOT NULL DEFAULT (0),
					retry_count INT NOT NULL DEFAULT (0),
					retry_delay INT NOT NULL DEFAULT (30),
					start_after TIMESTAMP NOT NULL DEFAULT now(),
					singleton_key TEXT,
					created_at TIMESTAMP NOT NULL DEFAULT now(),
					started_at TIMESTAMP,
					completed_at TIMESTAMP,
					output JSONB,
					CONSTRAINT pk_pg_queue PRIMARY KEY (id),
					CONSTRAINT uq_pg_queue_singleton UNIQUE (singleton_key)
				);
			`)
			return nil
		},
		DownTx: false,
		Down: func(db migrations.DB) error {
			db.Exec(`
				DROP TABLE IF EXISTS pg_queue;
			`)
			return nil
		},
	})
}
