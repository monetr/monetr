package recurring_jobs

import (
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/recurring"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type CalculateRecurringTransactionsArguments struct {
	AccountId            models.ID[models.Account]            `json:"accountId"`
	BankAccountId        models.ID[models.BankAccount]        `json:"bankAccountId"`
	TransactionClusterId models.ID[models.TransactionCluster] `json:"transactionClusterId"`
}

func CalculateRecurringTransactions(
	ctx queue.Context,
	args CalculateRecurringTransactionsArguments,
) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
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

		// TODO This will need to change once I support merging clusters.
		transactions, err := repo.GetTransactionsByCluster(
			ctx,
			args.BankAccountId,
			args.TransactionClusterId,
			1000, // Something high for this?
			0,
		)
		if err != nil {
			return errors.Wrap(err, "failed to read transactions in cluster")
		}

		account, err := repo.GetAccount(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve account for job")
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.WarnContext(ctx, "failed to get account's time zone, defaulting to UTC", "err", err)
			timezone = time.UTC
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
			args.BankAccountId,
			args.TransactionClusterId,
		)
		if err != nil {
			return err
		}

		diff := DiffTransactionRecurring(
			ctx,
			existing,
			results,
			ctx.Clock().Now(),
			args.BankAccountId,
			args.TransactionClusterId,
		)

		log.InfoContext(ctx, "recurring transaction diff calculated",
			"transactionClusterId", args.TransactionClusterId,
			"upsertRecurring", len(diff.UpsertRecurring),
			"deleteRecurring", len(diff.DeleteRecurringIds),
		)

		// Existing recurring transactions are updated in place before anything new
		// is created, and the ones that don't recur anymore are cleaned up last.
		if err := repo.UpsertTransactionRecurring(
			ctx,
			args.BankAccountId,
			diff.UpsertRecurring,
		); err != nil {
			return errors.Wrap(err, "failed to upsert recurring transactions")
		}

		if err := repo.DeleteTransactionRecurring(
			ctx,
			args.BankAccountId,
			diff.DeleteRecurringIds,
		); err != nil {
			return errors.Wrap(err, "failed to delete obsolete recurring transactions")
		}

		log.InfoContext(ctx, "finished updating recurring transactions",
			"transactionClusterId", args.TransactionClusterId,
			"upsertRecurring", len(diff.UpsertRecurring),
			"deleteRecurring", len(diff.DeleteRecurringIds),
		)

		return nil
	})
}
