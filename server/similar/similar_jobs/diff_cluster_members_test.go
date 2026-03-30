package similar_jobs

import (
	"context"
	"slices"
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func memberHelper(
	txn models.ID[models.Transaction],
	cluster models.ID[models.TransactionCluster],
) models.TransactionClusterMember {
	return models.TransactionClusterMember{
		TransactionId:        txn,
		AccountId:            "acct_test",
		BankAccountId:        "bac_test",
		TransactionClusterId: cluster,
	}
}

func clusterHelper(
	id models.ID[models.TransactionCluster],
	members ...models.ID[models.Transaction],
) models.TransactionCluster {
	return models.TransactionCluster{
		TransactionClusterId: id,
		Members:              members,
	}
}

func TestDiffClusterMembers(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		diff := DiffClusterMembers(context.Background(), nil, nil, "acct_test", "bac_test")
		assert.Empty(t, diff.UpsertClusters)
		assert.Empty(t, diff.DeleteClusterIds)
		assert.Empty(t, diff.InsertMembers)
		assert.Empty(t, diff.UpdateMembers)
		assert.Empty(t, diff.DeleteMemberIds)
	})

	t.Run("no existing all new", func(t *testing.T) {
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_new1", "txn_1", "txn_2"),
			clusterHelper("tcl_new2", "txn_3"),
		}

		diff := DiffClusterMembers(context.Background(), nil, newClusters, "acct_test", "bac_test")

		assert.Len(t, diff.UpsertClusters, 2)
		assert.Empty(t, diff.DeleteClusterIds)
		assert.Len(t, diff.InsertMembers, 3)
		assert.Empty(t, diff.UpdateMembers)
		assert.Empty(t, diff.DeleteMemberIds)
	})

	t.Run("existing but no new clusters", func(t *testing.T) {
		existing := []models.TransactionClusterMember{
			memberHelper("txn_1", "tcl_old1"),
			memberHelper("txn_2", "tcl_old1"),
			memberHelper("txn_3", "tcl_old2"),
		}

		diff := DiffClusterMembers(context.Background(), existing, nil, "acct_test", "bac_test")

		// Everything should be cleaned up since nothing was recalculated.
		assert.Empty(t, diff.UpsertClusters)
		assert.Len(t, diff.DeleteClusterIds, 2)
		assert.Empty(t, diff.InsertMembers)
		assert.Empty(t, diff.UpdateMembers)
		assert.Len(t, diff.DeleteMemberIds, 3)
	})

	t.Run("perfect match same membership", func(t *testing.T) {
		existing := []models.TransactionClusterMember{
			memberHelper("txn_1", "tcl_old1"),
			memberHelper("txn_2", "tcl_old1"),
			memberHelper("txn_3", "tcl_old2"),
		}
		// Same members but the algorithm generates fresh IDs every time, so these
		// should get matched back to the originals via jaccard.
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_fresh1", "txn_1", "txn_2"),
			clusterHelper("tcl_fresh2", "txn_3"),
		}

		diff := DiffClusterMembers(context.Background(), existing, newClusters, "acct_test", "bac_test")

		assert.Len(t, diff.UpsertClusters, 2)
		assert.Empty(t, diff.DeleteClusterIds)
		// Membership is identical so nothing to do on the member side.
		assert.Empty(t, diff.InsertMembers)
		assert.Empty(t, diff.UpdateMembers)
		assert.Empty(t, diff.DeleteMemberIds)
	})

	t.Run("cluster gains a new member", func(t *testing.T) {
		existing := []models.TransactionClusterMember{
			memberHelper("txn_1", "tcl_old1"),
			memberHelper("txn_2", "tcl_old1"),
		}
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_fresh1", "txn_1", "txn_2", "txn_3"),
		}

		diff := DiffClusterMembers(context.Background(), existing, newClusters, "acct_test", "bac_test")

		assert.Len(t, diff.UpsertClusters, 1)
		assert.Equal(t, models.ID[models.TransactionCluster]("tcl_old1"), diff.UpsertClusters[0].TransactionClusterId)
		assert.Empty(t, diff.DeleteClusterIds)
		assert.Len(t, diff.InsertMembers, 1)
		assert.Equal(t, models.ID[models.Transaction]("txn_3"), diff.InsertMembers[0].TransactionId)
		assert.Empty(t, diff.UpdateMembers)
		assert.Empty(t, diff.DeleteMemberIds)
	})

	t.Run("cluster loses a member", func(t *testing.T) {
		existing := []models.TransactionClusterMember{
			memberHelper("txn_1", "tcl_old1"),
			memberHelper("txn_2", "tcl_old1"),
			memberHelper("txn_3", "tcl_old1"),
		}
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_fresh1", "txn_1", "txn_2"),
		}

		diff := DiffClusterMembers(context.Background(), existing, newClusters, "acct_test", "bac_test")

		assert.Len(t, diff.UpsertClusters, 1)
		assert.Equal(t, models.ID[models.TransactionCluster]("tcl_old1"), diff.UpsertClusters[0].TransactionClusterId)
		assert.Empty(t, diff.DeleteClusterIds)
		assert.Empty(t, diff.InsertMembers)
		assert.Empty(t, diff.UpdateMembers)
		// The cluster itself survives, but txn_3 drops out of it.
		assert.Len(t, diff.DeleteMemberIds, 1)
		assert.Contains(t, diff.DeleteMemberIds, models.ID[models.Transaction]("txn_3"))
	})

	t.Run("two clusters merge into one", func(t *testing.T) {
		existing := []models.TransactionClusterMember{
			memberHelper("txn_1", "tcl_A"),
			memberHelper("txn_2", "tcl_A"),
			memberHelper("txn_3", "tcl_A"),
			memberHelper("txn_4", "tcl_B"),
			memberHelper("txn_5", "tcl_B"),
		}
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_fresh", "txn_1", "txn_2", "txn_3", "txn_4", "txn_5"),
		}

		diff := DiffClusterMembers(context.Background(), existing, newClusters, "acct_test", "bac_test")

		// A has 3/5 overlap vs B's 2/5, so A wins the match.
		assert.Len(t, diff.UpsertClusters, 1)
		assert.Equal(t, models.ID[models.TransactionCluster]("tcl_A"), diff.UpsertClusters[0].TransactionClusterId)
		assert.Len(t, diff.DeleteClusterIds, 1)
		assert.Contains(t, diff.DeleteClusterIds, models.ID[models.TransactionCluster]("tcl_B"))
		assert.Empty(t, diff.InsertMembers)
		// B's former members get reassigned to A.
		assert.Len(t, diff.UpdateMembers, 2)
		updateTxnIds := []models.ID[models.Transaction]{
			diff.UpdateMembers[0].TransactionId,
			diff.UpdateMembers[1].TransactionId,
		}
		slices.Sort(updateTxnIds)
		assert.Equal(t, []models.ID[models.Transaction]{"txn_4", "txn_5"}, updateTxnIds)
		assert.Empty(t, diff.DeleteMemberIds)
	})

	t.Run("one cluster splits into two", func(t *testing.T) {
		existing := []models.TransactionClusterMember{
			memberHelper("txn_1", "tcl_old"),
			memberHelper("txn_2", "tcl_old"),
			memberHelper("txn_3", "tcl_old"),
			memberHelper("txn_4", "tcl_old"),
		}
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_freshA", "txn_1", "txn_2"),
			clusterHelper("tcl_freshB", "txn_3", "txn_4"),
		}

		diff := DiffClusterMembers(context.Background(), existing, newClusters, "acct_test", "bac_test")

		assert.Len(t, diff.UpsertClusters, 2)
		// Both halves have equal overlap (2/4) so whichever wins the greedy
		// match gets the old ID. Either way the old cluster is consumed — it
		// shouldn't show up in deletes.
		assert.Empty(t, diff.DeleteClusterIds)
		assignedIds := []models.ID[models.TransactionCluster]{
			diff.UpsertClusters[0].TransactionClusterId,
			diff.UpsertClusters[1].TransactionClusterId,
		}
		assert.Contains(t, assignedIds, models.ID[models.TransactionCluster]("tcl_old"))

		assert.Empty(t, diff.InsertMembers)
		// The half that got a fresh ID needs its members updated.
		assert.Len(t, diff.UpdateMembers, 2)
		assert.Empty(t, diff.DeleteMemberIds)
	})

	t.Run("below similarity threshold treated as new", func(t *testing.T) {
		// 100 member cluster with only 1 shared transaction. Jaccard ends up at
		// ~0.0099 which is way below the 0.1 threshold, so it shouldn't match.
		existingTxns := make([]models.TransactionClusterMember, 100)
		for i := range existingTxns {
			existingTxns[i] = memberHelper(
				models.ID[models.Transaction]("txn_old_"+string(rune('a'+i))),
				"tcl_old",
			)
		}
		existingTxns[0] = memberHelper("txn_shared", "tcl_old")

		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_fresh", "txn_shared", "txn_brand_new"),
		}

		diff := DiffClusterMembers(context.Background(), existingTxns, newClusters, "acct_test", "bac_test")

		// 1 / (100 + 2 - 1) ~= 0.0099 - too low to match.
		assert.Len(t, diff.UpsertClusters, 1)
		assert.Equal(t, models.ID[models.TransactionCluster]("tcl_fresh"), diff.UpsertClusters[0].TransactionClusterId,
			"keeps its fresh ID because it didn't match anything")
		assert.Len(t, diff.DeleteClusterIds, 1)
		assert.Contains(t, diff.DeleteClusterIds, models.ID[models.TransactionCluster]("tcl_old"))
	})

	t.Run("member fields populated correctly", func(t *testing.T) {
		newClusters := []models.TransactionCluster{
			clusterHelper("tcl_new", "txn_1"),
		}

		diff := DiffClusterMembers(context.Background(), nil, newClusters, "acct_test", "bac_test")

		assert.Len(t, diff.InsertMembers, 1)
		m := diff.InsertMembers[0]
		assert.Equal(t, models.ID[models.Transaction]("txn_1"), m.TransactionId)
		assert.Equal(t, models.ID[models.Account]("acct_test"), m.AccountId)
		assert.Equal(t, models.ID[models.BankAccount]("bac_test"), m.BankAccountId)
		assert.Equal(t, models.ID[models.TransactionCluster]("tcl_new"), m.TransactionClusterId)
	})
}
