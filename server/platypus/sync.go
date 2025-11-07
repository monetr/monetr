package platypus

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/sirupsen/logrus"
)

type SyncResult struct {
	NextCursor string
	HasMore    bool
	New        []Transaction
	Updated    []Transaction
	Deleted    []string
	Accounts   []BankAccount
}

func (p *PlaidClient) Sync(ctx context.Context, cursor *string) (*SyncResult, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetTag("itemId", p.itemId)
	span.Data = map[string]any{
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
		log.WithError(err).Warn("failed to sync data with Plaid")
		return nil, err
	}
	span.SetTag("plaid.requestId", result.GetRequestId())

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

	plaidAccounts := result.GetAccounts()
	accounts := make([]BankAccount, len(plaidAccounts))

	// Once we have our data, convert all of the results from our request to our own bank account interface.
	for i, plaidAccount := range plaidAccounts {
		accounts[i], err = NewPlaidBankAccount(plaidAccount)
		if err != nil {
			log.WithError(err).
				WithField("bankAccountId", plaidAccount.GetAccountId()).
				Errorf("failed to convert bank account")
			crumbs.Error(span.Context(), "failed to convert bank account", "debug", map[string]any{
				// Maybe we don't want to report the entire account object here, but it'll sure save us a ton of time
				// if there is ever a problem with actually converting the account. This way we can actually see the
				// account object that caused the problem -> when it caused the problem.
				"bankAccount": plaidAccount,
			})
			return nil, err
		}
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
		Accounts:   accounts,
	}, nil
}
