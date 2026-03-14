package similar

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/recurring"
	"github.com/monetr/monetr/server/repository"
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

		clustering := recurring.NewSimilarTransactions_TFIDF_DBSCAN(log)

		limit := 2000
		offset := 0
		for {
			log.Log(ctx, logging.LevelTrace, "requesting next batch of transactions",
				"limit", limit,
				"offset", offset,
			)
			transactions, err := repo.GetTransactions(ctx, args.BankAccountId, limit, offset)
			if err != nil {
				return errors.Wrap(err, "failed to read transactions for clustering")
			}

			for i := range transactions {
				clustering.AddTransaction(&transactions[i])
			}

			if len(transactions) < limit {
				log.Log(ctx, logging.LevelTrace, "reached end of transactions",
					"count", len(transactions),
				)
				break
			}

			offset += len(transactions)
		}

		result := clustering.DetectSimilarTransactions(ctx)

		if len(result) == 0 {
			log.InfoContext(ctx, "no similar transactions detected, nothing to persist")
			return nil
		}

		log.InfoContext(ctx, "similar transaction clusters detected", "clusters", len(result))

		if _, err := repo.WriteTransactionClusters(ctx, args.BankAccountId, result); err != nil {
			return errors.Wrap(err, "failed to persist the calculated transaction clusters")
		}

		return nil
	})
}
