package similar_jobs

import (
	"cmp"
	"context"
	"slices"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
)

const (
	// Anything below this jaccard score means the overlap is too weak to
	// consider it the same cluster. Just treat it as a new one.
	minSimilarityThreshold = 0.1
)

// MemberDiff is everything that changed between the clusters we have stored and
// the ones we just calculated. Keep in mind the order you write these in matters
// because of the foreign keys.
type MemberDiff struct {
	// Every new cluster. If it matched an old one then it keeps the old ID,
	// otherwise it keeps the ID from the algorithm
	UpsertClusters []models.TransactionCluster
	// Old clusters that didn't match anything new, these can just be deleted
	DeleteClusterIds []models.ID[models.TransactionCluster]
	// Transactions that weren't in any cluster before.
	InsertMembers []models.Transaction
	// Transactions that moved from one cluster to a different one.
	UpdateMembers []models.Transaction
	// Transactions that aren't in any cluster anymore. Only needed for clusters
	// that still exist, if a cluster gets deleted then the FK clears the ID for us
	DeleteMemberIds []models.ID[models.Transaction]
}

// DiffClusterMembers figures out what needs to be written to get the database in
// line with the clusters we just calculated. New clusters get matched up with
// old ones based on how many members they share (jaccard), this way a cluster
// can keep its ID between calculations.
func DiffClusterMembers(
	ctx context.Context,
	existingMembers []models.Transaction,
	newClusters []models.TransactionCluster,
	accountId models.ID[models.Account],
	bankAccountId models.ID[models.BankAccount],
) MemberDiff {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// We need the existing members grouped by cluster for the jaccard scoring.
	existingByCluster := myownsanity.GroupByMapV(
		existingMembers,
		func(m models.Transaction) models.ID[models.TransactionCluster] {
			return *m.TransactionClusterId
		},
		func(m models.Transaction) models.ID[models.Transaction] {
			return m.TransactionId
		},
	)

	// Flat lookup of txnId -> clusterId for the current state in the database.
	oldOwner := make(
		map[models.ID[models.Transaction]]models.ID[models.TransactionCluster],
		len(existingMembers),
	)
	for _, m := range existingMembers {
		oldOwner[m.TransactionId] = *m.TransactionClusterId
	}

	// Score every (new, existing) cluster pair that shares at least one member.
	// No intersection means they can't be the same cluster so don't bother.
	type scoredMatch struct {
		newIdx     int
		existingId models.ID[models.TransactionCluster]
		score      float64
	}
	var matches []scoredMatch
	for i, nc := range newClusters {
		for existingId, existingTxnIds := range existingByCluster {
			intersect := myownsanity.Intersection(nc.Members, existingTxnIds)
			if len(intersect) == 0 {
				continue
			}
			union := myownsanity.Union(nc.Members, existingTxnIds)
			matches = append(matches, scoredMatch{
				newIdx:     i,
				existingId: existingId,
				score:      float64(len(intersect)) / float64(len(union)),
			})
		}
	}

	// Best matches first. The matches came from a map so their order is random,
	// if two scores tie we need to break it the same way every time or the ID
	// could bounce between clusters. Lower existing ID wins, then whichever new
	// cluster the algorithm returned first. Can't use the new cluster IDs since
	// those are random
	slices.SortFunc(matches, func(a, b scoredMatch) int {
		return cmp.Or(
			cmp.Compare(b.score, a.score),
			cmp.Compare(a.existingId, b.existingId),
			cmp.Compare(a.newIdx, b.newIdx),
		)
	})

	// Walk the matches best to worst and pair them up, a cluster on either side
	// can only be claimed once
	claimedNew := make(map[int]struct{}, len(newClusters))
	claimedExisting := make(
		map[models.ID[models.TransactionCluster]]struct{},
		len(existingByCluster),
	)
	matched := map[int]models.ID[models.TransactionCluster]{}

	for _, m := range matches {
		// Sorted, so everything after this is below the threshold too
		if m.score < minSimilarityThreshold {
			break
		}
		// If we have already assigned a match to the new cluster then we want to
		// throw aside any candidates that are a lesser match for that new cluster.
		if _, ok := claimedNew[m.newIdx]; ok {
			continue
		}
		// On the other side, if the existing cluster that we are matched against
		// has already been claimed by a preceding item (a cluster with a higher
		// overlap percentage) then we have to skip that too.
		if _, ok := claimedExisting[m.existingId]; ok {
			continue
		}
		matched[m.newIdx] = m.existingId
		claimedNew[m.newIdx] = struct{}{}
		claimedExisting[m.existingId] = struct{}{}
	}

	// Carry over the existing ID for matched clusters so things like transaction
	// rules that reference the cluster don't break between recalculations.
	clusters := make([]models.TransactionCluster, len(newClusters))
	copy(clusters, newClusters)
	for i := range clusters {
		if existingId, ok := matched[i]; ok {
			clusters[i].TransactionClusterId = existingId
		}
	}

	// Any existing cluster that wasn't claimed by a new cluster is gone entirely.
	var deleteClusterIds []models.ID[models.TransactionCluster]
	for existingId := range existingByCluster {
		// Basically if none of the newly calculated clusters had a claim towards
		// this particular existing cluster, then that means this cluster is gone
		// entirely.
		if _, ok := claimedExisting[existingId]; !ok {
			deleteClusterIds = append(deleteClusterIds, existingId)
		}
	}

	// Build the new ownership map from the clusters with their final IDs
	// assigned. If we merged a cluster with an existing one then we need to
	// update all the member records to reflect that new ID.
	newOwner := map[models.ID[models.Transaction]]models.ID[models.TransactionCluster]{}
	for _, c := range clusters {
		for _, txnId := range c.Members {
			newOwner[txnId] = c.TransactionClusterId
		}
	}

	// Compare the old and new ownership to figure out what actually changed.
	var diff MemberDiff
	diff.UpsertClusters = clusters
	diff.DeleteClusterIds = deleteClusterIds

	for txnId, newClusterId := range newOwner {
		member := models.Transaction{
			TransactionId:        txnId,
			AccountId:            accountId,
			BankAccountId:        bankAccountId,
			TransactionClusterId: &newClusterId,
		}
		oldClusterId, existed := oldOwner[txnId]
		if !existed {
			// Keeping inserts versus updates separate lets us fire off events based
			// on what transactions are being added as part of this calculation.
			diff.InsertMembers = append(diff.InsertMembers, member)
		} else if oldClusterId != newClusterId {
			diff.UpdateMembers = append(diff.UpdateMembers, member)
		}
		// If the transaction is in the same cluster as before then there is nothing
		// to do for this transaction.
	}

	for txnId := range oldOwner {
		if _, ok := newOwner[txnId]; !ok {
			diff.DeleteMemberIds = append(diff.DeleteMemberIds, txnId)
		}
	}

	return diff
}
