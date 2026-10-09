package recurring_jobs

import (
	"log/slog"
	"slices"
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

// CalculateRecurringTransactions will detect the recurring transactions for
// every transaction cluster in the bank account, all in one transaction. A
// cluster whose detection fails is logged and skipped, since retrying won't
// change the outcome. Anything that fails in the database fails the whole job,
// which rolls everything back and lets the queue retry it.
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

		account, err := repo.GetAccount(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve account for job")
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.WarnContext(
				ctx,
				"failed to get account's time zone, defaulting to UTC",
				"err", err,
			)
			timezone = time.UTC
		}

		clusterIds, err := repo.GetTransactionClusterIds(ctx, args.BankAccountId)
		if err != nil {
			return errors.Wrap(err, "failed to read transaction clusters")
		}

		autoAssignTransactionIds := make([]models.ID[models.Transaction], 0)
		for _, transactionClusterId := range clusterIds {
			transactionIds, err := calculateRecurringTransactionsForCluster(
				ctx,
				log,
				repo,
				args.BankAccountId,
				transactionClusterId,
				timezone,
			)
			if err != nil {
				return err
			}
			autoAssignTransactionIds = append(autoAssignTransactionIds, transactionIds...)
		}

		log.InfoContext(ctx, "finished calculating recurring transactions",
			"clusters", len(clusterIds),
			"autoAssignTransactions", len(autoAssignTransactionIds),
		)

		if len(autoAssignTransactionIds) > 0 {
			if err := queue.Enqueue(
				ctx,
				ctx.Enqueuer(),
				AutoAssignRecurringTransactions,
				AutoAssignRecurringTransactionsArguments{
					AccountId:      args.AccountId,
					BankAccountId:  args.BankAccountId,
					TransactionIds: autoAssignTransactionIds,
				},
			); err != nil {
				return errors.Wrap(err, "failed to enqueue auto assigning recurring transactions")
			}
		}

		// This is enqueued inside the transaction so the job only runs once the
		// recurring transactions are committed.
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			MatchRecurringTransactionsToSpending,
			MatchRecurringTransactionsToSpendingArguments{
				AccountId:     args.AccountId,
				BankAccountId: args.BankAccountId,
			},
		); err != nil {
			return errors.Wrap(err, "failed to enqueue matching recurring transactions to spending")
		}

		return nil
	})
}

func calculateRecurringTransactionsForCluster(
	ctx queue.Context,
	log *slog.Logger,
	repo repository.Repository,
	bankAccountId models.ID[models.BankAccount],
	transactionClusterId models.ID[models.TransactionCluster],
	timezone *time.Location,
) ([]models.ID[models.Transaction], error) {
	log = log.With("transactionClusterId", transactionClusterId)

	// TODO This will need to change once I support merging clusters.
	transactions, err := repo.GetTransactionsByCluster(
		ctx,
		bankAccountId,
		transactionClusterId,
		1000, // Something high for this?
		0,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read transactions in cluster")
	}

	results, err := recurring.DetectRecurringTransactions(
		ctx,
		ctx.Clock(),
		timezone,
		transactions,
	)
	if err != nil {
		// Nothing has been written for this cluster yet, so it can just be skipped.
		// Its existing recurring transactions are left as they are.
		log.ErrorContext(
			ctx,
			"failed to detect recurring transactions for cluster, skipping it",
			"err", err,
		)
		crumbs.ReportError(
			ctx,
			err,
			"Failed to detect recurring transactions for cluster",
			"job",
			map[string]any{
				"transactionClusterId": transactionClusterId,
			},
		)
		return nil, nil
	}

	existing, err := repo.GetTransactionRecurringByCluster(
		ctx,
		bankAccountId,
		transactionClusterId,
	)
	if err != nil {
		return nil, err
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
	// transactions are updated in place before anything new is created, they need
	// to exist before transactions can point at them, and the ones that don't
	// recur anymore are cleaned up last.
	if err := repo.UpsertTransactionRecurring(
		ctx,
		bankAccountId,
		diff.UpsertRecurring,
	); err != nil {
		return nil, errors.Wrap(err, "failed to upsert recurring transactions")
	}

	if err := repo.UpdateTransactionRecurringIds(
		ctx,
		bankAccountId,
		slices.Concat(diff.InsertMembers, diff.UpdateMembers),
	); err != nil {
		return nil, errors.Wrap(err, "failed to update transaction recurring ids")
	}

	if err := repo.DeleteTransactionRecurring(
		ctx,
		bankAccountId,
		diff.DeleteRecurringIds,
	); err != nil {
		return nil, errors.Wrap(err, "failed to delete obsolete recurring transactions")
	}

	// This runs for every cluster in the bank account, so keep it at debug.
	log.DebugContext(ctx, "finished updating recurring transactions for cluster",
		"upsertRecurring", len(diff.UpsertRecurring),
		"deleteRecurring", len(diff.DeleteRecurringIds),
		"insertMembers", len(diff.InsertMembers),
		"updateMembers", len(diff.UpdateMembers),
	)

	// Only transactions that were just added to a recurring transaction the user
	// wants auto assigned need to be looked at. New recurring transactions can't
	// have auto assign turned on yet, so only the existing ones matter.
	autoAssign := make(map[models.ID[models.TransactionRecurring]]bool, len(existing))
	for _, item := range existing {
		if item.AutoAssign && item.SpendingId != nil && item.DeletedAt == nil {
			autoAssign[item.TransactionRecurringId] = true
		}
	}

	autoAssignTransactionIds := make([]models.ID[models.Transaction], 0)
	for _, member := range diff.InsertMembers {
		if autoAssign[*member.TransactionRecurringId] {
			autoAssignTransactionIds = append(autoAssignTransactionIds, member.TransactionId)
		}
	}

	return autoAssignTransactionIds, nil
}
