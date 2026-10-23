package platypus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	plaid "github.com/plaid/plaid-go/v49"
	"github.com/plaid/plaid-go/v49/client"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=client.go -package=mockgen -destination=../internal/mockgen/platypus_client.go Client
type (
	Client interface {
		GetAccounts(ctx context.Context, accountIds ...string) ([]BankAccount, error)
		GetAllTransactions(ctx context.Context, start, end time.Time, accountIds []string) ([]Transaction, error)
		// UpdateItem will create a LinkToken that is used to allow the client to update this particular link. This can be
		// used to resolve issues with the link and its authentication. Or can be used to add/remove accounts that monetr
		// has access to via Plaid's API.
		UpdateItem(ctx context.Context, updateAccountSelection bool) (LinkToken, error)
		// UpdateWebhook will update the webhook endpoint for the current link based
		// on the currently configured webhook domain.
		UpdateWebhook(ctx context.Context) error
		RemoveItem(ctx context.Context) error
		// Sync takes a cursor (or lack of one) and retrieves transaction data from Plaid that is newer than that cursor.
		Sync(ctx context.Context, cursor *string) (*SyncResult, error)

		RefeshTransactions(ctx context.Context) error
	}
)

var (
	_ Client = &PlaidClient{}
)

type PlaidClient struct {
	accountId   models.ID[models.Account]
	linkId      models.ID[models.Link]
	accessToken string
	itemId      string
	log         *slog.Logger
	client      *client.Client
	config      config.Plaid
}

func (p *PlaidClient) getLog(span *sentry.Span) *slog.Logger {
	return p.log.With(
		"operation", span.Op,
		slog.Group("plaid",
			"itemId", p.itemId,
		),
		// TODO Eventually remove this from the log structure.
		"itemId", p.itemId,
	)
}

func (p *PlaidClient) GetAccounts(ctx context.Context, accountIds ...string) ([]BankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("itemId", p.itemId)

	log := p.getLog(span)

	// By default report the accountIds as "all accounts" to sentry. This way we know that if we are not requesting
	// specific accounts then we are requesting all of them.
	span.Data = map[string]any{
		"accountIds": "ALL_BANK_ACCOUNTS",
	}

	log.Log(span.Context(), logging.LevelTrace, "retrieving bank accounts from plaid")

	var options *plaid.AccountsGetRequestOptions
	// If however we are requesting specific accounts, overwrite the value.
	if len(accountIds) > 0 {
		span.Data["accountIds"] = accountIds
		options = &plaid.AccountsGetRequestOptions{
			AccountIDs: accountIds,
		}
	}

	// Send the request.
	result, err := p.client.Accounts.Get(
		span.Context(),
		&plaid.AccountsGetRequest{
			AccessToken: p.accessToken,
			Options:     options,
		},
	)
	// And handle the response.
	if err = after(
		span,
		err,
		"Retrieving bank accounts from Plaid",
		"failed to retrieve bank accounts from plaid",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve bank accounts from plaid", "err", err)
		return nil, err
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

	return accounts, nil
}

func (p *PlaidClient) GetAllTransactions(
	ctx context.Context,
	start, end time.Time,
	accountIds []string,
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("itemId", p.itemId)

	transactions := make([]Transaction, 0)

	var limit int32 = 500
	var offset int32 = 0
	for {
		someTransactions, err := p.GetTransactions(
			span.Context(),
			start, end,
			limit, offset,
			accountIds,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, someTransactions...)
		if retrieved := int32(len(someTransactions)); retrieved == limit {
			offset += retrieved
			continue
		}

		break
	}

	return transactions, nil
}

func (p *PlaidClient) GetTransactions(
	ctx context.Context,
	start, end time.Time,
	count, offset int32,
	bankAccountIds []string,
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("itemId", p.itemId)

	span.Data = map[string]any{
		"accountIds": bankAccountIds,
		"start":      start.Format("2006-01-02"),
		"end":        end.Format("2006-01-02"),
	}

	log := p.getLog(span)

	log.Log(span.Context(), logging.LevelTrace, "retrieving transactions")

	// Send the request.
	result, err := p.client.Transactions.Get(
		span.Context(),
		&plaid.TransactionsGetRequest{
			Options: &plaid.TransactionsGetRequestOptions{
				AccountIDs:                 bankAccountIds,
				Count:                      new(int(count)),
				Offset:                     new(int(offset)),
				IncludeOriginalDescription: nil,
			},
			AccessToken: p.accessToken,
			StartDate:   start,
			EndDate:     end,
		},
	)
	// And handle the response.
	if err = after(
		span,
		err,
		"Retrieving transactions from Plaid",
		"failed to retrieve transactions from plaid",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve transactions from plaid", "err", err)
		return nil, err
	}

	// We do our own paging in GetAllTransactions, so we only want the results
	// from this single page.
	transactions := make([]Transaction, len(result.Results))
	for i, transaction := range result.Results {
		transactions[i], err = NewTransactionFromPlaid(transaction)
		if err != nil {
			return nil, err
		}
	}

	return transactions, nil
}

func (p *PlaidClient) UpdateItem(
	ctx context.Context,
	updateAccountSelection bool,
) (LinkToken, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("itemId", p.itemId)

	log := p.getLog(span)

	var redirectUri *string
	if p.config.OAuthDomain != "" {
		// Normally we would substitute the configured protocol, but Plaid
		// _requires_ that we use HTTPS for oauth callbacks. So if the monetr server
		// is not configured for TLS that sucks because this won't work.
		redirectUri = new(fmt.Sprintf(
			"https://%s/plaid/oauth-return", p.config.OAuthDomain,
		))
		log = log.With("redirectUri", *redirectUri)
	}

	var webhooksUrl *string
	if p.config.WebhooksEnabled {
		if p.config.WebhooksDomain == "" {
			crumbs.Warn(span.Context(), "BUG: Plaid webhook domain is not present but webhooks are enabled.", "bug", nil)
		} else {
			webhooksUrl = new(p.config.GetWebhooksURL())
			log = log.With("webhooksUrl", *webhooksUrl)
		}
	}

	log.Log(span.Context(), logging.LevelTrace, "creating link token for update")

	result, err := p.client.Link.CreateToken(
		span.Context(),
		&plaid.LinkTokenCreateRequest{
			ClientName:   consts.PlaidClientName,
			Language:     consts.PlaidLanguage,
			CountryCodes: p.config.CountryCodes,
			User: &plaid.LinkTokenCreateRequestUser{
				ClientUserID: p.accountId.String(),
				EmailAddress: nil,
			},
			Webhook:               webhooksUrl,
			AccessToken:           &p.accessToken,
			LinkCustomizationName: nil,
			RedirectURI:           redirectUri,
			Update: &plaid.LinkTokenCreateRequestUpdate{
				AccountSelectionEnabled: new(updateAccountSelection),
			},
		},
	)
	if err = after(
		span,
		err,
		"Updating Plaid link token",
		"failed to update Plaid link token",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to create link token", "err", err)
		return nil, err
	}

	return PlaidLinkToken{
		LinkToken: result.LinkToken,
		Expires:   result.Expiration,
	}, nil
}

func (p *PlaidClient) RemoveItem(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("itemId", p.itemId)

	log := p.getLog(span).With("itemId", p.itemId)

	log.DebugContext(span.Context(), "removing item from Plaid")

	// Send the request.
	_, err := p.client.Item.Remove(
		span.Context(),
		&plaid.ItemRemoveRequest{
			AccessToken: p.accessToken,
		},
	)
	// And handle the response.
	if err = after(
		span,
		err,
		"Removing Plaid item",
		"failed to remove Plaid item",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to remove Plaid item", "err", err)
		return err
	}

	return nil
}

func (p *PlaidClient) RefeshTransactions(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "http.client")
	defer span.Finish()
	span.Description = "Plaid - ForceRefresh"

	span.SetTag("itemId", p.itemId)

	log := p.getLog(span)

	log.Log(span.Context(), logging.LevelTrace, "force refreshing transactions for item")

	// Send the request.
	_, err := p.client.Transactions.Refresh(
		span.Context(),
		&plaid.TransactionsRefreshRequest{
			AccessToken: p.accessToken,
		},
	)
	// And handle the response.
	if err = after(
		span,
		err,
		"Triggering transaction refresh for Plaid item",
		"failed to trigger transaction refresh for Plaid item",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to trigger transaction refresh for Plaid item", "err", err)
		return err
	}

	return nil
}

func (p *PlaidClient) UpdateWebhook(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var webhookUrl *string = nil
	if p.config.WebhooksEnabled {
		webhookUrl = new(p.config.GetWebhooksURL())
	}

	log := p.getLog(span).With(
		slog.Group("plaid",
			"itemId", p.itemId,
			"webhookUrl", webhookUrl,
		),
	)

	log.InfoContext(span.Context(), "updating webhook URL for plaid link")

	request := &plaid.ItemWebhookUpdateRequest{
		AccessToken: p.accessToken,
	}
	// Plaid removes the webhook from the item when we send null, but the SDK
	// leaves nil fields off entirely. The setter makes sure null is actually sent
	// when webhooks are disabled.
	request.SetWebhook(webhookUrl)

	_, err := p.client.Item.UpdateWebhook(span.Context(), request)
	if err = after(
		span,
		err,
		"Updating webhook URL for Plaid link",
		"failed to update webhook url for plaid link",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to update webhook url for plaid link", "err", err)
		return err
	}

	return nil
}
