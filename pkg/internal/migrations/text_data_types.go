package migrations

import (
	"fmt"

	"github.com/uptrace/bun/dialect"
)

type TextDataType interface {
	DataType
	Capacity() int
}

var (
	_ TextDataType = TextOrVarChar(0)
	_ TextDataType = VarChar(0)
)

// TextOrVarChar will result in a TEXT data type for PostgreSQL databases, with no size limit. But will result in a
// VARCHAR data type with the specified size limit for any other database.
type TextOrVarChar VarChar

func (v TextOrVarChar) dataType() {}

func (v TextOrVarChar) String() string {
	return fmt.Sprintf("TextOrVarChar(%d)", v)
}

func (v TextOrVarChar) DatabaseType(dbDialect dialect.Name) (string, error) {
	switch dbDialect {
	case dialect.PG:
		return "TEXT", nil
	default:
		return (VarChar)(v).DatabaseType(dbDialect)
	}
}

func (v TextOrVarChar) Capacity() int {
	return int(v)
}

type VarChar uint32

func (v VarChar) dataType() {}

func (v VarChar) String() string {
	return fmt.Sprintf("VarChar(%d)", v)
}

func (v VarChar) DatabaseType(dialect.Name) (string, error) {
	return fmt.Sprintf("VARCHAR(%d)", v), nil
}

func (v VarChar) Capacity() int {
	return int(v)
}
