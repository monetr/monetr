package recurring_jobs

import (
	"log/slog"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/recurring"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type CalculateRecurringTransactionsArguments struct {
	AccountId     models.ID[models.Account]     `json:"accountId"`
	BankAccountId models.ID[models.BankAccount] `json:"bankAccountId"`
}

// CalculateRecurringTransactions detects the recurring transactions for every
// transaction cluster in the bank account. Each cluster is calculated in its own
// transaction, if one fails then it is logged and skipped and the rest of the
// clusters are still calculated. The job only fails when every cluster does,
// since that is probably something bigger than a single bad cluster.
func CalculateRecurringTransactions(
	ctx queue.Context,
	args CalculateRecurringTransactionsArguments,
) error {
	crumbs.IncludeUserInScope(ctx, args.AccountId)
	log := ctx.Log().With(
		"accountId", args.AccountId,
		"bankAccountId", args.BankAccountId,
	)

	repo := repository.NewRepositoryFromSession(
		ctx.Clock(),
		"user_system",
		args.AccountId,
		ctx.DB(),
		log,
	)

	account, err := repo.GetAccount(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve account for job")
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		log.WarnContext(ctx, "failed to get account's time zone, defaulting to UTC", "err", err)
		timezone = time.UTC
	}

	clusterIds, err := repo.GetTransactionClusterIds(ctx, args.BankAccountId)
	if err != nil {
		return errors.Wrap(err, "failed to read transaction clusters")
	}

	var lastErr error
	failed := 0
	for _, transactionClusterId := range clusterIds {
		// Don't keep chewing through clusters if the job is being stopped, every
		// one of them would just fail anyway.
		if err := ctx.Err(); err != nil {
			return errors.Wrap(err, "stopped calculating recurring transactions")
		}

		err := ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
			return calculateRecurringTransactionsForCluster(
				ctx,
				log,
				args.AccountId,
				args.BankAccountId,
				transactionClusterId,
				timezone,
			)
		})
		if err != nil {
			failed++
			lastErr = err
			log.ErrorContext(ctx, "failed to calculate recurring transactions for cluster, skipping it",
				"transactionClusterId", transactionClusterId,
				"err", err,
			)
			crumbs.ReportError(
				ctx,
				err,
				"Failed to calculate recurring transactions for cluster",
				"job",
				map[string]any{
					"transactionClusterId": transactionClusterId,
				},
			)
			continue
		}
	}

	log.InfoContext(ctx, "finished calculating recurring transactions",
		"clusters", len(clusterIds),
		"failed", failed,
	)

	if failed > 0 && failed == len(clusterIds) {
		return errors.Wrap(lastErr, "failed to calculate recurring transactions for every cluster")
	}

	return nil
}

func calculateRecurringTransactionsForCluster(
	ctx queue.Context,
	log *slog.Logger,
	accountId models.ID[models.Account],
	bankAccountId models.ID[models.BankAccount],
	transactionClusterId models.ID[models.TransactionCluster],
	timezone *time.Location,
) error {
	log = log.With("transactionClusterId", transactionClusterId)
	repo := repository.NewRepositoryFromSession(
		ctx.Clock(),
		"user_system",
		accountId,
		ctx.DB(),
		log,
	)

	// TODO This will need to change once I support merging clusters.
	transactions, err := repo.GetTransactionsByCluster(
		ctx,
		bankAccountId,
		transactionClusterId,
		1000, // Something high for this?
		0,
	)
	if err != nil {
		return errors.Wrap(err, "failed to read transactions in cluster")
	}

	results, err := recurring.DetectRecurringTransactions(
		ctx,
		ctx.Clock(),
		timezone,
		transactions,
	)
	if err != nil {
		return errors.Wrap(err, "failed to detect recurring transactions")
	}

	existing, err := repo.GetTransactionRecurringByCluster(
		ctx,
		bankAccountId,
		transactionClusterId,
	)
	if err != nil {
		return err
	}

	diff := DiffTransactionRecurring(
		ctx,
		existing,
		results,
		transactions,
		ctx.Clock().Now(),
		timezone,
		bankAccountId,
		transactionClusterId,
	)

	// Order matters here because of the foreign keys. Existing recurring
	// transactions are updated in place before anything new is created, they
	// need to exist before transactions can point at them, and the ones that
	// don't recur anymore are cleaned up last.
	if err := repo.UpsertTransactionRecurring(
		ctx,
		bankAccountId,
		diff.UpsertRecurring,
	); err != nil {
		return errors.Wrap(err, "failed to upsert recurring transactions")
	}

	if err := repo.UpdateTransactionRecurringIds(
		ctx,
		bankAccountId,
		diff.UpdateMembers,
	); err != nil {
		return errors.Wrap(err, "failed to update transaction recurring ids")
	}

	if err := repo.DeleteTransactionRecurring(
		ctx,
		bankAccountId,
		diff.DeleteRecurringIds,
	); err != nil {
		return errors.Wrap(err, "failed to delete obsolete recurring transactions")
	}

	// This runs for every cluster in the bank account, so keep it at debug.
	log.DebugContext(ctx, "finished updating recurring transactions for cluster",
		"upsertRecurring", len(diff.UpsertRecurring),
		"deleteRecurring", len(diff.DeleteRecurringIds),
		"updateMembers", len(diff.UpdateMembers),
	)

	return nil
}
