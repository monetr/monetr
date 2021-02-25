package testutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
