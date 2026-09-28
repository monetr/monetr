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

		{ // The diff only knows about clusters that a transaction points at. A
			// cluster nothing points at would never get deleted, and would block the
			// same signature + centroid from being inserted under a new ID.
			storedClusterIds, err := repo.GetTransactionClusterIds(
				ctx,
				args.BankAccountId,
			)
			if err != nil {
				return errors.Wrap(err, "failed to read transaction cluster ids")
			}
			referenced := make(
				map[models.ID[models.TransactionCluster]]struct{},
				len(existingMembers),
			)
			for _, m := range existingMembers {
				referenced[*m.TransactionClusterId] = struct{}{}
			}
			for _, clusterId := range storedClusterIds {
				if _, ok := referenced[clusterId]; !ok {
					diff.DeleteClusterIds = append(diff.DeleteClusterIds, clusterId)
				}
			}
		}

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
				"transactionClusterId", *item.TransactionClusterId,
			)
		}

		return nil
	})
}
