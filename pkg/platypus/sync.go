package platypus

import (
	"context"

	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
)

type SyncResult struct {
	NextCursor string
	HasMore    bool
	New        map[string]Transaction
	Updated    map[string]Transaction
	Deleted    []string
}

func (p *PlaidClient) Sync(ctx context.Context, cursor *string) (*SyncResult, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"cursor": cursor,
	}

	log := p.getLog(span).WithFields(logrus.Fields{
		"cursor": cursor,
	})

	log.Trace("syncing with plaid")

	request := p.client.PlaidApi.
		TransactionsSync(span.Context()).
		TransactionsSyncRequest(plaid.TransactionsSyncRequest{
			AccessToken: p.accessToken,
			Cursor:      cursor,
			Count:       myownsanity.Int32P(500),
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Syncing transactions with Plaid",
		"failed to sync transaction data with Plaid",
	); err != nil {
		log.WithError(err).Error("failed to sync transaction data with Plaid")
		return nil, err
	}

	added := map[string]Transaction{}
	for _, transaction := range result.Added {
		added[transaction.TransactionId], err = NewTransactionFromPlaid(transaction)
		if err != nil {
			return nil, err
		}
	}

	modified := map[string]Transaction{}
	for _, transaction := range result.Modified {
		modified[transaction.TransactionId], err = NewTransactionFromPlaid(transaction)
		if err != nil {
			return nil, err
		}
	}

	removed := make([]string, len(result.Removed))
	for i, transaction := range result.Removed {
		removed[i] = transaction.GetTransactionId()
	}

	if len(added)+len(modified)+len(removed) == 0 {
		log.Debug("no changes observed from Plaid via sync")
	} else {
		log.WithFields(logrus.Fields{
			"added":    len(added),
			"modified": len(modified),
			"removed":  len(removed),
		}).Debug("received changes from Plaid via sync")
	}

	return &SyncResult{
		NextCursor: result.GetNextCursor(),
		HasMore:    result.GetHasMore(),
		New:        added,
		Updated:    modified,
		Deleted:    removed,
	}, nil
}
