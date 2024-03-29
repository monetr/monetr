package platypus

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/plaid/plaid-go/v20/plaid"
	"github.com/sirupsen/logrus"
)

type SyncResult struct {
	NextCursor string
	HasMore    bool
	New        []Transaction
	Updated    []Transaction
	Deleted    []string
}

func (p *PlaidClient) Sync(ctx context.Context, cursor *string) (*SyncResult, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"cursor": cursor,
	}

	log := p.getLog(span)
	if cursor != nil {
		log = log.WithField("cursor", cursor)
	} else {
		log = log.WithField("cursor", nil)
	}

	log.Trace("syncing with plaid")

	request := p.client.PlaidApi.
		TransactionsSync(span.Context()).
		TransactionsSyncRequest(plaid.TransactionsSyncRequest{
			AccessToken: p.accessToken,
			Cursor:      cursor,
			Count:       myownsanity.Int32P(500),
			Options: &plaid.TransactionsSyncRequestOptions{
				// Why does the constructor for the nullable bool return a pointer to a
				// nullable wrapper type? What the fuck? Absolutely fucking garbage
				// openapi code generator.
				IncludeOriginalDescription: *plaid.NewNullableBool(myownsanity.BoolP(true)),
				// Why the fuck is this a boolean pointer, but the field above is a
				// nullable boolean.
				IncludePersonalFinanceCategory: myownsanity.BoolP(true),
				IncludeLogoAndCounterpartyBeta: myownsanity.BoolP(true),
			},
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Syncing with Plaid",
		"failed to sync data with Plaid",
	); err != nil {
		log.WithError(err).Error("failed to sync data with Plaid")
		return nil, err
	}

	added := make([]Transaction, len(result.Added))
	for i, transaction := range result.Added {
		added[i], err = NewTransactionFromPlaid(transaction)
		if err != nil {
			return nil, err
		}
	}

	modified := make([]Transaction, len(result.Modified))
	for i, transaction := range result.Modified {
		modified[i], err = NewTransactionFromPlaid(transaction)
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
