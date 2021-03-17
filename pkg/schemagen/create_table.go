package schemagen

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/go-pg/pg/v10/types"
	"sort"
	"strings"
)

type CreateTableQuery struct {
	options orm.CreateTableOptions
	model   orm.TableModel
}

func NewCreateTableQuery(model interface{}, options orm.CreateTableOptions) *CreateTableQuery {
	return &CreateTableQuery{
		options: options,
		model:   pg.Model(model).TableModel(),
	}
}

func (c *CreateTableQuery) String() string {
	buffer := strings.Builder{}

	table := c.model.Table()

	unquotedTableName := strings.ReplaceAll(string(table.SQLName), "\"", "")

	buffer.WriteString("CREATE ")
	if c.options.Temp {
		buffer.WriteString(" TEMP ")
	}

	buffer.WriteString("TABLE ")
	if c.options.IfNotExists {
		buffer.WriteString("IF NOT EXISTS ")
	}

	buffer.WriteString(string(table.SQLName))
	buffer.WriteString(" (")

	for i, field := range table.Fields {
		if i > 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteString(string(field.Column))
		buffer.WriteRune(' ')

		if field.UserSQLType != "" {
			buffer.WriteString(strings.ToUpper(field.UserSQLType))
		} else {
			buffer.WriteString(strings.ToUpper(field.SQLType))
		}

		if strings.Contains(field.Field.Tag.Get("pg"), "notnull") {
			buffer.WriteString(" NOT NULL")
		}

		if field.Default != "" {
			buffer.WriteString(" DEFAULT ")
			buffer.WriteString(string(field.Default))
		}
	}

	if len(table.PKs) > 0 {
		buffer.WriteString(", CONSTRAINT \"pk_")
		buffer.WriteString(unquotedTableName)
		buffer.WriteString("\" PRIMARY KEY (")
		appendColumns(&buffer, "", table.PKs)
		buffer.WriteString(")")
	}

	if len(table.Unique) > 0 {
		uniqueConstraints := make([]string, 0, len(table.Unique))
		for _, unique := range table.Unique {

			uniqueString := strings.Builder{}

			uniqueString.WriteString(`, CONSTRAINT "uq_`)
			uniqueString.WriteString(unquotedTableName)
			uniqueString.WriteString("_")
			for i, field := range unique {
				if i > 0 {
					uniqueString.WriteString("_")
				}

				unquotedFieldName := strings.ReplaceAll(field.SQLName, `"`, "")
				uniqueString.WriteString(unquotedFieldName)
			}

			uniqueString.WriteString(`" UNIQUE (`)
			appendColumns(&uniqueString, "", unique)
			uniqueString.WriteString(")")

			uniqueConstraints = append(uniqueConstraints, uniqueString.String())
		}
		sort.Strings(uniqueConstraints)

		buffer.WriteString(strings.Join(uniqueConstraints, " "))
	}

	if c.options.FKConstraints {
		foreignConstraints := make([]string, 0, len(table.Relations))
		for _, relation := range table.Relations {
			if relation.Type != orm.HasOneRelation {
				continue
			}

			fkString := strings.Builder{}

			fkString.WriteString(`, CONSTRAINT "fk_`)
			fkString.WriteString(unquotedTableName)
			fkString.WriteString("_")
			unquotedJoinTable := strings.ReplaceAll(string(relation.JoinTable.SQLName), `"`, "")
			fkString.WriteString(unquotedJoinTable)
			for _, column := range relation.BaseFKs {
				fkString.WriteString("_")
				fkString.WriteString(strings.ReplaceAll(column.SQLName, `"`, ""))
			}
			fkString.WriteString(`" FOREIGN KEY (`)
			appendColumns(&fkString, "", relation.BaseFKs)
			fkString.WriteString(`) REFERENCES `)
			fkString.WriteString(string(relation.JoinTable.SQLName))
			fkString.WriteString(" (")
			appendColumns(&fkString, "", relation.JoinFKs)
			fkString.WriteString(")")

			foreignConstraints = append(foreignConstraints, fkString.String())
		}
		sort.Strings(foreignConstraints)

		buffer.WriteString(strings.Join(foreignConstraints, " "))
	}

	buffer.WriteString(");")

	return buffer.String()
}

func appendColumns(buffer *strings.Builder, table types.Safe, fields []*orm.Field) {
	for i, f := range fields {
		if i > 0 {
			buffer.WriteString(", ")
		}

		if len(table) > 0 {
			buffer.WriteString(string(table))
			buffer.WriteRune('.')
		}
		buffer.WriteString(string(f.Column))
	}
}
