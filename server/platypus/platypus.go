package platypus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/round"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	plaid "github.com/plaid/plaid-go/v49"
	"github.com/plaid/plaid-go/v49/client"
	"github.com/plaid/plaid-go/v49/option"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uptrace/bun"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=platypus.go -package=mockgen -destination=../internal/mockgen/platypus.go Platypus
type (
	Platypus interface {
		CreateLinkToken(ctx context.Context, options LinkTokenOptions) (LinkToken, error)
		ExchangePublicToken(ctx context.Context, publicToken string) (*ItemToken, error)
		GetWebhookVerificationKey(ctx context.Context, keyId string) (*WebhookVerificationKey, error)
		GetInstitution(ctx context.Context, institutionId string) (*plaid.Institution, error)
		NewClientFromItemId(ctx context.Context, itemId string) (Client, error)
		NewClientFromLink(ctx context.Context, accountId models.ID[models.Account], linkId models.ID[models.Link]) (Client, error)
		NewClient(ctx context.Context, link *models.Link, accessToken, itemId string) (Client, error)
		Close() error
	}
)

// after is a wrapper around some of the basic operations we would want to perform after each request. Mainly that we
// want to keep track of things like the request Id and some information about the request itself. It also handles error
// wrapping.
func after(span *sentry.Span, err error, _, errorMessage string) error {
	if err == nil {
		span.Status = sentry.SpanStatusOK
		return nil
	}

	span.Status = sentry.SpanStatusInternalError

	// Plaid returns a different error type for each status code, but they all
	// have the parsed Plaid error on them. If the body wasn't a Plaid error then
	// this is nil and we just wrap whatever we got.
	var plaidErr interface{ GetBody() *plaid.PlaidError }
	if errors.As(err, &plaidErr) && plaidErr.GetBody() != nil {
		return errors.Wrap(
			&PlatypusError{*plaidErr.GetBody()},
			errorMessage,
		)
	}

	return errors.Wrap(err, errorMessage)
}

// deref will return the value the pointer points to, or the zero value if the
// pointer is nil. Plaid's SDK gives us pointers for anything nullable.
func deref[T any](input *T) T {
	var result T
	if input != nil {
		result = *input
	}

	return result
}

var (
	_ Platypus = &Plaid{}
)

func NewPlaid(
	log *slog.Logger,
	clock clock.Clock,
	kms secrets.KeyManagement,
	db bun.IDB,
	options config.Plaid,
) *Plaid {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: round.NewObservabilityRoundTripper(http.DefaultTransport,
			func(
				ctx context.Context,
				request *http.Request,
				response *http.Response,
				_ error,
				duration time.Duration,
			) {
				requestLog := log.With(
					"plaid_method", request.Method,
					"plaid_url", request.URL.String(),
				)
				var statusCode int
				var requestId string
				// Plaid endpoints are all fixed paths like /transactions/sync so this
				// is safe to use as a label.
				endpoint := request.URL.Path
				// If we never got a response then the request didn't make it to Plaid
				// at all, like a timeout.
				status := "error"
				if response != nil {
					statusCode = response.StatusCode
					status = strconv.Itoa(statusCode)
					requestLog = requestLog.With("plaid_statusCode", statusCode)

					var responseData map[string]any
					buffer := bytes.NewBuffer(nil)
					tee := io.TeeReader(response.Body, buffer)
					if err := json.NewDecoder(tee).Decode(&responseData); err != nil {
						log.WarnContext(ctx, "failed to decode plaid response as json for logging", "err", err)
					} else {
						if id, ok := responseData["request_id"].(string); ok {
							requestId = id
							requestLog = requestLog.With("plaid_requestId", requestId)
						}
						if errorCode, ok := responseData["error_code"].(string); ok && errorCode != "" {
							plaidApiErrors.With(prometheus.Labels{
								"endpoint":   endpoint,
								"error_code": errorCode,
							}).Inc()
						}
					}

					// Close the existing body before we replace it.
					response.Body.Close()
					// Then swap out the body
					response.Body = io.NopCloser(buffer)
				}

				plaidApiRequests.With(prometheus.Labels{
					"endpoint": endpoint,
					"status":   status,
				}).Inc()
				plaidApiRequestTime.With(prometheus.Labels{
					"endpoint": endpoint,
				}).Observe(float64(duration.Milliseconds()))

				// If you get a nil reference panic here during testing, its probably because you forgot to mock a certain endpoint.
				// Check to see if the error is a "no responder found" error.
				crumbs.HTTP(ctx,
					"Plaid API Call",
					"plaid",
					request.URL.String(),
					request.Method,
					statusCode,
					map[string]any{
						"Request-Id": requestId,
					},
				)
				requestLog.DebugContext(ctx, "Plaid API call")
			}),
	}

	return &Plaid{
		clock: clock,
		client: client.NewClient(
			option.WithHTTPClient(httpClient),
			option.WithBaseURL(options.Environment),
			option.WithClientID(new(options.ClientID)),
			option.WithSecret(options.ClientSecret),
		),
		db:     db,
		log:    log,
		kms:    kms,
		repo:   repository.NewPlaidRepository(db),
		config: options,
	}
}

type Plaid struct {
	clock  clock.Clock
	client *client.Client
	db     bun.IDB
	log    *slog.Logger
	kms    secrets.KeyManagement
	repo   repository.PlaidRepository
	config config.Plaid
}

func (p *Plaid) CreateLinkToken(ctx context.Context, options LinkTokenOptions) (LinkToken, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log

	var redirectUri *string
	if p.config.OAuthDomain != "" {
		redirectUri = new(fmt.Sprintf("https://%s/plaid/oauth-return", p.config.OAuthDomain))
	}

	var webhooksUrl *string
	if p.config.WebhooksEnabled {
		if p.config.WebhooksDomain == "" {
			crumbs.Warn(span.Context(), "BUG: Plaid webhook domain is not present but webhooks are enabled.", "bug", nil)
		} else {
			webhooksUrl = new(p.config.GetWebhooksURL())
		}
	}

	result, err := p.client.Link.CreateToken(
		span.Context(),
		&plaid.LinkTokenCreateRequest{
			ClientName:   consts.PlaidClientName,
			Language:     consts.PlaidLanguage,
			CountryCodes: p.config.CountryCodes,
			User: &plaid.LinkTokenCreateRequestUser{
				ClientUserID:             options.ClientUserID,
				LegalName:                &options.LegalName,
				PhoneNumber:              options.PhoneNumber,
				PhoneNumberVerifiedTime:  options.PhoneNumberVerifiedTime,
				EmailAddress:             &options.EmailAddress,
				EmailAddressVerifiedTime: options.EmailAddressVerifiedTime,
				Ssn:                      nil,
				DateOfBirth:              nil,
			},
			Products:              consts.PlaidProducts,
			Webhook:               webhooksUrl,
			AccessToken:           nil,
			LinkCustomizationName: nil,
			RedirectURI:           redirectUri,
			AndroidPackageName:    nil,
			AccountFilters:        nil,
			EuConfig:              nil,
			InstitutionID:         nil,
			PaymentInitiation:     nil,
			IncomeVerification:    nil,
			Auth:                  nil,
			Transactions: &plaid.LinkTokenTransactions{
				DaysRequested: new(2 * 365), // 2 years
			},
		},
	)
	if err = after(
		span,
		err,
		"Creating link token with Plaid",
		"failed to create link token",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to create link token", "err", err)
		return nil, err
	}

	return PlaidLinkToken{
		LinkToken: result.LinkToken,
		Expires:   result.Expiration,
	}, nil
}

func (p *Plaid) ExchangePublicToken(ctx context.Context, publicToken string) (*ItemToken, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log

	result, err := p.client.Item.ExchangePublicToken(
		span.Context(),
		&plaid.ItemPublicTokenExchangeRequest{
			PublicToken: publicToken,
		},
	)
	if err = after(
		span,
		err,
		"Exchanging public token with Plaid",
		"failed to exchange public token with Plaid",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to exchange public token with Plaid", "err", err)
		return nil, err
	}

	token, err := NewItemTokenFromPlaid(result)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (p *Plaid) GetWebhookVerificationKey(ctx context.Context, keyId string) (*WebhookVerificationKey, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log

	result, err := p.client.Webhook.GetVerificationKey(
		span.Context(),
		&plaid.WebhookVerificationKeyGetRequest{
			KeyID: keyId,
		},
	)
	if err = after(
		span,
		err,
		"Retrieving webhook verification key",
		"failed to retrieve webhook verification key from Plaid",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve webhook verification key from Plaid", "err", err)
		return nil, err
	}

	webhook, err := NewWebhookVerificationKeyFromPlaid(result.Key)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

func (p *Plaid) GetInstitution(ctx context.Context, institutionId string) (*plaid.Institution, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log

	result, err := p.client.Institutions.GetByID(
		span.Context(),
		&plaid.InstitutionsGetByIDRequest{
			InstitutionID: institutionId,
			CountryCodes:  p.config.CountryCodes,
			Options: &plaid.InstitutionsGetByIDRequestOptions{
				IncludeOptionalMetadata:          new(true),
				IncludeStatus:                    new(true),
				IncludeAuthMetadata:              new(false),
				IncludePaymentInitiationMetadata: new(false),
			},
		},
	)
	if err = after(
		span,
		err,
		"Retrieving Plaid institution status",
		"failed to retrieve Plaid institution status",
	); err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve Plaid institution status", "err", err)
		return nil, err
	}

	return result.Institution, nil
}

func (p *Plaid) NewClientFromItemId(ctx context.Context, itemId string) (Client, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link, err := p.repo.GetLinkByItemId(span.Context(), itemId)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create client without link")
	}

	return p.newClient(span.Context(), link)
}

func (p *Plaid) NewClientFromLink(ctx context.Context, accountId models.ID[models.Account], linkId models.ID[models.Link]) (Client, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link, err := p.repo.GetLink(span.Context(), accountId, linkId)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Plaid client from link")
	}

	return p.newClient(span.Context(), link)
}

func (p *Plaid) NewClient(_ context.Context, link *models.Link, accessToken, itemId string) (Client, error) {
	if accessToken == "" {
		return nil, errors.New("plaid access token is required to create a client")
	}

	if itemId == "" {
		return nil, errors.New("plaid itemId is required to create a client")
	}

	return &PlaidClient{
		accountId:   link.AccountId,
		linkId:      link.LinkId,
		accessToken: accessToken,
		itemId:      itemId,
		log: p.log.With(
			"accountId", link.AccountId,
			"linkId", link.LinkId,
			"itemId", itemId,
		),
		client: p.client,
		config: p.config,
	}, nil
}

func (p *Plaid) newClient(ctx context.Context, link *models.Link) (Client, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if link == nil {
		return nil, errors.New("cannot create client without link")
	}

	if link.PlaidLink == nil {
		return nil, errors.New("cannot create client without link")
	}

	plaidLink := link.PlaidLink
	secretRepo := repository.NewSecretsRepository(
		p.log,
		p.clock,
		p.db,
		p.kms,
		plaidLink.AccountId,
	)
	secret, err := secretRepo.Read(
		span.Context(),
		plaidLink.SecretId,
	)
	if err != nil {
		return nil, err
	}

	return p.NewClient(span.Context(), link, secret.Value, link.PlaidLink.PlaidId)
}

func (*Plaid) Close() error {
	panic("implement me")
}
