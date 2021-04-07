package repository

import (
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

type JobRepository interface {
	GetBankAccountsToSync() ([]models.BankAccount, error)
	GetBankAccountsWithPendingTransactions() ([]CheckingPendingTransactionsItem, error)
	GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error)
}

type ProcessFundingSchedulesItem struct {
	AccountId          uint64   `pg:"account_id"`
	BankAccountId      uint64   `pg:"bank_account_id"`
	FundingScheduleIds []uint64 `pg:"funding_schedule_ids,type:bigint[]"`
}

type CheckingPendingTransactionsItem struct {
	AccountId uint64 `pg:"account_id"`
	LinkId    uint64 `pg:"link_id"`
}

type jobRepository struct {
	txn *pg.Tx
}

func NewJobRepository(txn *pg.Tx) JobRepository {
	return &jobRepository{
		txn: txn,
	}
}

func (j *jobRepository) GetBankAccountsToSync() ([]models.BankAccount, error) {
	var result []models.BankAccount
	err := j.txn.Model(&result).
		Relation("Link").
		Relation("Link.PlaidLink").
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve bank accounts to sync")
}

func (j *jobRepository) GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error) {
	var items []ProcessFundingSchedulesItem
	_, err := j.txn.Query(&items, `
		SELECT
			"funding_schedules"."account_id",
			"funding_schedules"."bank_account_id",
			array_agg("funding_schedules"."funding_schedule_id") AS "funding_schedule_ids"
		FROM "funding_schedules"
		WHERE "funding_schedules"."next_occurrence" < (now() AT TIME ZONE 'UTC')
		GROUP BY "funding_schedules"."account_id", "funding_schedules"."bank_account_id"
	`)
	if err != nil {
		// TODO (elliotcourant) Can pg.NoRows return here? If it can this error is useless.
		return nil, errors.Wrap(err, "failed to retrieve accounts and their funding schedules")
	}

	return items, nil
}

func (j *jobRepository) GetBankAccountsWithPendingTransactions() ([]CheckingPendingTransactionsItem, error) {
	var items []CheckingPendingTransactionsItem
	_, err := j.txn.Query(&items, `
		SELECT DISTINCT
			"bank_account"."account_id",
			"bank_account"."link_id"
		FROM "transactions" AS "transaction"
		INNER JOIN "bank_accounts" AS "bank_account" ON "bank_account"."account_id" = "transaction"."account_id" AND "bank_account"."bank_account_id" = "transaction"."bank_account_id"
		INNER JOIN "links" AS "link" ON "link"."account_id" = "bank_account"."account_id" AND "link"."link_id" = "bank_account"."link_id"
		WHERE
			"link"."link_type" = ? AND
			"transaction"."is_pending" = true
	`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve bank accounts with pending transactions")
	}

	return items, nil
}

func (r *repositoryBase) GetJob(jobId string) (models.Job, error) {
	var result models.Job
	err := r.txn.Model(&result).
		Where(`"job"."account_id" = ?`, r.AccountId()).
		Where(`"job"."job_id" = ?`, jobId).
		Limit(1).
		Select(&result)

	return result, errors.Wrap(err, "failed to retrieve job")
}
