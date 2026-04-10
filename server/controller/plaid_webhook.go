package controller

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/datasources/plaid/plaid_jobs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type PlaidWebhook struct {
	WebhookType           string         `json:"webhook_type"`
	WebhookCode           string         `json:"webhook_code"`
	ItemId                string         `json:"item_id"`
	Error                 map[string]any `json:"error"`
	NewWebhookURL         string         `json:"new_webhook_url"`
	NewTransactions       int64          `json:"new_transactions"`
	RemovedTransactions   []string       `json:"removed_transactions"`
	ConsentExpirationTime *time.Time     `json:"consent_expiration_time"`
}

type PlaidClaims struct {
	jwt.RegisteredClaims
	RequestBodySHA256 string `json:"request_body_sha256"`
}

func (c *Controller) postPlaidWebhook(ctx echo.Context) error {
	if !c.Configuration.Plaid.Enabled || !c.Configuration.Plaid.WebhooksEnabled {
		return c.notFound(ctx, "plaid webhooks are not enabled")
	}

	verification := ctx.Request().Header.Get("Plaid-Verification")
	if strings.TrimSpace(verification) == "" {
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// Stream the request body through a SHA256 hasher while decoding the JSON
	// payload. This computes the body signature in a single pass without
	// buffering the entire request into memory. The decode happens before the
	// JWT verification so we can hash and parse in one read; the body size is
	// already capped by middleware so the decoder is not exposed to unbounded
	// untrusted input.
	hasher := sha256.New()
	tee := io.TeeReader(ctx.Request().Body, hasher)
	decoder := json.NewDecoder(tee)

	var hook PlaidWebhook
	if err := decoder.Decode(&hook); err != nil {
		return c.badRequest(ctx, "malformed JSON")
	}

	// The webhook body must be exactly one JSON value. Decoding into a fresh
	// value should hit io.EOF after skipping any trailing whitespace; anything
	// else means there is extra content (a second JSON value or garbage) which
	// we reject. This second Decode also pulls any remaining body bytes through
	// the tee, so the hasher sees the entire request payload Plaid signed.
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return c.badRequest(ctx, "malformed JSON")
	}

	signature := []byte(hex.EncodeToString(hasher.Sum(nil)))

	var kid string
	var claims PlaidClaims

	result, err := jwt.ParseWithClaims(
		verification,
		&claims,
		func(token *jwt.Token) (any, error) {
			// Make sure the signing method for the JWT token is ES256 per Plaid's
			// documentation. Anything else should be rejected.
			method, ok := token.Method.(*jwt.SigningMethodECDSA)
			if !ok || method.Name != "ES256" {
				return nil, errors.Errorf("invalid signing method")
			}

			// Look for a kid field, we are going to use this to exchange for a public
			// key that we can use to verify the JWT token.
			value, ok := token.Header["kid"]
			if !ok {
				return nil, errors.Errorf("malformed JWT token, missing data")
			}

			// Make sure the value is a string, anything else is not valid and should
			// be thrown out.
			kid, ok = value.(string)
			if !ok {
				return nil, errors.Errorf("malformed JWT token, expected string")
			}

			// Make sure that string has some kind of non-whitespace value.
			if strings.TrimSpace(kid) == "" {
				return nil, errors.Errorf("malformed JWT token, empty data")
			}

			log := c.Log.With("kid", kid)
			log.Log(c.getContext(ctx), logging.LevelTrace, "exchanging key Id for public key")

			keyFunction, err := c.PlaidWebhookVerification.GetVerificationKey(c.getContext(ctx), kid)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get verification key for webhook")
			}

			return keyFunction.Keyfunc(token)
		},
		// Enable iat validation and allow 5 seconds of leeway to account for
		// clock drift between our servers and Plaid's.
		jwt.WithIssuedAt(),
		jwt.WithLeeway(5*time.Second),
		jwt.WithTimeFunc(c.Clock.Now),
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusUnauthorized, "unauthorized")
	}

	if !result.Valid {
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if subtle.ConstantTimeCompare(
		signature,
		[]byte(strings.ToLower(claims.RequestBodySHA256)),
	) != 1 {
		c.getLog(ctx).ErrorContext(c.getContext(ctx), "received plaid request with valid token but invalid signature!", "expected", signature, "received", claims.RequestBodySHA256)
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	if err = c.processWebhook(ctx, hook); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to handle webhook")
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) processWebhook(ctx echo.Context, hook PlaidWebhook) error {
	log := c.getLog(ctx).With("webhookType", hook.WebhookType, "webhookCode", hook.WebhookCode, "itemId", hook.ItemId)

	crumbs.Debug(c.getContext(ctx), "Handling webhook from Plaid.", map[string]any{
		"webhook": hook,
	})

	crumbs.AddTag(c.getContext(ctx), "webhook", "plaid")
	crumbs.AddTag(c.getContext(ctx), "plaid.item_id", hook.ItemId)
	crumbs.AddTag(c.getContext(ctx), "plaid.webhook.type", hook.WebhookType)
	crumbs.AddTag(c.getContext(ctx), "plaid.webhook.code", hook.WebhookCode)

	repo := c.mustGetUnauthenticatedRepository(ctx)

	log.Log(c.getContext(ctx), logging.LevelTrace, "retrieving link for webhook")
	link, err := repo.GetLinksForItem(c.getContext(ctx), hook.ItemId)
	if err != nil {
		crumbs.Error(c.getContext(ctx),
			"Failed to retrieve a link for the item Id provided by the Plaid webhook.",
			"plaid",
			map[string]any{
				"itemId": hook.ItemId,
			},
		)

		// If the link is not even in the database then there is nothing to be done.
		if errors.Is(err, pg.ErrNoRows) {
			log.WarnContext(c.getContext(ctx), "link is not in database, webhook cannot be handled", "err", err)
			return nil
		}

		log.ErrorContext(c.getContext(ctx), "failed to retrieve link for item Id in webhook", "err", err)
		return err
	}

	// Set the user for this webhook for sentry.
	crumbs.IncludeUserInScope(c.getContext(ctx), link.AccountId)
	crumbs.AddTag(c.getContext(ctx), "linkId", link.LinkId.String())
	crumbs.AddTag(c.getContext(ctx), "plaid.institution_id", link.PlaidLink.InstitutionId)
	crumbs.AddTag(c.getContext(ctx), "plaid.institution_name", link.PlaidLink.InstitutionName)

	if link.PlaidLink != nil {
		// If we have the plaid link in scope then add the institution ID onto the sentry scope.
		crumbs.AddTag(c.getContext(ctx), "plaid.institution_id", link.PlaidLink.InstitutionId)
	} else {
		// If we don't have it for some reason, indicate that there is a bug.
		crumbs.IndicateBug(c.getContext(ctx), "Plaid link should be in scope when retrieved by Plaid item ID", map[string]any{
			"itemId": hook.ItemId,
			"link":   link,
		})
		log.WarnContext(c.getContext(ctx), "Plaid link should be in scope when retrieved by Plaid item ID")
		return c.returnError(ctx, http.StatusFailedDependency, "failed to find record for plaid link")
	}

	log = log.With("accountId", link.AccountId, "linkId", link.LinkId)

	log.InfoContext(c.getContext(ctx), "processing Plaid webhook")

	authenticatedRepo := repository.NewRepositoryFromSession(
		c.Clock,
		link.CreatedBy,
		link.AccountId,
		c.mustGetDatabase(ctx),
		log,
	)

	if hook.Error != nil {
		crumbs.Warn(c.getContext(ctx), "Webhook has an error", "plaid", hook.Error)
	}

	plaidLink := link.PlaidLink

	switch hook.WebhookType {
	case "TRANSACTIONS":
		switch hook.WebhookCode {
		case "SYNC_UPDATES_AVAILABLE", "INITIAL_UPDATE", "HISTORICAL_UPDATE":
			err = enqueueJob(
				c,
				ctx,
				plaid_jobs.SyncPlaid,
				plaid_jobs.SyncPlaidArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "webhook",
				},
			)
		case "RECURRING_TRANSACTIONS_UPDATE":
			log.WarnContext(c.getContext(ctx), "received a recurring transaction update webhook, monetr does nothing with these events at this time")
		default:
			log.DebugContext(c.getContext(ctx), "ignoring Plaid webhook to avoid double syncing")
		}
	case "ITEM":
		switch hook.WebhookCode {
		case "LOGIN_REPAIRED":
			plaidLink.Status = models.PlaidLinkStatusSetup
			plaidLink.ErrorCode = nil
			plaidLink.ExpirationDate = nil
			log.InfoContext(c.getContext(ctx), "plaid link has been repaired")
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "ERROR":
			code := hook.Error["error_code"]
			plaidLink.Status = models.PlaidLinkStatusError
			plaidLink.ErrorCode = myownsanity.Pointer(code.(string))
			log.WarnContext(c.getContext(ctx), "plaid link is in an error state, updating")
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "PENDING_EXPIRATION":
			plaidLink.Status = models.PlaidLinkStatusPendingExpiration
			plaidLink.ExpirationDate = hook.ConsentExpirationTime
			log.WarnContext(c.getContext(ctx), "plaid link is pending expiration")
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "USER_PERMISSION_REVOKED", "USER_ACCOUNT_REVOKED":
			code := hook.Error["error_code"]
			plaidLink.Status = models.PlaidLinkStatusRevoked
			plaidLink.ErrorCode = myownsanity.Pointer(code.(string))
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "WEBHOOK_UPDATE_ACKNOWLEDGED":
			err = enqueueJob(
				c,
				ctx,
				plaid_jobs.SyncPlaid,
				plaid_jobs.SyncPlaidArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "webhook",
				},
			)
		case "NEW_ACCOUNTS_AVAILABLE":
			plaidLink.NewAccountsAvailable = true
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		default:
			log.WarnContext(c.getContext(ctx), "plaid webhook will not be handled, it is not implemented")
			crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
		}
	default:
		log.WarnContext(c.getContext(ctx), "plaid webhook will not be handled, it is not implemented")
		crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
	}

	return err
}
