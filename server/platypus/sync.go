package platypus

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	plaid "github.com/plaid/plaid-go/v49"
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
	log = log.With("cursor", cursor)

	log.Log(span.Context(), logging.LevelTrace, "syncing with plaid")

	result, err := p.client.Transactions.Sync(
		span.Context(),
		&plaid.TransactionsSyncRequest{
			AccessToken: p.accessToken,
			Cursor:      cursor,
			Count:       new(500),
			Options: &plaid.TransactionsSyncRequestOptions{
				IncludeOriginalDescription:     new(true),
				IncludePersonalFinanceCategory: new(true),
				IncludeLogoAndCounterpartyBeta: new(true),
			},
		},
	)
	if err = after(
		span,
		err,
		"Syncing with Plaid",
		"failed to sync data with Plaid",
	); err != nil {
		log.WarnContext(span.Context(), "failed to sync data with Plaid", "err", err)
		return nil, err
	}
	span.SetTag("plaid.requestId", result.RequestID)

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
		removed[i] = transaction.GetTransactionID()
	}

	plaidAccounts := result.Accounts
	accounts := make([]BankAccount, len(plaidAccounts))

	// Once we have our data, convert all of the results from our request to our own bank account interface.
	for i, plaidAccount := range plaidAccounts {
		accounts[i], err = NewPlaidBankAccount(plaidAccount)
		if err != nil {
			log.ErrorContext(span.Context(), "failed to convert bank account",
				"err", err,
				"bankAccountId", plaidAccount.GetAccountID(),
			)
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
		log.DebugContext(span.Context(), "no changes observed from Plaid via sync")
	} else {
		log.DebugContext(span.Context(), "received changes from Plaid via sync",
			"added", len(added),
			"modified", len(modified),
			"removed", len(removed),
		)
	}

	return &SyncResult{
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
		New:        added,
		Updated:    modified,
		Deleted:    removed,
		Accounts:   accounts,
	}, nil
}
