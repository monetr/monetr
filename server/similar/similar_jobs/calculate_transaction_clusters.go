package similar_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
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

		existingMembers, err := repo.GetTransactionClusterMembersByBankAccount(
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

		// Execute the diff in FK-safe order. Clusters must exist before members
		// can reference them, and members must be moved away from obsolete
		// clusters before those clusters can be deleted.
		if err := repo.UpsertTransactionClusters(
			ctx,
			args.BankAccountId,
			diff.UpsertClusters,
		); err != nil {
			return errors.Wrap(err, "failed to upsert transaction clusters")
		}

		if err := repo.UpsertTransactionClusterMembers(
			ctx,
			append(diff.InsertMembers, diff.UpdateMembers...),
		); err != nil {
			return errors.Wrap(err, "failed to upsert transaction cluster members")
		}

		if err := repo.DeleteTransactionClusterMembers(
			ctx,
			args.BankAccountId,
			diff.DeleteMemberIds,
		); err != nil {
			return errors.Wrap(err, "failed to delete removed cluster members")
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

		for _, item := range diff.UpsertClusters {
			log.DebugContext(
				ctx,
				"placeholder, triggering recurring transaction detection on transaction cluster",
				"transactionClusterId", item.TransactionClusterId,
			)
		}

		for _, item := range diff.InsertMembers {
			log.DebugContext(
				ctx,
				"placeholder, triggering similar transaction rules for transaction",
				"transactionId", item.TransactionId,
				"transactionClusterId", item.TransactionClusterId,
			)
		}

		return nil
	})
}
