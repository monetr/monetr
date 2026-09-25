package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPgDatabaseTxn(t *testing.T) {
	txn := GetPgDatabaseTxn(t)
	assert.NotNil(t, txn, "txn must not be nil")
	var one int
	err := txn.NewRaw(`SELECT 1 as one;`).Scan(t.Context(), &one)
	assert.NoError(t, err, "should succeed")
	assert.Equal(t, 1, one, "should equal one")
}

func TestGetPgDatabase(t *testing.T) {
	originalDb := GetPgDatabase(t)
	assert.NotNil(t, originalDb, "database must not be nil")

	{ // Make sure the database actually works.
		var one int
		err := originalDb.NewRaw(`SELECT 1 as one;`).Scan(t.Context(), &one)
		assert.NoError(t, err, "should succeed")
		assert.Equal(t, 1, one, "should equal one")
	}

	secondDb := GetPgDatabase(t)
	assert.Equal(t, originalDb, secondDb, "should return the same database if called again")

	t.Run("separate test", func(t *testing.T) {
		separateDb := GetPgDatabase(t)
		assert.NotEqual(t, originalDb, separateDb, "requesting a db in a separate test should be different")
	})

	t.Run("isolated db", func(t *testing.T) {
		isolatedDb := GetPgDatabase(t, IsolatedDatabase)

		var originalName, isolatedName string
		assert.NoError(t, originalDb.NewRaw(`SELECT current_database();`).Scan(t.Context(), &originalName))
		assert.NoError(t, isolatedDb.NewRaw(`SELECT current_database();`).Scan(t.Context(), &isolatedName))
		assert.NotEqual(t, originalName, isolatedName, "database name should not be equal")
	})
}
