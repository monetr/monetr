package platypus

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/consts"
	"github.com/monetr/monetr/pkg/internal/myownsanity"

	"github.com/getsentry/sentry-go"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
)

type (
	Client interface {
		GetAccounts(ctx context.Context, accountIds ...string) ([]BankAccount, error)
		GetAllTransactions(ctx context.Context, start, end time.Time, accountIds []string) ([]Transaction, error)
		UpdateItem(ctx context.Context) (LinkToken, error)
		RemoveItem(ctx context.Context) error
	}
)

var (
	_ Client = &PlaidClient{}
)

type PlaidClient struct {
	accountId   uint64
	linkId      uint64
	accessToken string
	log         *logrus.Entry
	client      *plaid.APIClient
	config      config.Plaid
}

func (p *PlaidClient) getLog(span *sentry.Span) *logrus.Entry {
	return p.log.WithContext(span.Context()).WithField("plaid", span.Op)
}

func (p *PlaidClient) GetAccounts(ctx context.Context, accountIds ...string) ([]BankAccount, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetAccount")
	defer span.Finish()

	log := p.getLog(span)

	// By default report the accountIds as "all accounts" to sentry. This way we know that if we are not requesting
	// specific accounts then we are requesting all of them.
	span.Data = map[string]interface{}{
		"accountIds": "ALL_BANK_ACCOUNTS",
	}

	// If however we are requesting specific accounts, overwrite the value.
	if len(accountIds) > 0 {
		span.Data["accountIds"] = accountIds
	}

	log.Trace("retrieving bank accounts from plaid")

	// Build the get accounts request.
	request := p.client.PlaidApi.
		AccountsGet(span.Context()).
		AccountsGetRequest(plaid.AccountsGetRequest{
			AccessToken: p.accessToken,
			Options: &plaid.AccountsGetRequestOptions{
				// This might not work, if it does not we should just add a nil check somehow here.
				AccountIds: &accountIds,
			},
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

func (p *PlaidClient) GetAllTransactions(ctx context.Context, start, end time.Time, accountIds []string) ([]Transaction, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetAllTransactions")
	defer span.Finish()

	transactions := make([]Transaction, 0)

	var perPage int32 = 500
	var offset int32 = 0
	for {
		someTransactions, err := p.GetTransactions(span.Context(), start, end, perPage, offset, accountIds)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, someTransactions...)
		if retrieved := int32(len(someTransactions)); retrieved == perPage {
			offset += retrieved
			continue
		}

		break
	}

	return transactions, nil
}

func (p *PlaidClient) GetTransactions(ctx context.Context, start, end time.Time, count, offset int32, bankAccountIds []string) ([]Transaction, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetTransactions")
	defer span.Finish()

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
			Secret:      nil,
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

func (p *PlaidClient) UpdateItem(ctx context.Context) (LinkToken, error) {
	span := sentry.StartSpan(ctx, "Plaid - UpdateItem")
	defer span.Finish()

	log := p.getLog(span)

	log.Trace("creating link token for update")

	redirectUri := fmt.Sprintf("https://%s/plaid/oauth-return", p.config.OAuthDomain)

	var webhooksUrl *string
	if p.config.WebhooksEnabled {
		if p.config.WebhooksDomain == "" {
			crumbs.Warn(span.Context(), "BUG: Plaid webhook domain is not present but webhooks are enabled.", "bug", nil)
		} else {
			webhooksUrl = myownsanity.StringP(p.config.GetWebhooksURL())
		}
	}

	request := p.client.PlaidApi.
		LinkTokenCreate(span.Context()).
		LinkTokenCreateRequest(plaid.LinkTokenCreateRequest{
			ClientName:   consts.PlaidClientName,
			Language:     consts.PlaidLanguage,
			CountryCodes: consts.PlaidCountries,
			User: plaid.LinkTokenCreateRequestUser{
				ClientUserId: strconv.FormatUint(p.accountId, 10),
			},
			Webhook:               webhooksUrl,
			AccessToken:           &p.accessToken,
			LinkCustomizationName: nil,
			RedirectUri:           &redirectUri,
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
	span := sentry.StartSpan(ctx, "Plaid - RemoveItem")
	defer span.Finish()

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
