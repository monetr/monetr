package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPgDatabaseTxn(t *testing.T) {
	txn := GetPgDatabaseTxn(t)
	assert.NotNil(t, txn, "txn must not be nil")
	var ping struct {
		One int `pg:"one"`
	}
	result, err := txn.QueryOne(&ping, `SELECT 1 as one;`)
	assert.NoError(t, err, "should succeed")
	assert.Equal(t, 1, result.RowsReturned(), "should return one row")
	assert.Equal(t, 1, ping.One, "should equal one")
}

func TestGetPgDatabase(t *testing.T) {
	originalDb := GetPgDatabase(t)
	assert.NotNil(t, originalDb, "database must not be nil")

	{ // Make sure the database actually works.
		var ping struct {
			One int `pg:"one"`
		}
		result, err := originalDb.QueryOne(&ping, `SELECT 1 as one;`)
		assert.NoError(t, err, "should succeed")
		assert.Equal(t, 1, result.RowsReturned(), "should return one row")
		assert.Equal(t, 1, ping.One, "should equal one")
	}

	secondDb := GetPgDatabase(t)
	assert.Equal(t, originalDb, secondDb, "should return the same database if called again")

	t.Run("separate test", func(t *testing.T) {
		separateDb := GetPgDatabase(t)
		assert.NotEqual(t, originalDb, separateDb, "requesting a db in a separate test should be different")
	})

	t.Run("isolated db", func(t *testing.T) {
		isolatedDb := GetPgDatabase(t, IsolatedDatabase)
		assert.NotEqual(t, originalDb.Options().Database, isolatedDb.Options().Database, "database name should not be equal")
	})
}
