package repository

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type JobRepository interface {
	GetFundingSchedulesToProcess(ctx context.Context) ([]ProcessFundingSchedulesItem, error)
	GetLinksForExpiredAccounts(ctx context.Context) ([]Link, error)
	GetBankAccountsWithStaleSpending(ctx context.Context) ([]BankAccountWithStaleSpendingItem, error)
	GetAccountsWithTooManyFiles(ctx context.Context) ([]AccountWithTooManyFiles, error)
	GetStaleAccounts(ctx context.Context) ([]Account, error)
}

type ProcessFundingSchedulesItem struct {
	AccountId          ID[Account]           `pg:"account_id"`
	BankAccountId      ID[BankAccount]       `pg:"bank_account_id"`
	FundingScheduleIds []ID[FundingSchedule] `pg:"funding_schedule_ids,type:varchar(32)[]"`
}

type CheckingPendingTransactionsItem struct {
	AccountId ID[Account] `pg:"account_id"`
	LinkId    ID[Link]    `pg:"link_id"`
}

type PlaidLinksForAccount struct {
	tableName string `pg:"links"`

	AccountId ID[Account] `pg:"account_id"`
	LinkIds   []ID[Link]  `pg:"link_ids,type:varchar(32)[]"`
}

type BankAccountWithStaleSpendingItem struct {
	AccountId     ID[Account]     `pg:"account_id"`
	BankAccountId ID[BankAccount] `pg:"bank_account_id"`
}

type jobRepository struct {
	txn   pg.DBI
	clock clock.Clock
}

func NewJobRepository(db pg.DBI, clock clock.Clock) JobRepository {
	return &jobRepository{
		txn:   db,
		clock: clock,
	}
}

func (j *jobRepository) GetFundingSchedulesToProcess(ctx context.Context) ([]ProcessFundingSchedulesItem, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var items []ProcessFundingSchedulesItem
	// TODO Exclude deleted bank accounts from processing
	_, err := j.txn.QueryContext(
		span.Context(),
		&items,
		`
		SELECT
			"funding_schedules"."account_id",
			"funding_schedules"."bank_account_id",
			array_agg("funding_schedules"."funding_schedule_id") AS "funding_schedule_ids"
		FROM "funding_schedules"
		WHERE "funding_schedules"."next_recurrence" < ?
		GROUP BY "funding_schedules"."account_id", "funding_schedules"."bank_account_id"
		`,
		j.clock.Now(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve accounts and their funding schedules")
	}

	return items, nil
}

func (r *repositoryBase) GetJob(jobId string) (Job, error) {
	var result Job
	err := r.txn.Model(&result).
		Where(`"job"."account_id" = ?`, r.AccountId()).
		Where(`"job"."job_id" = ?`, jobId).
		Limit(1).
		Select(&result)

	return result, errors.Wrap(err, "failed to retrieve job")
}

func (j *jobRepository) GetLinksForExpiredAccounts(ctx context.Context) ([]Link, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Links should be seen as expired if the account subscription is not active for 90 days.
	expirationCutoff := j.clock.Now().Add(-90 * 24 * time.Hour).UTC()

	var result []Link
	err := j.txn.ModelContext(span.Context(), &result).
		Join(`INNER JOIN "accounts" AS "account"`).
		JoinOn(`"account"."account_id" = "link"."account_id"`).
		Where(`"link"."link_type" = ?`, PlaidLinkType).
		Where(`GREATEST("account"."subscription_active_until", "account"."trial_ends_at") < ?`, expirationCutoff).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Plaid links for expired accounts")
	}

	return result, nil
}

func (j *jobRepository) GetStaleAccounts(ctx context.Context) ([]Account, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Accounts that have been expired for at least 100 days should be considered
	// stale.
	staleCutoff := j.clock.Now().Add(-100 * 24 * time.Hour).UTC()

	var result []Account
	err := j.txn.ModelContext(span.Context(), &result).
		Where(`GREATEST("account"."subscription_active_until", "account"."trial_ends_at") < ?`, staleCutoff).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve stale accounts")
	}

	return result, nil
}

// GetBankAccountsWithStaleSpending will return all of the bank accounts globally that have a non-paused spending object
// with a next recurrence that is in the past. This is used to find spending objects that need to be updated as they
// have not been spent from for at least once cycle.
func (j *jobRepository) GetBankAccountsWithStaleSpending(ctx context.Context) ([]BankAccountWithStaleSpendingItem, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result []BankAccountWithStaleSpendingItem
	err := j.txn.ModelContext(span.Context(), &BankAccount{}).
		ColumnExpr(`"bank_account"."account_id"`).
		ColumnExpr(`"bank_account"."bank_account_id"`).
		Join(`INNER JOIN "spending" AS "spending"`).
		JoinOn(`"spending"."account_id" = "bank_account"."account_id" AND "spending"."bank_account_id" = "bank_account"."bank_account_id"`).
		Where(`"spending"."next_recurrence" < ?`, j.clock.Now()).
		Where(`"spending"."is_paused" = ?`, false).
		GroupExpr(`"bank_account"."account_id"`).
		GroupExpr(`"bank_account"."bank_account_id"`).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve bank accounts with stale spending objects")
	}

	return result, err
}

type AccountWithTooManyFiles struct {
	tableName string `pg:"files"`

	AccountId ID[Account] `pg:"account_id"`
	Count     int64       `pg:"count"`
}

func (j *jobRepository) GetAccountsWithTooManyFiles(ctx context.Context) ([]AccountWithTooManyFiles, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result []AccountWithTooManyFiles
	err := j.txn.ModelContext(span.Context(), &result).
		ColumnExpr(`"account_id"`).
		ColumnExpr(`COUNT("file_id") AS "count"`).
		GroupExpr(`"account_id"`).
		Having(`COUNT("file_id") > ?`, 10).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find accounts with too many files")
	}

	return result, nil
}
