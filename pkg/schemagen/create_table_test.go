package schemagen

import (
	"github.com/go-pg/pg/v10/orm"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTableQuery_String(t *testing.T) {
	t.Run("with unique", func(t *testing.T) {
		log := testutils.GetLog(t)

		type Login struct {
			tableName string `pg:"logins"`

			LoginId      uint64 `json:"loginId" pg:"login_id,notnull,pk,type:'bigserial'"`
			Email        string `json:"email" pg:"email,notnull,unique"`
			PasswordHash string `json:"-" pg:"password_hash,notnull"`
		}

		createTable := NewCreateTableQuery(&Login{}, orm.CreateTableOptions{
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})

		query := createTable.String()
		assert.NotEmpty(t, query, "query should not be empty")

		log.Debug(query)
	})
}
