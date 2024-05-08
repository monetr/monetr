package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PlaidWebhook struct {
	WebhookType           string                 `json:"webhook_type"`
	WebhookCode           string                 `json:"webhook_code"`
	ItemId                string                 `json:"item_id"`
	Error                 map[string]interface{} `json:"error"`
	NewWebhookURL         string                 `json:"new_webhook_url"`
	NewTransactions       int64                  `json:"new_transactions"`
	RemovedTransactions   []string               `json:"removed_transactions"`
	ConsentExpirationTime *time.Time             `json:"consent_expiration_time"`
}

type PlaidClaims struct {
	jwt.StandardClaims
}

func (c PlaidClaims) Valid() error {
	vErr := new(jwt.ValidationError)
	now := jwt.TimeFunc().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if !c.VerifyExpiresAt(now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	// I'm adding 5 seconds onto the now timestamp because I was running into an issue periodically that the clock on
	// the server would be slightly different than the clock on Plaid's side. And the issued at was causing a conflict
	// where it was just a few seconds (sometimes just one) out of bounds for this to be handled. So adding a buffer of
	// 5 seconds to account for any clock drift between our servers and Plaid's.
	if !c.VerifyIssuedAt(now+5, false) {
		vErr.Inner = fmt.Errorf("token used before issued, %d | %d", now, c.IssuedAt)
		vErr.Errors |= jwt.ValidationErrorIssuedAt
	}

	if !c.VerifyNotBefore(now, false) {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= jwt.ValidationErrorNotValidYet
	}

	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}

func (c *Controller) postPlaidWebhook(ctx echo.Context) error {
	if !c.configuration.Plaid.Enabled || !c.configuration.Plaid.WebhooksEnabled {
		return c.notFound(ctx, "plaid webhooks are not enabled")
	}

	verification := ctx.Request().Header.Get("Plaid-Verification")
	if strings.TrimSpace(verification) == "" {
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var kid string
	var claims PlaidClaims

	result, err := jwt.ParseWithClaims(verification, &claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure the signing method for the JWT token is ES256 per Plaid's documentation. Anything else should be
		// rejected.
		method, ok := token.Method.(*jwt.SigningMethodECDSA)
		if !ok || method.Name != "ES256" {
			return nil, errors.Errorf("invalid signing method")
		}

		// Look for a kid field, we are going to use this to exchange for a public key that we can use to verify the
		// JWT token.
		value, ok := token.Header["kid"]
		if !ok {
			return nil, errors.Errorf("malformed JWT token, missing data")
		}

		// Make sure the value is a string, anything else is not valid and should be thrown out.
		kid, ok = value.(string)
		if !ok {
			return nil, errors.Errorf("malformed JWT token, expected string")
		}

		// Make sure that string has some kind of non-whitespace value.
		if strings.TrimSpace(kid) == "" {
			return nil, errors.Errorf("malformed JWT token, empty data")
		}

		log := c.log.WithField("kid", kid)
		log.Trace("exchanging key Id for public key")

		keyFunction, err := c.plaidWebhookVerification.GetVerificationKey(c.getContext(ctx), kid)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get verification key for webhook")
		}

		return keyFunction.Keyfunc(token)
	})
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusUnauthorized, "unauthorized")
	}

	if !result.Valid {
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var hook PlaidWebhook
	if err = ctx.Bind(&hook); err != nil {
		return c.badRequest(ctx, "malformed JSON")
	}

	if err = c.processWebhook(ctx, hook); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to handle webhook")
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) processWebhook(ctx echo.Context, hook PlaidWebhook) error {
	log := c.log.WithFields(logrus.Fields{
		"webhookType": hook.WebhookType,
		"webhookCode": hook.WebhookCode,
	})

	{
		fields := map[string]interface{}{
			"type":   hook.WebhookType,
			"code":   hook.WebhookCode,
			"itemId": hook.ItemId,
		}
		switch strings.ToUpper(fmt.Sprintf("%s.%s", hook.WebhookType, hook.WebhookCode)) {
		case "TRANSACTIONS.DEFAULT_UPDATE":
			fields["newTransactions"] = hook.NewTransactions
		case "TRANSACTIONS.TRANSACTIONS_REMOVED":
			fields["removedTransactions"] = hook.RemovedTransactions
		}
		crumbs.Debug(c.getContext(ctx), "Handling webhook from Plaid.", fields)
	}

	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("webhook", "plaid")
			scope.SetTag("plaid.item_id", hook.ItemId)
			scope.SetTag("plaid.webhook.type", hook.WebhookType)
			scope.SetTag("plaid.webhook.code", hook.WebhookCode)
		})
	}

	repo := c.mustGetUnauthenticatedRepository(ctx)

	log.Trace("retrieving link for webhook")
	link, err := repo.GetLinksForItem(c.getContext(ctx), hook.ItemId)
	if err != nil {
		crumbs.Error(c.getContext(ctx),
			"Failed to retrieve a link for the item Id provided by the Plaid webhook.",
			"plaid",
			map[string]interface{}{
				"itemId": hook.ItemId,
			},
		)
		log.WithError(err).Errorf("failed to retrieve link for item Id in webhook")
		return err
	}

	// Set the user for this webhook for sentry.
	crumbs.IncludeUserInScope(c.getContext(ctx), link.AccountId)
	crumbs.AddTag(c.getContext(ctx), "linkId", link.LinkId.String())

	if link.PlaidLink != nil {
		// If we have the plaid link in scope then add the institution ID onto the sentry scope.
		crumbs.AddTag(c.getContext(ctx), "plaid.institution_id", link.PlaidLink.InstitutionId)
	} else {
		// If we don't have it for some reason, indicate that there is a bug.
		crumbs.IndicateBug(c.getContext(ctx), "Plaid link should be in scope when retrieved by Plaid item ID", map[string]interface{}{
			"itemId": hook.ItemId,
			"link":   link,
		})
		// TODO Return here, we can't process it!
	}

	log = c.log.WithFields(logrus.Fields{
		"accountId": link.AccountId,
		"linkId":    link.LinkId,
	})

	log.Trace("processing webhook")

	authenticatedRepo := repository.NewRepositoryFromSession(
		c.clock,
		link.CreatedBy,
		link.AccountId,
		c.mustGetDatabase(ctx),
	)

	if hook.Error != nil {
		crumbs.Warn(c.getContext(ctx), "Webhook has an error", "plaid", hook.Error)
	}

	plaidLink := link.PlaidLink

	switch hook.WebhookType {
	case "TRANSACTIONS":
		err = background.TriggerSyncPlaid(c.getContext(ctx), c.jobRunner, background.SyncPlaidArguments{
			AccountId: link.AccountId,
			LinkId:    link.LinkId,
			Trigger:   "webhook",
		})
	case "ITEM":
		switch hook.WebhookCode {
		case "ERROR":
			code := hook.Error["error_code"]
			plaidLink.Status = models.PlaidLinkStatusError
			plaidLink.ErrorCode = myownsanity.StringP(code.(string))
			log.Warn("plaid link is in an error state, updating")
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "PENDING_EXPIRATION":
			plaidLink.Status = models.PlaidLinkStatusPendingExpiration
			plaidLink.ExpirationDate = hook.ConsentExpirationTime
			log.Warn("plaid link is pending expiration")
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "USER_PERMISSION_REVOKED":
			code := hook.Error["error_code"]
			plaidLink.Status = models.PlaidLinkStatusRevoked
			plaidLink.ErrorCode = myownsanity.StringP(code.(string))
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		case "WEBHOOK_UPDATE_ACKNOWLEDGED":
			err = background.TriggerSyncPlaid(c.getContext(ctx), c.jobRunner, background.SyncPlaidArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Trigger:   "webhook",
			})
		case "NEW_ACCOUNTS_AVAILABLE":
			plaidLink.NewAccountsAvailable = true
			err = authenticatedRepo.UpdatePlaidLink(c.getContext(ctx), plaidLink)
		default:
			crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
		}
	default:
		crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
	}

	return err
}
