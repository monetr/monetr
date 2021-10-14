package migrations

import (
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type DataType interface {
	dataType()
	String() string
	DatabaseType(dbDialect dialect.Name) (string, error)
}

type ColumnFlags uint8

const (
	None          ColumnFlags = 0
	NotNull       ColumnFlags = 1
	AutoIncrement ColumnFlags = 2
	PrimaryKey    ColumnFlags = 4
)

type ColumnDefinition struct {
	Name         string
	Type         DataType
	Flags        ColumnFlags
	DefaultValue interface{}
}

func CreateColumnDefinition(create *bun.CreateTableQuery, definitions ...ColumnDefinition) *bun.CreateTableQuery {
	dbDialect := create.DB().Dialect().Name()
	for _, definition := range definitions {
		builder := strings.Builder{}
		builder.WriteRune('?')
		builder.WriteRune(' ')

		dataTypeString, err := definition.Type.DatabaseType(dbDialect)
		if err != nil {
			panic(err)
		}

		// If we are doing auto-increment on Postgres then we want to change the type to be BigSerial
		if dbDialect == dialect.PG && definition.Type == BigInteger && definition.Flags&AutoIncrement != 0 {
			dataTypeString = "BIGSERIAL"
		}

		builder.WriteString(dataTypeString)
		builder.WriteRune(' ')

		if definition.Flags&PrimaryKey != 0 {
			builder.WriteString(`PRIMARY KEY `)
		}

		if definition.Flags&NotNull != 0 {
			builder.WriteString(`NOT NULL `)
		}

		if dbDialect == dialect.MySQL && definition.Flags&AutoIncrement != 0 {
			builder.WriteString(`AUTO_INCREMENT `)
		}

		params := []interface{}{
			bun.Ident(definition.Name),
		}
		if definition.DefaultValue != nil {
			builder.WriteString(`DEFAULT ?`)
			params = append(params, definition.DefaultValue)
		}

		create.ColumnExpr(strings.TrimSpace(builder.String()), params...)
	}

	return create
}
