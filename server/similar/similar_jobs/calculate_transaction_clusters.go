package similar_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/recurring/recurring_jobs"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/similar"
	"github.com/pkg/errors"
)

type CalculateTransactionClustersArguments struct {
	AccountId     models.ID[models.Account]     `json:"accountId"`
	BankAccountId models.ID[models.BankAccount] `json:"bankAccountId"`
}

func CalculateTransactionClusters(ctx queue.Context, args CalculateTransactionClustersArguments) error {
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

		clustering := similar.NewSimilarTransactions_TFIDF_DBSCAN(log)

		{ // Read the entire bank accounts transaction data into the dataset.
			transactions, err := repo.GetTransactionsForSimilarity(
				ctx,
				args.BankAccountId,
			)
			if err != nil {
				return errors.Wrap(err, "failed to read transactions for clustering")
			}
			for i := range transactions {
				clustering.AddTransaction(&transactions[i])
			}
		}

		result := clustering.DetectSimilarTransactions(ctx)

		log.InfoContext(ctx, "similar transaction clusters detected", "clusters", len(result))

		existingMembers, err := repo.GetClusteredTransactions(
			ctx,
			args.BankAccountId,
		)
		if err != nil {
			return err
		}

		diff := DiffClusterMembers(
			ctx,
			existingMembers,
			result,
			args.AccountId,
			args.BankAccountId,
		)

		log.InfoContext(ctx, "cluster membership diff calculated",
			"upsertClusters", len(diff.UpsertClusters),
			"deleteClusters", len(diff.DeleteClusterIds),
			"insertMembers", len(diff.InsertMembers),
			"updateMembers", len(diff.UpdateMembers),
			"deleteMembers", len(diff.DeleteMemberIds),
		)

		// Order matters here because of the foreign keys. Clusters need to exist
		// before transactions can point at them, and old clusters go last. If
		// anything still points at an old cluster the FK will just null it out
		if err := repo.UpsertTransactionClusters(
			ctx,
			args.BankAccountId,
			diff.UpsertClusters,
		); err != nil {
			return errors.Wrap(err, "failed to upsert transaction clusters")
		}

		changedMembers := make(
			[]models.Transaction,
			0,
			len(diff.InsertMembers)+len(diff.UpdateMembers)+len(diff.DeleteMemberIds),
		)
		changedMembers = append(changedMembers, diff.InsertMembers...)
		changedMembers = append(changedMembers, diff.UpdateMembers...)
		for _, transactionId := range diff.DeleteMemberIds {
			changedMembers = append(changedMembers, models.Transaction{
				TransactionId:        transactionId,
				TransactionClusterId: nil,
			})
		}

		if err := repo.UpdateTransactionClusterIds(
			ctx,
			args.BankAccountId,
			changedMembers,
		); err != nil {
			return errors.Wrap(err, "failed to update transaction cluster ids")
		}

		if err := repo.DeleteTransactionClusters(
			ctx,
			args.BankAccountId,
			diff.DeleteClusterIds,
		); err != nil {
			return errors.Wrap(err, "failed to delete obsolete transaction clusters")
		}

		log.InfoContext(ctx, "finished updating transaction clusters",
			"upsertClusters", len(diff.UpsertClusters),
			"deleteClusters", len(diff.DeleteClusterIds),
			"insertMembers", len(diff.InsertMembers),
			"updateMembers", len(diff.UpdateMembers),
			"deleteMembers", len(diff.DeleteMemberIds),
		)

		// Every cluster that still exists gets its recurring transactions
		// recalculated, the clusters that were deleted take their recurring
		// transactions with them. This is enqueued inside the transaction so the
		// jobs only run once the clusters are committed.
		if err := queue.BulkEnqueue(
			ctx,
			ctx.Enqueuer(),
			recurring_jobs.CalculateRecurringTransactions,
			myownsanity.Map(
				diff.UpsertClusters,
				func(item models.TransactionCluster) recurring_jobs.CalculateRecurringTransactionsArguments {
					return recurring_jobs.CalculateRecurringTransactionsArguments{
						AccountId:            args.AccountId,
						BankAccountId:        args.BankAccountId,
						TransactionClusterId: item.TransactionClusterId,
					}
				},
			),
		); err != nil {
			return errors.Wrap(err, "failed to enqueue recurring transaction calculations")
		}

		log.InfoContext(ctx, "enqueued recurring transaction calculations",
			"clusters", len(diff.UpsertClusters),
		)

		for _, item := range diff.InsertMembers {
			log.DebugContext(
				ctx,
				"placeholder, triggering similar transaction rules for transaction",
				"transactionId", item.TransactionId,
				"transactionClusterId", *item.TransactionClusterId,
			)
		}

		return nil
	})
}
