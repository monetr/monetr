package functional

import (
	"fmt"

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

					IdOld uint64        `pg:"login_id,pk"`
					IdNew identifier.ID `pg:"login_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.LoginKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Accounts
				var rows []struct {
					tableName string `pg:"accounts"`

					IdOld uint64        `pg:"account_id,pk"`
					IdNew identifier.ID `pg:"account_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.AccountKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Users
				var rows []struct {
					tableName string `pg:"users"`

					IdOld uint64        `pg:"user_id,pk"`
					IdNew identifier.ID `pg:"user_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.UserKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Links
				var rows []struct {
					tableName string `pg:"links"`

					IdOld uint64        `pg:"link_id,pk"`
					IdNew identifier.ID `pg:"link_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.UserKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Links
				var rows []struct {
					tableName string `pg:"secrets"`

					IdOld uint64        `pg:"secret_id,pk"`
					IdNew identifier.ID `pg:"secret_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.UserKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
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
