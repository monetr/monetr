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
					rows[i].IdNew = identifier.New(identifier.LinkKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Secrets
				var rows []struct {
					tableName string `pg:"secrets"`

					IdOld uint64        `pg:"secret_id,pk"`
					IdNew identifier.ID `pg:"secret_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.SecretKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Bank Accounts
				var rows []struct {
					tableName string `pg:"bank_accounts"`

					IdOld uint64        `pg:"bank_account_id,pk"`
					IdNew identifier.ID `pg:"bank_account_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.BankAccountKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Transactions
				var rows []struct {
					tableName string `pg:"transactions"`

					IdOld uint64        `pg:"transaction_id,pk"`
					IdNew identifier.ID `pg:"transaction_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.TransactionKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Transaction clusters
				var rows []struct {
					tableName string `pg:"transaction_clusters"`

					IdOld uint64        `pg:"transaction_cluster_id,pk"`
					IdNew identifier.ID `pg:"transaction_cluster_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.TransactionClusterKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Spending
				var rows []struct {
					tableName string `pg:"spending"`

					IdOld uint64        `pg:"spending_id,pk"`
					IdNew identifier.ID `pg:"spending_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.SpendingKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Funding Schedules
				var rows []struct {
					tableName string `pg:"funding_schedules"`

					IdOld uint64        `pg:"funding_schedule_id,pk"`
					IdNew identifier.ID `pg:"funding_schedule_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.FundingSchedulekind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Files
				var rows []struct {
					tableName string `pg:"files"`

					IdOld uint64        `pg:"file_id,pk"`
					IdNew identifier.ID `pg:"file_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.FileKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Jobs
				var rows []struct {
					tableName string `pg:"jobs"`

					IdOld uint64        `pg:"job_id,pk"`
					IdNew identifier.ID `pg:"job_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.JobKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Plaid Links
				var rows []struct {
					tableName string `pg:"plaid_links"`

					IdOld uint64        `pg:"plaid_link_id,pk"`
					IdNew identifier.ID `pg:"plaid_link_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.PlaidLinkKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Plaid Syncs
				var rows []struct {
					tableName string `pg:"plaid_syncs"`

					IdOld uint64        `pg:"plaid_sync_id,pk"`
					IdNew identifier.ID `pg:"plaid_sync_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.PlaidSyncKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Plaid Bank Accounts
				var rows []struct {
					tableName string `pg:"plaid_bank_accounts"`

					IdOld uint64        `pg:"plaid_bank_account_id,pk"`
					IdNew identifier.ID `pg:"plaid_bank_account_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.PlaidBankAccountKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Plaid Transactions
				var rows []struct {
					tableName string `pg:"plaid_transactions"`

					IdOld uint64        `pg:"plaid_transaction_id,pk"`
					IdNew identifier.ID `pg:"plaid_transaction_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.PlaidTransactionKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Teller Links
				var rows []struct {
					tableName string `pg:"teller_links"`

					IdOld uint64        `pg:"teller_link_id,pk"`
					IdNew identifier.ID `pg:"teller_link_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.TellerLinkKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Teller Bank Accounts
				var rows []struct {
					tableName string `pg:"teller_bank_accounts"`

					IdOld uint64        `pg:"teller_bank_account_id,pk"`
					IdNew identifier.ID `pg:"teller_bank_account_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.TellerBankAccountKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Teller Syncs
				var rows []struct {
					tableName string `pg:"teller_syncs"`

					IdOld uint64        `pg:"teller_sync_id,pk"`
					IdNew identifier.ID `pg:"teller_sync_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.TellerSyncKind)
				}
				if _, err := db.Model(&rows).WherePK().Update(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update %s", db.Model(&rows).TableModel().Table().SQLName))
				}
			}

			{ // Teller Tranasactions
				var rows []struct {
					tableName string `pg:"teller_transacations"`

					IdOld uint64        `pg:"teller_transacation_id,pk"`
					IdNew identifier.ID `pg:"teller_transacation_id_new"`
				}
				if err := db.Model(&rows).Select(&rows); err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to query %s", db.Model(&rows).TableModel().Table().SQLName))
				}
				for i := range rows {
					rows[i].IdNew = identifier.New(identifier.TellerTransactionKind)
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
