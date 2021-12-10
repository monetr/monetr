package migrations

import (
	"github.com/pkg/errors"
	"github.com/uptrace/bun/dialect"
)

var (
	_ DataType = BasicDataType(0)
)

//go:generate stringer -type=BasicDataType -output=basic_data_types.strings.go
type BasicDataType uint16

const (
	SmallInteger BasicDataType = iota
	Integer
	BigInteger
	Boolean
	Blob
	Date
	Timestamp
	JSON
)

var (
	basicDataTypeMappings = map[dialect.Name]map[BasicDataType]string{
		dialect.MySQL: {
			SmallInteger: "SMALLINT",
			Integer:      "INTEGER",
			BigInteger:   "BIGINT",
			Boolean:      "BOOLEAN",
			Blob:         "BLOB",
			Date:         "DATE",
			Timestamp:    "DATETIME", // https://dev.mysql.com/doc/refman/8.0/en/datetime.html
			JSON:         "BLOB",
		},
		dialect.PG: {
			SmallInteger: "SMALLINT",
			Integer:      "INTEGER",
			BigInteger:   "BIGINT",
			Boolean:      "BOOLEAN",
			Blob:         "BYTEA",
			Date:         "DATE",
			Timestamp:    "TIMESTAMP",
			JSON:         "JSONB",
		},
		dialect.SQLite: {
			SmallInteger: "SMALLINT",
			Integer:      "INTEGER",
			BigInteger:   "INTEGER",
			Boolean:      "BOOLEAN",
			Blob:         "BLOB",
			Date:         "DATE",
			Timestamp:    "DATETIME",
			JSON:         "BLOB",
		},
	}
)

func (i BasicDataType) dataType() {}

func (i BasicDataType) DatabaseType(dbDialect dialect.Name) (string, error) {
	databaseTypes, ok := basicDataTypeMappings[dbDialect]
	if !ok {
		return "", errors.New("database is not supported for data type")
	}

	typeString, ok := databaseTypes[i]
	if !ok {
		return "", errors.New("data type is not supported by this dialect")
	}

	return typeString, nil
}
