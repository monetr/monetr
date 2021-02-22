package repository

import (
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

type JobRepository interface {
	GetBankAccountsToSync() ([]models.BankAccount, error)
	GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error)
}

type ProcessFundingSchedulesItem struct {
	AccountId          uint64   `pg:"account_id"`
	FundingScheduleIds []uint64 `pg:"funding_schedule_ids,type:bigint[]"`
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
			"accounts"."account_id",
			array_agg("funding_schedules"."funding_schedule_id") AS "funding_schedule_ids"
		FROM "accounts"
		INNER JOIN "funding_schedules" ON "funding_schedules"."account_id" = "accounts"."account_id"
		WHERE "funding_schedules"."next_occurrence" < (now() AT TIME ZONE 'UTC')
	`)
	if err != nil {
		// TODO (elliotcourant) Can pg.NoRows return here? If it can this error is useless.
		return nil, errors.Wrap(err, "failed to retrieve accounts and their funding schedules")
	}

	return items, nil
}
