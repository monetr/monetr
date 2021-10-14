package migrations

import (
	"fmt"

	"github.com/uptrace/bun/dialect"
)

var (
	_ DataType = Array{}
)

type Array struct {
	DataType
}

func (t Array) String() string {
	return "TextArray"
}

func (t Array) DatabaseType(dbDialect dialect.Name) (string, error) {
	innerTypeString, err := t.DataType.DatabaseType(dbDialect)
	if err != nil {
		return "", err
	}

	switch dbDialect {
	case dialect.PG:
		return fmt.Sprintf("%s[]", innerTypeString), nil
	default:
		panic("non text arrays are not implemented")
	}
}
