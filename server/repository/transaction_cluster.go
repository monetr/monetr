package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (r *repositoryBase) WriteTransactionClusters(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	clusters []TransactionCluster,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range clusters {
		clusters[i].AccountId = r.AccountId()
		clusters[i].BankAccountId = bankAccountId
	}

	{ // Will be removed as soon as I know more about how effective this is!
		for i := range clusters {
			cluster := clusters[i]
			if cluster.Centroid == nil {
				continue
			}
			// Throw the error away, we don't need it here, we just want to make sure we
			// are testing this to see how good the signature system is!
			_ = r.TestTransactionClustersBySignatureAndCentroid(
				span.Context(),
				bankAccountId,
				cluster.Signature,
				*cluster.Centroid,
			)
		}
	}

	_, err := r.txn.ModelContext(span.Context(), new(TransactionCluster)).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Delete()
	if err != nil {
		return errors.Wrap(err, "failed to delete existing transaction clusters")
	}

	_, err = r.txn.ModelContext(span.Context(), &clusters).Insert(&clusters)
	if err != nil {
		return errors.Wrap(err, "failed to insert the new transaction clusters")
	}

	return nil
}

func (r *repositoryBase) GetTransactionClusterByMember(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionId ID[Transaction],
) (*TransactionCluster, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var cluster TransactionCluster
	err := r.txn.ModelContext(span.Context(), &cluster).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Where(`? = ANY ("members")`, transactionId).
		Limit(1).
		Select(&cluster)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find cluster containing transaction")
	}

	return &cluster, nil
}

// TestTransactionClustersBySignatureAndCentroid takes the current bank account
// as well as the signature and centroid of a transaction cluster and looks for
// existing transaction clusters that have the same signature and centroid. This
// will be a debugging run in order to see if it is viable to use this to
// deduplicate transaction clusters in the future. This function does not return
// anything other than an error if it fails. It does however log something
// depending on the results and that log entry should monitored!
func (r *repositoryBase) TestTransactionClustersBySignatureAndCentroid(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	signature string,
	centroid ID[Transaction],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := r.log.WithContext(span.Context()).WithFields(logrus.Fields{
		// Monitor for this needle!
		"needle":        "transaction-cluster-deduplication",
		"bankAccountId": bankAccountId,
		"signature":     signature,
		"centroid":      centroid,
	})

	count, err := r.txn.ModelContext(span.Context(), &TransactionCluster{}).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Where(`"signature" = ?`, signature).
		Where(`"centroid" = ?`, centroid).
		Count()
	if err != nil {
		log.WithError(err).Error("couldn't count transaction clusters for some reason?")
		return errors.WithStack(err)
	}

	// Capture this count via the needle and look for how many of these are more
	// than 1
	log.WithFields(logrus.Fields{
		"count": count,
		// Not sure if Loki lets me filter by whether or not a field is greater than
		// or less than so just to make it easy I'll make this field a boolean so
		// its easier for me to filter.
		"gt_one": count > 1,
	}).Debug("testing transaction cluster deduplication!")

	return nil
}
