package functional

import (
	"github.com/go-pg/migrations/v8"
	"github.com/monetr/monetr/server/identifier"
	"github.com/pkg/errors"
)

type Login struct {
	tableName string `pg:"logins"`

	LoginIdOld uint64 `pg:"login_id,notnull,pk,type:'bigserial'"`
	LoginIdNew string `pg:"login_id_new,notnull,pk"`
}

func init() {
	FunctionalMigrations = append(FunctionalMigrations, &migrations.Migration{
		Version: 2024041101,
		UpTx:    true,
		Up: func(db migrations.DB) error {
			{ // Logins
				var rows []struct {
					tableName string `pg:"logins"`

					IdOld uint64        `pg:"login_id"`
					IdNew identifier.ID `pg:"login_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, "failed to query logins")
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.LoginKind)
				}
			}

			{ // Accounts
				var rows []struct {
					tableName string `pg:"accounts"`

					IdOld uint64        `pg:"account_id"`
					IdNew identifier.ID `pg:"account_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, "failed to query accounts")
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.AccountKind)
				}
			}

			{ // Users
				var rows []struct {
					tableName string `pg:"users"`

					IdOld uint64        `pg:"user_id"`
					IdNew identifier.ID `pg:"user_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, "failed to query users")
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.UserKind)
				}
			}

			return nil
		},
		DownTx: false,
		Down: func(db migrations.DB) error {
			return nil
		},
	})
}
