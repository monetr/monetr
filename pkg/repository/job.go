package repository

import (
	"context"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

type JobRepository interface {
	GetBankAccountsToSync() ([]models.BankAccount, error)
	GetBankAccountsWithPendingTransactions() ([]CheckingPendingTransactionsItem, error)
	GetFundingSchedulesToProcess() ([]ProcessFundingSchedulesItem, error)
	GetPlaidLinksByAccount(ctx context.Context) ([]PlaidLinksForAccount, error)
	GetLinksForExpiredAccounts(ctx context.Context) ([]models.Link, error)
	GetBankAccountsWithStaleSpending(ctx context.Context) ([]BankAccountWithStaleSpendingItem, error)
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

type PlaidLinksForAccount struct {
	tableName string `pg:"links"`

	AccountId uint64   `pg:"account_id"`
	LinkIds   []uint64 `pg:"link_ids,type:bigint[]"`
}

type BankAccountWithStaleSpendingItem struct {
	AccountId     uint64 `pg:"account_id"`
	BankAccountId uint64 `pg:"bank_account_id"`
}

type jobRepository struct {
	txn pg.DBI
}

func NewJobRepository(db pg.DBI) JobRepository {
	return &jobRepository{
		txn: db,
	}
}

func (j *jobRepository) GetPlaidLinksByAccount(ctx context.Context) ([]PlaidLinksForAccount, error) {
	links := make([]PlaidLinksForAccount, 0)
	err := j.txn.ModelContext(ctx, &links).
		ColumnExpr(`"account_id"`).
		ColumnExpr(`array_agg("link_id") "link_ids"`).
		Where(`"link_type" = ?`, models.PlaidLinkType).
		Where(`"plaid_link_id" IS NOT NULL`).
		Group("account_id").
		Select(&links)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plaid links by account")
	}

	return links, nil
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
	query := `
		SELECT
			"funding_schedules"."account_id",
			"funding_schedules"."bank_account_id",
			array_agg("funding_schedules"."funding_schedule_id") AS "funding_schedule_ids"
		FROM "funding_schedules"
		WHERE "funding_schedules"."next_occurrence" < (now() AT TIME ZONE 'UTC')
		GROUP BY "funding_schedules"."account_id", "funding_schedules"."bank_account_id"
	`

	query = strings.NewReplacer(
		"\n", "",
		"\t", "",
		"  ", " ",
	).Replace(query)

	_, err := j.txn.Query(&items, query)
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

func (j *jobRepository) GetLinksForExpiredAccounts(ctx context.Context) ([]models.Link, error) {
	span := sentry.StartSpan(ctx, "GetLinksForExpiredAccounts")
	defer span.Finish()

	// Links should be seen as expired if the account subscription is not active for 90 days.
	expirationCutoff := time.Now().UTC().Add(-90 * 24 * time.Hour)

	var result []models.Link
	err := j.txn.ModelContext(span.Context(), &result).
		Join(`INNER JOIN "accounts" AS "account"`).
		JoinOn(`"account"."account_id" = "link"."account_id"`).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Where(`"account"."subscription_active_until" IS NOT NULL`).
		Where(`"account"."subscription_active_until" < ?`, expirationCutoff).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Plaid links for expired accounts")
	}

	return result, nil
}

// GetBankAccountsWithStaleSpending will return all of the bank accounts globally that have a non-paused spending object
// with a next recurrence that is in the past. This is used to find spending objects that need to be updated as they
// have not been spent from for at least once cycle.
func (j *jobRepository) GetBankAccountsWithStaleSpending(ctx context.Context) ([]BankAccountWithStaleSpendingItem, error) {
	span := sentry.StartSpan(ctx, "GetBankAccountsWithStaleSpending")
	defer span.Finish()

	var result []BankAccountWithStaleSpendingItem
	err := j.txn.ModelContext(span.Context(), &models.BankAccount{}).
		ColumnExpr(`"bank_account"."account_id"`).
		ColumnExpr(`"bank_account"."bank_account_id"`).
		Join(`INNER JOIN "spending" AS "spending"`).
		JoinOn(`"spending"."account_id" = "bank_account"."account_id" AND "spending"."bank_account_id" = "bank_account"."bank_account_id"`).
		Where(`"spending"."next_recurrence" < ?`, time.Now()).
		Where(`"spending"."is_paused" = ?`, false).
		GroupExpr(`"bank_account"."account_id"`).
		GroupExpr(`"bank_account"."bank_account_id"`).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve bank accounts with stale spending objects")
	}

	return result, err
}
