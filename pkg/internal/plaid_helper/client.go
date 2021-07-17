package plaid_helper

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Client interface {
	CreateLinkToken(ctx context.Context, config plaid.LinkTokenConfigs) (*plaid.CreateLinkTokenResponse, error)
	ExchangePublicToken(ctx context.Context, publicToken string) (*plaid.ExchangePublicTokenResponse, error)
	GetAccounts(ctx context.Context, accessToken string, options plaid.GetAccountsOptions) ([]plaid.Account, error)
	GetAllTransactions(ctx context.Context, accessToken string, start, end time.Time, accountIds []string) ([]plaid.Transaction, error)
	GetAllInstitutions(ctx context.Context, countryCodes []string, options plaid.GetInstitutionsOptions) ([]plaid.Institution, error)
	GetInstitutions(ctx context.Context, count, offset int, countryCodes []string, options plaid.GetInstitutionsOptions) (total int, _ []plaid.Institution, _ error)
	GetInstitution(ctx context.Context, institutionId string, includeMetadata bool, countryCodes []string) (*plaid.Institution, error)
	GetWebhookVerificationKey(ctx context.Context, keyId string) (plaid.GetWebhookVerificationKeyResponse, error)
	Close() error
}

var (
	_ Client = &plaidClient{}
)

func NewPlaidClient(log *logrus.Entry, options plaid.ClientOptions) Client {
	client, err := plaid.NewClient(options)
	if err != nil {
		// There currently isn't a code path that actually returns an error from the client. So if something happens
		// then its new.
		panic(err)
	}

	return &plaidClient{
		log:               log,
		client:            client,
		institutionTicker: time.NewTicker(2400 * time.Millisecond), // Limit our institution API calls to 25 per minute.
	}
}

type plaidClient struct {
	log               *logrus.Entry
	client            *plaid.Client
	institutionTicker *time.Ticker
}

func (p *plaidClient) CreateLinkToken(ctx context.Context, config plaid.LinkTokenConfigs) (*plaid.CreateLinkTokenResponse, error) {
	span := sentry.StartSpan(ctx, "Plaid - CreateLinkToken")
	defer span.Finish()

	span.Data = map[string]interface{}{}

	result, err := p.client.CreateLinkToken(config)
	span.Data["plaidRequestId"] = result.RequestID
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		err = errors.Wrap(err, "failed to create link token")
	} else {
		span.Status = sentry.SpanStatusOK
	}

	return &result, err
}

func (p *plaidClient) ExchangePublicToken(ctx context.Context, publicToken string) (*plaid.ExchangePublicTokenResponse, error) {
	span := sentry.StartSpan(ctx, "Plaid - ExchangePublicToken")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	result, err := p.client.ExchangePublicToken(publicToken)
	span.Data["plaidRequestId"] = result.RequestID
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to exchange public token")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (p *plaidClient) GetAccounts(ctx context.Context, accessToken string, options plaid.GetAccountsOptions) ([]plaid.Account, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetAccounts")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}
	span.Data["options"] = options

	result, err := p.client.GetAccountsWithOptions(accessToken, options)
	span.Data["plaidRequestId"] = result.RequestID
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve plaid accounts")
	}

	return result.Accounts, nil
}

func (p *plaidClient) GetAllTransactions(ctx context.Context, accessToken string, start, end time.Time, accountIds []string) ([]plaid.Transaction, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetAllTransactions")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.Data["start"] = start
	span.Data["end"] = start
	if len(accountIds) > 0 {
		span.Data["accountIds"] = accountIds
	}

	perPage := 100

	options := plaid.GetTransactionsOptions{
		StartDate:  start.Format("2006-01-02"),
		EndDate:    end.Format("2006-01-02"),
		AccountIDs: accountIds,
		Count:      perPage,
		Offset:     0,
	}

	transactions := make([]plaid.Transaction, 0)
	for {
		options.Offset = len(transactions)
		total, items, err := p.GetTransactions(span.Context(), accessToken, options)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, items...)

		if len(items) < perPage {
			break
		}

		if len(transactions) >= total {
			break
		}
	}

	return transactions, nil
}

func (p *plaidClient) GetTransactions(ctx context.Context, accessToken string, options plaid.GetTransactionsOptions) (total int, _ []plaid.Transaction, _ error) {
	span := sentry.StartSpan(ctx, "Plaid - GetTransactions")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.Data["options"] = options

	result, err := p.client.GetTransactionsWithOptions(accessToken, options)
	span.Data["plaidRequestId"] = result.RequestID
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to retrieve plaid transactions")
	}

	return result.TotalTransactions, result.Transactions, nil
}

func (p *plaidClient) GetInstitution(ctx context.Context, institutionId string, includeMetadata bool, countryCodes []string) (*plaid.Institution, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetInstitution")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.Data["institutionId"] = institutionId
	span.Data["includeMetadata"] = includeMetadata
	span.Data["countryCodes"] = countryCodes

	result, err := p.client.GetInstitutionByIDWithOptions(institutionId, countryCodes, plaid.GetInstitutionByIDOptions{
		IncludeOptionalMetadata:          true,
		IncludePaymentInitiationMetadata: false,
		IncludeStatus:                    false,
	})
	span.Data["plaidRequestId"] = result.RequestID
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve plaid institution")
	}

	return &result.Institution, nil
}

func (p *plaidClient) GetAllInstitutions(ctx context.Context, countryCodes []string, options plaid.GetInstitutionsOptions) ([]plaid.Institution, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetAllInstitutions")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}
	span.Data["countryCodes"] = countryCodes
	span.Data["options"] = options

	perPage := 500
	institutions := make([]plaid.Institution, 0)
	for {
		total, items, err := p.GetInstitutions(span.Context(), perPage, len(institutions), countryCodes, options)
		if err != nil {
			span.Status = sentry.SpanStatusInternalError
			return nil, err
		}

		institutions = append(institutions, items...)

		// If we received fewer items than we requested, then we have reached the end.
		if len(items) < perPage {
			break
		}

		// If we have received at least what we expect to be the total amount, then we are also done.
		if len(institutions) >= total {
			break
		}
	}

	return institutions, nil
}

func (p *plaidClient) GetInstitutions(ctx context.Context, count, offset int, countryCodes []string, options plaid.GetInstitutionsOptions) (total int, _ []plaid.Institution, _ error) {
	span := sentry.StartSpan(ctx, "Plaid - GetInstitutions")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.Data["count"] = count
	span.Data["offset"] = offset
	span.Data["countryCodes"] = countryCodes
	span.Data["options"] = options

	log := p.log.WithFields(logrus.Fields{
		"count":        count,
		"offset":       offset,
		"countryCodes": strings.Join(countryCodes, ","),
	})

	log.Debug("retrieving plaid institutions")

	rateLimitTimeout := time.NewTimer(30 * time.Second)
	select {
	// The institution ticker handles rate limiting for the get institutions endpoint. It makes sure that even
	// concurrently, we should not be able to exceed our request limit. At least on a single replica.
	case <-p.institutionTicker.C:
		result, err := p.client.GetInstitutionsWithOptions(count, offset, countryCodes, options)
		span.Data["plaidRequestId"] = result.RequestID
		log = log.WithField("plaidRequestId", result.RequestID)
		if err != nil {
			span.Status = sentry.SpanStatusInternalError
			log.WithError(err).Errorf("failed to retrieve plaid institutions")
			return 0, nil, errors.Wrap(err, "failed to retrieve plaid institutions")
		}

		log.Debugf("successfully retrieved %d institutions", len(result.Institutions))

		span.Status = sentry.SpanStatusOK

		return result.Total, result.Institutions, nil
	case <-rateLimitTimeout.C:
		return 0, nil, errors.Errorf("timed out waiting for rate limit")
	}
}

func (p *plaidClient) GetWebhookVerificationKey(ctx context.Context, keyId string) (plaid.GetWebhookVerificationKeyResponse, error) {
	span := sentry.StartSpan(ctx, "Plaid - GetWebhookVerificationKey")
	defer span.Finish()
	span.Data = map[string]interface{}{}

	result, err := p.client.GetWebhookVerificationKey(keyId)
	span.Data["plaidRequestId"] = result.RequestID
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
	} else {
		span.Status = sentry.SpanStatusOK
	}

	return result, errors.Wrap(err, "failed to retrieve webhook verification key")
}

func (p *plaidClient) Close() error {
	p.client = nil
	p.institutionTicker.Stop()
	return nil
}
