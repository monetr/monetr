package platypus

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/sirupsen/logrus"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=client.go -package=mockgen -destination=../internal/mockgen/platypus_client.go Client
type (
	Client interface {
		GetAccounts(ctx context.Context, accountIds ...string) ([]BankAccount, error)
		GetAllTransactions(ctx context.Context, start, end time.Time, accountIds []string) ([]Transaction, error)
		// UpdateItem will create a LinkToken that is used to allow the client to update this particular link. This can be
		// used to resolve issues with the link and its authentication. Or can be used to add/remove accounts that monetr
		// has access to via Plaid's API.
		UpdateItem(ctx context.Context, updateAccountSelection bool) (LinkToken, error)
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
	log         *logrus.Entry
	client      *plaid.APIClient
	config      config.Plaid
}

func (p *PlaidClient) getLog(span *sentry.Span) *logrus.Entry {
	return p.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"plaid":  span.Op,
		"itemId": p.itemId,
	})
}

func (p *PlaidClient) toTransactionMap(input []plaid.Transaction) (map[string]Transaction, error) {
	var err error
	transactions := map[string]Transaction{}
	for _, transaction := range input {
		transactions[transaction.TransactionId], err = NewTransactionFromPlaid(transaction)
		if err != nil {
			return nil, err
		}
	}

	return transactions, nil
}

func (p *PlaidClient) GetAccounts(ctx context.Context, accountIds ...string) ([]BankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("itemId", p.itemId)

	log := p.getLog(span)

	// By default report the accountIds as "all accounts" to sentry. This way we know that if we are not requesting
	// specific accounts then we are requesting all of them.
	span.Data = map[string]interface{}{
		"accountIds": "ALL_BANK_ACCOUNTS",
	}

	log.Trace("retrieving bank accounts from plaid")

	var options *plaid.AccountsGetRequestOptions
	// If however we are requesting specific accounts, overwrite the value.
	if len(accountIds) > 0 {
		span.Data["accountIds"] = accountIds
		options = &plaid.AccountsGetRequestOptions{
			AccountIds: &accountIds,
		}
	}

	// Build the get accounts request.
	request := p.client.PlaidApi.
		AccountsGet(span.Context()).
		AccountsGetRequest(plaid.AccountsGetRequest{
			AccessToken: p.accessToken,
			Options:     options,
		})

	// Send the request.
	result, response, err := request.Execute()
	// And handle the response.
	if err = after(
		span,
		response,
		err,
		"Retrieving bank accounts from Plaid",
		"failed to retrieve bank accounts from plaid",
	); err != nil {
		log.WithError(err).Errorf("failed to retrieve bank accounts from plaid")
		return nil, err
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
			crumbs.Error(span.Context(), "failed to convert bank account", "debug", map[string]interface{}{
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
	span := sentry.StartSpan(ctx, "http.client")
	defer span.Finish()
	span.Description = "Plaid - GetTransactions"

	span.SetTag("itemId", p.itemId)

	span.Data = map[string]interface{}{
		"accountIds": bankAccountIds,
		"start":      start.Format("2006-01-02"),
		"end":        end.Format("2006-01-02"),
	}

	log := p.getLog(span)

	log.Trace("retrieving transactions")

	request := p.client.PlaidApi.
		TransactionsGet(span.Context()).
		TransactionsGetRequest(plaid.TransactionsGetRequest{
			Options: &plaid.TransactionsGetRequestOptions{
				AccountIds:                 &bankAccountIds,
				Count:                      &count,
				Offset:                     &offset,
				IncludeOriginalDescription: plaid.NullableBool{},
			},
			AccessToken: p.accessToken,
			StartDate:   start.Format("2006-01-02"),
			EndDate:     end.Format("2006-01-02"),
		})

	// Send the request.
	result, response, err := request.Execute()
	// And handle the response.
	if err = after(
		span,
		response,
		err,
		"Retrieving transactions from Plaid",
		"failed to retrieve transactions from plaid",
	); err != nil {
		log.WithError(err).Errorf("failed to retrieve transactions from plaid")
		return nil, err
	}

	transactions := make([]Transaction, len(result.Transactions))
	for i, transaction := range result.Transactions {
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
		// Normally we would substitute the configured protocol, but Plaid _requires_ that we use HTTPS for oauth callbacks.
		// So if the monetr server is not configured for TLS that sucks because this won't work.
		redirectUri = myownsanity.StringP(fmt.Sprintf(
			"https://%s/plaid/oauth-return", p.config.OAuthDomain,
		))
		log = log.WithField("redirectUri", *redirectUri)
	}

	var webhooksUrl *string
	if p.config.WebhooksEnabled {
		if p.config.WebhooksDomain == "" {
			crumbs.Warn(span.Context(), "BUG: Plaid webhook domain is not present but webhooks are enabled.", "bug", nil)
		} else {
			webhooksUrl = myownsanity.StringP(p.config.GetWebhooksURL())
			log = log.WithField("webhooksUrl", *webhooksUrl)
		}
	}

	log.Trace("creating link token for update")

	request := p.client.PlaidApi.
		LinkTokenCreate(span.Context()).
		LinkTokenCreateRequest(plaid.LinkTokenCreateRequest{
			ClientName:   consts.PlaidClientName,
			Language:     consts.PlaidLanguage,
			CountryCodes: p.config.CountryCodes,
			User: plaid.LinkTokenCreateRequestUser{
				ClientUserId: p.accountId.String(),
				EmailAddress: nil,
			},
			Webhook:               webhooksUrl,
			AccessToken:           *plaid.NewNullableString(&p.accessToken),
			LinkCustomizationName: nil,
			RedirectUri:           redirectUri,
			Update: &plaid.LinkTokenCreateRequestUpdate{
				AccountSelectionEnabled: myownsanity.BoolP(updateAccountSelection),
			},
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Updating Plaid link token",
		"failed to update Plaid link token",
	); err != nil {
		log.WithError(err).Errorf("failed to create link token")
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

	log := p.getLog(span)

	log.Trace("removing item")

	// Build the get accounts request.
	request := p.client.PlaidApi.
		ItemRemove(span.Context()).
		ItemRemoveRequest(plaid.ItemRemoveRequest{
			AccessToken: p.accessToken,
		})

	// Send the request.
	_, response, err := request.Execute()
	// And handle the response.
	if err = after(
		span,
		response,
		err,
		"Removing Plaid item",
		"failed to remove Plaid item",
	); err != nil {
		log.WithError(err).Errorf("failed to remove Plaid item")
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

	log.Trace("force refreshing transactions for item")

	request := p.client.PlaidApi.
		TransactionsRefresh(span.Context()).
		TransactionsRefreshRequest(plaid.TransactionsRefreshRequest{
			AccessToken: p.accessToken,
		})

	// Send the request.
	_, response, err := request.Execute()
	// And handle the response.
	if err = after(
		span,
		response,
		err,
		"Triggering transaction refresh for Plaid item",
		"failed to trigger transaction refresh for Plaid item",
	); err != nil {
		log.WithError(err).Errorf("failed to trigger transaction refresh for Plaid item")
		return err
	}

	return nil
}
