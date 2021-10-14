package repository

import (
	"context"
	"time"

	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type JobRepository interface {
	GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error)
}

type ProcessFundingSchedulesItem struct {
	bun.BaseModel `bun:"funding_schedules"`

	AccountId          uint64   `pg:"account_id"`
	BankAccountId      uint64   `pg:"bank_account_id"`
	FundingScheduleIds []uint64 `pg:"funding_schedule_ids,type:bigint[]"`
}

type CheckingPendingTransactionsItem struct {
	AccountId uint64 `pg:"account_id"`
	LinkId    uint64 `pg:"link_id"`
}

type jobRepository struct {
	db bun.IDB
}

func NewJobRepository(db bun.IDB) JobRepository {
	return &jobRepository{
		db: db,
	}
}

func (j *jobRepository) GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error) {
	var items []ProcessFundingSchedulesItem
	err := j.db.NewSelect().
		Model(&items).
		Where(`funding_schedules.next_occurrence < ?`, time.Now().UTC()).
		Group(`account_id`, `bank_account_id`).
		Scan(context.Background(), &items)
	if err != nil {
		// TODO (elliotcourant) Can pg.NoRows return here? If it can this error is useless.
		return nil, errors.Wrap(err, "failed to retrieve accounts and their funding schedules")
	}

	return items, nil
}

func (r *repositoryBase) GetJob(jobId string) (models.Job, error) {
	var result models.Job
	err := r.db.NewSelect().
		Model(&result).
		Where(`job.account_id = ?`, r.AccountId()).
		Where(`job.job_id = ?`, jobId).
		Limit(1).
		Scan(context.Background(), &result)

	return result, errors.Wrap(err, "failed to retrieve job")
}
