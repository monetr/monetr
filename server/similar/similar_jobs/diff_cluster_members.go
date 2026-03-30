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

// MemberDiff is the result of diffing existing cluster membership against a
// freshly calculated set of clusters. Each field maps to a specific database
// operation that the caller needs to execute in FK-safe order.
type MemberDiff struct {
	// All new clusters with stable IDs assigned. Matched clusters keep their
	// existing ID, brand new clusters keep the ID from the algorithm.
	UpsertClusters []models.TransactionCluster
	// Existing clusters that dissolved entirely — no new cluster claimed them.
	DeleteClusterIds []models.ID[models.TransactionCluster]
	// Transactions that weren't in any cluster before.
	InsertMembers []models.TransactionClusterMember
	// Transactions that moved from one cluster to a different one.
	UpdateMembers []models.TransactionClusterMember
	// DeleteMemberIds are transaction IDs that are no longer in any cluster and
	// should be removed from the members table. This is different from delete
	// cluster IDs because those will automatically cascade their member deletes.
	DeleteMemberIds []models.ID[models.Transaction]
}

// DiffClusterMembers takes the existing cluster membership and a freshly
// calculated set of clusters and figures out the minimal set of DB operations
// to get from one state to the other. Clusters are matched to their previous
// incarnation via jaccard similarity so we can preserve their IDs.
func DiffClusterMembers(
	ctx context.Context,
	existingMembers []models.TransactionClusterMember,
	newClusters []models.TransactionCluster,
	accountId models.ID[models.Account],
	bankAccountId models.ID[models.BankAccount],
) MemberDiff {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// We need the existing members grouped by cluster for the jaccard scoring.
	existingByCluster := myownsanity.GroupByMapV(
		existingMembers,
		func(m models.TransactionClusterMember) models.ID[models.TransactionCluster] {
			return m.TransactionClusterId
		},
		func(m models.TransactionClusterMember) models.ID[models.Transaction] {
			return m.TransactionId
		},
	)

	// Flat lookup of txnId -> clusterId for the current state in the database.
	oldOwner := make(
		map[models.ID[models.Transaction]]models.ID[models.TransactionCluster],
		len(existingMembers),
	)
	for _, m := range existingMembers {
		oldOwner[m.TransactionId] = m.TransactionClusterId
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

	// Best matches first.
	slices.SortFunc(matches, func(a, b scoredMatch) int {
		return cmp.Compare(b.score, a.score)
	})

	// Greedy 1:1 assignment. Walk the sorted matches and claim each pair if
	// neither side has been claimed yet.
	claimedNew := make(map[int]struct{}, len(newClusters))
	claimedExisting := make(
		map[models.ID[models.TransactionCluster]]struct{},
		len(existingByCluster),
	)
	matched := map[int]models.ID[models.TransactionCluster]{}

	for _, m := range matches {
		// Sorted descending, everything after this is below threshold.
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
		member := models.TransactionClusterMember{
			TransactionId:        txnId,
			AccountId:            accountId,
			BankAccountId:        bankAccountId,
			TransactionClusterId: newClusterId,
		}
		oldClusterId, existed := oldOwner[txnId]
		if !existed {
			// Keeping inserts versus updates separate lets us fire off events based
			// on what transactions are being added as part of this calculation.
			diff.InsertMembers = append(diff.InsertMembers, member)
		} else if oldClusterId != newClusterId {
			diff.UpdateMembers = append(diff.UpdateMembers, member)
		}
		// If the transaction is in the same cluster as before then there is
		// nothing to do for this member row.
	}

	for txnId := range oldOwner {
		if _, ok := newOwner[txnId]; !ok {
			diff.DeleteMemberIds = append(diff.DeleteMemberIds, txnId)
		}
	}

	return diff
}
