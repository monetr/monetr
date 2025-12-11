package platypus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/round"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/sirupsen/logrus"
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
func after(span *sentry.Span, response *http.Response, err error, message, errorMessage string) error {
	// if response != nil {
	// 	requestId := response.Header.Get("X-Request-Id")
	// 	if span.Data == nil {
	// 		span.Data = map[string]any{}
	// 	}
	// 	span.Description = fmt.Sprintf(
	// 		"%s %s",
	// 		response.Request.Method,
	// 		response.Request.URL.String(),
	// 	)
	//
	// 	data := map[string]any{}
	//
	// 	// With plaid responses we can actually still use the body of the response :tada:. The request Id is also stored on
	// 	// the response body itself in most of my testing. I could have sworn the documentation cited X-Request-Id as being
	// 	// a possible source for it, but I have not seen that yet. This bit of code extracts the body into a map. I know to
	// 	// some degree of certainty that the response will always be an object and not an array. So a map with a string key
	// 	// is safe. I can then extract the request Id and store that with my logging and diagnostic data.
	// 	{
	// 		var extractedResponseBody map[string]any
	// 		if e := json.NewDecoder(response.Body).Decode(&extractedResponseBody); e == nil {
	// 			if requestId == "" {
	// 				requestId = extractedResponseBody["request_id"].(string)
	// 			}
	//
	// 			// But if our request was not successful, then I also want to yoink that body and throw it into my diagnostic
	// 			// data as well. This will help me if I ever need to track down bugs with Plaid's API or problems with requests
	// 			// that I am making incorrectly.
	// 			if response.StatusCode != http.StatusOK {
	// 				data["body"] = extractedResponseBody
	// 			}
	// 		}
	// 	}
	//
	// 	{ // Make sure we put the request ID everywhere, this is easily the most important diagnostic data we need.
	// 		data["X-RequestId"] = requestId
	// 		span.Data["plaidRequestId"] = requestId
	// 		span.SetTag("plaidRequestId", requestId)
	// 		span.SetTag("http.request.method", response.Request.Method)
	// 		span.SetTag("server.address", response.Request.URL.Hostname())
	// 		span.SetTag("url.full", response.Request.URL.String())
	// 		span.SetTag("http.response.status_code", fmt.Sprint(response.StatusCode))
	// 	}
	//
	// 	crumbs.HTTP(
	// 		span.Context(),
	// 		message,
	// 		"plaid",
	// 		response.Request.URL.String(),
	// 		response.Request.Method,
	// 		response.StatusCode,
	// 		data,
	// 	)
	// }

	switch e := err.(type) {
	case nil:
		span.Status = sentry.SpanStatusOK
	case plaid.GenericOpenAPIError:
		span.Status = sentry.SpanStatusInternalError
		var plaidError plaid.PlaidError
		if jsonErr := json.Unmarshal(e.Body(), &plaidError); jsonErr != nil {
			return errors.Wrap(err, errorMessage)
		}

		return errors.Wrap(
			&PlatypusError{plaidError},
			errorMessage,
		)
	default:
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, errorMessage)
	}

	return nil
}

var (
	_ Platypus = &Plaid{}
)

func NewPlaid(
	log *logrus.Entry,
	clock clock.Clock,
	kms secrets.KeyManagement,
	db pg.DBI,
	options config.Plaid,
) *Plaid {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: round.NewObservabilityRoundTripper(http.DefaultTransport,
			func(
				ctx context.Context,
				request *http.Request,
				response *http.Response,
				err error,
			) {
				requestLog := log.WithContext(ctx).WithFields(logrus.Fields{
					"plaid_method": request.Method,
					"plaid_url":    request.URL.String(),
				})
				var statusCode int
				var requestId string
				if response != nil {
					statusCode = response.StatusCode
					requestLog = requestLog.WithField("plaid_statusCode", statusCode)

					var responseData map[string]any
					buffer := bytes.NewBuffer(nil)
					tee := io.TeeReader(response.Body, buffer)
					if err := json.NewDecoder(tee).Decode(&responseData); err != nil {
						log.WithError(err).Warn("failed to decode plaid response as json for logging")
					} else if id, ok := responseData["request_id"].(string); ok {
						requestId = id
						requestLog = requestLog.WithField("plaid_requestId", requestId)
					}

					// Close the existing body before we replace it.
					response.Body.Close()
					// Then swap out the body
					response.Body = io.NopCloser(buffer)
				}

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
				requestLog.Debug("Plaid API call")
			}),
	}

	conf := plaid.NewConfiguration()
	conf.HTTPClient = httpClient
	conf.UseEnvironment(options.Environment)
	conf.AddDefaultHeader("PLAID-CLIENT-ID", options.ClientID)
	conf.AddDefaultHeader("PLAID-SECRET", options.ClientSecret)

	client := plaid.NewAPIClient(conf)

	return &Plaid{
		clock:  clock,
		client: client,
		db:     db,
		log:    log,
		kms:    kms,
		repo:   repository.NewPlaidRepository(db),
		config: options,
	}
}

type Plaid struct {
	clock  clock.Clock
	client *plaid.APIClient
	db     pg.DBI
	log    *logrus.Entry
	kms    secrets.KeyManagement
	repo   repository.PlaidRepository
	config config.Plaid
}

func (p *Plaid) CreateLinkToken(ctx context.Context, options LinkTokenOptions) (LinkToken, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log.WithContext(span.Context())

	var redirectUri *string
	if p.config.OAuthDomain != "" {
		redirectUri = myownsanity.Pointer(fmt.Sprintf("https://%s/plaid/oauth-return", p.config.OAuthDomain))
	}

	var webhooksUrl *string
	if p.config.WebhooksEnabled {
		if p.config.WebhooksDomain == "" {
			crumbs.Warn(span.Context(), "BUG: Plaid webhook domain is not present but webhooks are enabled.", "bug", nil)
		} else {
			webhooksUrl = myownsanity.Pointer(p.config.GetWebhooksURL())
		}
	}

	request := p.client.PlaidApi.
		LinkTokenCreate(span.Context()).
		LinkTokenCreateRequest(plaid.LinkTokenCreateRequest{
			ClientName:   consts.PlaidClientName,
			Language:     consts.PlaidLanguage,
			CountryCodes: p.config.CountryCodes,
			User: plaid.LinkTokenCreateRequestUser{
				ClientUserId:             options.ClientUserID,
				LegalName:                &options.LegalName,
				PhoneNumber:              options.PhoneNumber,
				PhoneNumberVerifiedTime:  *plaid.NewNullableTime(options.PhoneNumberVerifiedTime),
				EmailAddress:             &options.EmailAddress,
				EmailAddressVerifiedTime: *plaid.NewNullableTime(options.EmailAddressVerifiedTime),
				Ssn:                      nil,
				DateOfBirth:              *plaid.NewNullableString(nil),
			},
			Products:              consts.PlaidProducts,
			Webhook:               webhooksUrl,
			AccessToken:           *plaid.NewNullableString(nil),
			LinkCustomizationName: nil,
			RedirectUri:           redirectUri,
			AndroidPackageName:    nil,
			AccountFilters:        nil,
			EuConfig:              nil,
			InstitutionId:         nil,
			PaymentInitiation:     nil,
			DepositSwitch:         nil,
			IncomeVerification:    nil,
			Auth:                  nil,
			Transactions: &plaid.LinkTokenTransactions{
				DaysRequested: myownsanity.Pointer[int32](2 * 365), // 2 years
			},
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Creating link token with Plaid",
		"failed to create link token",
	); err != nil {
		log.WithError(err).Errorf("failed to create link token")
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

	log := p.log.WithContext(span.Context())

	request := p.client.PlaidApi.
		ItemPublicTokenExchange(span.Context()).
		ItemPublicTokenExchangeRequest(plaid.ItemPublicTokenExchangeRequest{
			PublicToken: publicToken,
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Exchanging public token with Plaid",
		"failed to exchange public token with Plaid",
	); err != nil {
		log.WithError(err).Errorf("failed to exchange public token with Plaid")
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

	log := p.log.WithContext(span.Context())

	request := p.client.PlaidApi.
		WebhookVerificationKeyGet(span.Context()).
		WebhookVerificationKeyGetRequest(plaid.WebhookVerificationKeyGetRequest{
			KeyId: keyId,
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Retrieving webhook verification key",
		"failed to retrieve webhook verification key from Plaid",
	); err != nil {
		log.WithError(err).Errorf("failed to retrieve webhook verification key from Plaid")
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

	log := p.log.WithContext(span.Context())

	request := p.client.PlaidApi.
		InstitutionsGetById(span.Context()).
		InstitutionsGetByIdRequest(plaid.InstitutionsGetByIdRequest{
			InstitutionId: institutionId,
			CountryCodes:  p.config.CountryCodes,
			Options: &plaid.InstitutionsGetByIdRequestOptions{
				IncludeOptionalMetadata:          myownsanity.BoolP(true),
				IncludeStatus:                    myownsanity.BoolP(true),
				IncludeAuthMetadata:              myownsanity.BoolP(false),
				IncludePaymentInitiationMetadata: myownsanity.BoolP(false),
			},
		})

	result, response, err := request.Execute()
	if err = after(
		span,
		response,
		err,
		"Retrieving Plaid institution status",
		"failed to retrieve Plaid institution status",
	); err != nil {
		log.WithError(err).Errorf("failed to retrieve Plaid institution status")
		return nil, err
	}

	return &result.Institution, nil
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

func (p *Plaid) NewClient(ctx context.Context, link *models.Link, accessToken, itemId string) (Client, error) {
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
		log: p.log.WithFields(logrus.Fields{
			"accountId": link.AccountId,
			"linkId":    link.LinkId,
			"itemId":    itemId,
		}),
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

func (p *Plaid) Close() error {
	panic("implement me")
}
