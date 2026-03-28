package similar_jobs

import (
	"slices"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
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

		if len(result) == 0 {
			log.InfoContext(ctx, "no similar transactions detected, nothing to persist")
			return nil
		}

		log.InfoContext(ctx, "similar transaction clusters detected", "clusters", len(result))

		// Go I really wish I could write this part in Clojure, I hate working with
		// data like this in Golang its so fucking ugly :(
		existingMembers, err := repo.GetTransactionClusterMembersByBankAccount(
			ctx,
			args.BankAccountId,
		)
		if err != nil {
			return err
		}

		// Basically what we want to do here is group all of the existing
		// transaction cluster members by their transaction cluster. Then we want to
		// compare the membership to the newly calculated clusters. So first things
		// first, group by cluster ID.
		existingGroups := myownsanity.GroupByMapV(
			existingMembers,
			func(
				item models.TransactionClusterMember,
			) models.ID[models.TransactionCluster] {
				return item.TransactionClusterId
			},
			func(
				item models.TransactionClusterMember,
			) models.ID[models.Transaction] {
				return item.TransactionId
			},
		)
		type Overlap struct {
			Existing models.ID[models.TransactionCluster]
			Count    int
		}
		type Joined struct {
			New   models.TransactionCluster
			Items []Overlap
		}

		candiates := map[models.ID[models.TransactionCluster]]struct{}{}

		// Don't join anything yet, but determine which candiates of the existing
		// transaction clusters would even be eligible for merging.
		for _, a := range result {
			for existingClusterId, items := range existingGroups {
				if _, ok := candiates[existingClusterId]; ok {
					continue
				}
				intersect := myownsanity.Intersection(a.Members, items)
				if len(intersect) == 0 {
					continue
				}
				candiates[existingClusterId] = struct{}{}
			}
		}

		type Score struct {
			ResultIndex int
			Score       float64
		}
		type Pair struct {
			New      models.TransactionCluster
			Existing models.ID[models.TransactionCluster]
		}
		pairs := make([]Pair, 0, len(candiates))
		for existingClusterId := range candiates {
			score := make([]Score, len(result))
			for i, a := range result {
				intersect := myownsanity.Intersection(
					a.Members,
					existingGroups[existingClusterId],
				)
				union := myownsanity.Union(
					a.Members,
					existingGroups[existingClusterId],
				)
				score[i] = Score{
					ResultIndex: i,
					// Take the number of intersecting items and divide it by the number
					// of unique items in both sets. This should give us the overlap
					// percentage.
					Score: float64(len(intersect)) / float64(len(union)),
				}
			}
			slices.SortStableFunc(score, func(a Score, b Score) int {
				if a.Score < b.Score {
					return -1
				} else if a.Score > b.Score {
					return 1
				}
				return 0
			})
			// TODO minimum score?
			// Whoever has the highest score gets put at the top this is the one that
			// will get merged into the existing cluster.
			pairs = append(pairs, Pair{
				New:      result[score[0].ResultIndex],
				Existing: existingClusterId,
			})
		}

		// upsert := make([]models.TransactionCluster, 0, len(result))
		// for _, item := range merged {
		// 	// If the cluster doesn't overlap at all with any existing clusters then
		// 	// we can simply insert/upsert this cluster without any issues!
		// 	if len(item.Items) == 0 {
		// 		upsert = append(upsert, item.New)
		// 		continue
		// 	}
		//
		// 	// If the cluster does overlap with any clusters then we want to find out
		// 	// how much!
		//
		// }

		if _, err := repo.WriteTransactionClusters(ctx, args.BankAccountId, result); err != nil {
			return errors.Wrap(err, "failed to persist the calculated transaction clusters")
		}

		return nil
	})
}
