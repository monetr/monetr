package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

type JobRepository interface {
	CreateInstitutions(ctx context.Context, institutions []*models.Institution) error
	GetBankAccountsToSync() ([]models.BankAccount, error)
	GetBankAccountsWithPendingTransactions() ([]CheckingPendingTransactionsItem, error)
	GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error)
	GetInstitutionsByPlaidID(ctx context.Context, plaidIds []string) (map[string]models.Institution, error)
	UpdateInstitutions(ctx context.Context, institutions []*models.Institution) error
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

func (j *jobRepository) GetInstitutionsByPlaidID(ctx context.Context, plaidIds []string) (map[string]models.Institution, error) {
	span := sentry.StartSpan(ctx, "GetInstitutionByPlaidID")
	defer span.Finish()

	institutions := make([]models.Institution, 0)
	err := j.txn.ModelContext(span.Context(), &institutions).
		WhereIn(`"institution"."plaid_institution_id" IN (?)`, plaidIds).
		Select(&institutions)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve institutions by plaid Id")
	}
	span.Status = sentry.SpanStatusOK

	byPlaidId := map[string]models.Institution{}
	for _, institution := range institutions {
		if institution.PlaidInstitutionId == nil {
			continue
		}

		byPlaidId[*institution.PlaidInstitutionId] = institution
	}

	return byPlaidId, nil
}

func (j *jobRepository) UpdateInstitutions(ctx context.Context, institutions []*models.Institution) error {
	span := sentry.StartSpan(ctx, "UpdateInstitutions")
	defer span.Finish()

	result, err := j.txn.ModelContext(span.Context(), &institutions).
		WherePK().
		Update(&institutions)
	if err != nil {
		return errors.Wrap(err, "failed to update institutions")
	}

	if affected := result.RowsAffected(); affected != len(institutions) {
		return errors.Errorf("unexpected institutions updated, expected: %d updated: %d", len(institutions), affected)
	}

	return nil
}

func (j *jobRepository) CreateInstitutions(ctx context.Context, institutions []*models.Institution) error {
	span := sentry.StartSpan(ctx, "CreateInstitutions")
	defer span.Finish()

	_, err := j.txn.ModelContext(span.Context(), &institutions).
		Insert(&institutions)

	return errors.Wrap(err, "failed to create institutions")
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
	`, models.PlaidLinkType)
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
