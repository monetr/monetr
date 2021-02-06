package testutils

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func GetPgDatabaseTxn(t *testing.T) *pg.Tx {
	options := &pg.Options{
		Network:         "tcp",
		Addr:            os.Getenv("POSTGRES_HOST") + ":5432",
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		Database:        os.Getenv("POSTGRES_DB"),
		ApplicationName: "harder - api - tests",
	}
	db := pg.Connect(options)

	require.NoError(t, db.Ping(context.Background()), "must ping database")

	txn, err := db.Begin()
	require.NoError(t, err, "must begin transaction")

	t.Cleanup(func() {
		require.NoError(t, txn.Rollback(), "must rollback database transaction")
		require.NoError(t, db.Close(), "must close database connection")
	})

	return txn
}
