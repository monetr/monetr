package platypus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v14/plaid"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=platypus.go -package=mockgen -destination=../internal/mockgen/platypus.go Platypus
type (
	Platypus interface {
		CreateLinkToken(ctx context.Context, options LinkTokenOptions) (LinkToken, error)
		ExchangePublicToken(ctx context.Context, publicToken string) (*ItemToken, error)
		GetWebhookVerificationKey(ctx context.Context, keyId string) (*WebhookVerificationKey, error)
		GetInstitution(ctx context.Context, institutionId string) (*plaid.Institution, error)
		NewClientFromItemId(ctx context.Context, itemId string) (Client, error)
		NewClientFromLink(ctx context.Context, accountId uint64, linkId uint64) (Client, error)
		NewClient(ctx context.Context, link *models.Link, accessToken, itemId string) (Client, error)
		Close() error
	}
)

// after is a wrapper around some of the basic operations we would want to perform after each request. Mainly that we
// want to keep track of things like the request Id and some information about the request itself. It also handles error
// wrapping.
func after(span *sentry.Span, response *http.Response, err error, message, errorMessage string) error {
	if response != nil {
		requestId := response.Header.Get("X-Request-Id")
		if span.Data == nil {
			span.Data = map[string]interface{}{}
		}

		data := map[string]interface{}{}

		// With plaid responses we can actually still use the body of the response :tada:. The request Id is also stored on
		// the response body itself in most of my testing. I could have sworn the documentation cited X-Request-Id as being
		// a possible source for it, but I have not seen that yet. This bit of code extracts the body into a map. I know to
		// some degree of certainty that the response will always be an object and not an array. So a map with a string key
		// is safe. I can then extract the request Id and store that with my logging and diagnostic data.
		{
			var extractedResponseBody map[string]interface{}
			if e := json.NewDecoder(response.Body).Decode(&extractedResponseBody); e == nil {
				if requestId == "" {
					requestId = extractedResponseBody["request_id"].(string)
				}

				// But if our request was not successful, then I also want to yoink that body and throw it into my diagnostic
				// data as well. This will help me if I ever need to track down bugs with Plaid's API or problems with requests
				// that I am making incorrectly.
				if response.StatusCode != http.StatusOK {
					data["body"] = extractedResponseBody
				}
			}
		}

		{ // Make sure we put the request ID everywhere, this is easily the most important diagnostic data we need.
			data["X-RequestId"] = requestId
			span.Data["plaidRequestId"] = requestId
			span.SetTag("plaidRequestId", requestId)
		}

		crumbs.HTTP(
			span.Context(),
			message,
			"plaid",
			response.Request.URL.String(),
			response.Request.Method,
			response.StatusCode,
			data,
		)
	}

	switch e := err.(type) {
	case nil:
		span.Status = sentry.SpanStatusOK
	case plaid.GenericOpenAPIError:
		span.Status = sentry.SpanStatusInternalError
		var plaidError plaid.PlaidError
		if jsonErr := json.Unmarshal(e.Body(), &plaidError); jsonErr != nil {
			return errors.Wrap(err, errorMessage)
		}

		// Only include the plaid error message if it is provided.
		plaidMessage := ""
		if plaidError.ErrorMessage != "" {
			plaidMessage += " " + plaidError.ErrorMessage
		}

		return errors.Wrap(errors.Errorf(
			"plaid API call failed with [%s - %s]%s",
			plaidError.ErrorType, plaidError.ErrorCode, plaidMessage,
		), errorMessage)
	default:
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, errorMessage)
	}

	return nil
}

var (
	_ Platypus = &Plaid{}
)

func NewPlaid(log *logrus.Entry, secret secrets.PlaidSecretsProvider, repo repository.PlaidRepository, options config.Plaid) *Plaid {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	conf := plaid.NewConfiguration()
	conf.HTTPClient = httpClient
	conf.UseEnvironment(options.Environment)
	conf.AddDefaultHeader("PLAID-CLIENT-ID", options.ClientID)
	conf.AddDefaultHeader("PLAID-SECRET", options.ClientSecret)

	client := plaid.NewAPIClient(conf)

	return &Plaid{
		client: client,
		log:    log,
		secret: secret,
		repo:   repo,
		config: options,
	}
}

type Plaid struct {
	client *plaid.APIClient
	log    *logrus.Entry
	secret secrets.PlaidSecretsProvider
	repo   repository.PlaidRepository
	config config.Plaid
}

func (p *Plaid) CreateLinkToken(ctx context.Context, options LinkTokenOptions) (LinkToken, error) {
	span := sentry.StartSpan(ctx, "http.client")
	defer span.Finish()
	span.Description = "Plaid - CreateLinkToken"

	log := p.log.WithContext(span.Context())

	var redirectUri *string
	if p.config.OAuthDomain != "" {
		redirectUri = myownsanity.StringP(fmt.Sprintf("https://%s/plaid/oauth-return", p.config.OAuthDomain))
	}

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
				ClientUserId:             options.ClientUserID,
				LegalName:                &options.LegalName,
				PhoneNumber:              options.PhoneNumber,
				PhoneNumberVerifiedTime:  *plaid.NewNullableTime(options.PhoneNumberVerifiedTime),
				EmailAddress:             &options.EmailAddress,
				EmailAddressVerifiedTime: *plaid.NewNullableTime(options.EmailAddressVerifiedTime),
				Ssn:                      nil,
				DateOfBirth:              *plaid.NewNullableString(nil),
			},
			Products:              &consts.PlaidProducts,
			Webhook:               webhooksUrl,
			AccessToken:           nil,
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
	span := sentry.StartSpan(ctx, "http.client")
	defer span.Finish()
	span.Description = "Plaid - ExchangePublicToken"

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
	span := sentry.StartSpan(ctx, "http.client")
	defer span.Finish()
	span.Description = "Plaid - GetWebhookVerificationKey"

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
	span := sentry.StartSpan(ctx, "http.client")
	defer span.Finish()
	span.Description = "Plaid - GetInstitution"

	log := p.log.WithContext(span.Context())

	request := p.client.PlaidApi.
		InstitutionsGetById(span.Context()).
		InstitutionsGetByIdRequest(plaid.InstitutionsGetByIdRequest{
			InstitutionId: institutionId,
			CountryCodes:  consts.PlaidCountries,
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
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "Plaid - NewClientFromItemId"

	link, err := p.repo.GetLinkByItemId(span.Context(), itemId)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create client without link")
	}

	return p.newClient(span.Context(), link)
}

func (p *Plaid) NewClientFromLink(ctx context.Context, accountId uint64, linkId uint64) (Client, error) {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "Plaid - NewClientFromLink"

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
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "Plaid - newClient"

	if link == nil {
		return nil, errors.New("cannot create client without link")
	}

	if link.PlaidLink == nil {
		return nil, errors.New("cannot create client without link")
	}

	accessToken, err := p.secret.GetAccessTokenForPlaidLinkId(span.Context(), link.AccountId, link.PlaidLink.ItemId)
	if err != nil {
		return nil, err
	}

	return p.NewClient(span.Context(), link, accessToken, link.PlaidLink.ItemId)
}

func (p *Plaid) Close() error {
	panic("implement me")
}
