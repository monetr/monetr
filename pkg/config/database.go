package config

import (
	"strings"

	"github.com/pkg/errors"
)

//go:generate stringer -type=DatabaseEngine -output=database.strings.go
type DatabaseEngine uint8

const (
	PostgreSQLDatabaseEngine DatabaseEngine = iota + 1
	MySQLDatabaseEngine
	SQLiteDatabaseEngine
)

func ParseDatabaseEngine(input string) (DatabaseEngine, error) {
	switch strings.ToLower(input) {
	case "pg", "postgres", "postgresql":
		return PostgreSQLDatabaseEngine, nil
	case "mysql", "mariadb":
		return MySQLDatabaseEngine, nil
	case "sqlite":
		return SQLiteDatabaseEngine, nil
	default:
		return 0, errors.Errorf("invalid database engine: %s", input)
	}
}
