package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetr/rest-api/pkg/crumbs"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
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
	if c.VerifyExpiresAt(now, false) == false {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	// I'm adding 5 seconds onto the now timestamp because I was running into an issue periodically that the clock on
	// the server would be slightly different than the clock on Plaid's side. And the issued at was causing a conflict
	// where it was just a few seconds (sometimes just one) out of bounds for this to be handled. So adding a buffer of
	// 5 seconds to account for any clock drift between our servers and Plaid's.
	if c.VerifyIssuedAt(now+5, false) == false {
		vErr.Inner = fmt.Errorf("Token used before issued, %d | %d", now, c.IssuedAt)
		vErr.Errors |= jwt.ValidationErrorIssuedAt
	}

	if c.VerifyNotBefore(now, false) == false {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= jwt.ValidationErrorNotValidYet
	}

	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}

func (c *Controller) handlePlaidWebhook(ctx *context.Context) {
	verification := ctx.GetHeader("Plaid-Verification")
	if strings.TrimSpace(verification) == "" {
		c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
		return
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

		return keyFunction.KeyFuncF3T(token)
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "unauthorized")
		return
	}

	if !result.Valid {
		c.returnError(ctx, http.StatusForbidden, "unauthorized")
		return
	}

	var hook PlaidWebhook
	if err = ctx.ReadJSON(&hook); err != nil {
		c.badRequest(ctx, "malformed JSON")
		return
	}

	if err = c.processWebhook(ctx, hook); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to handle webhook")
		return
	}
}

func (c *Controller) processWebhook(ctx iris.Context, hook PlaidWebhook) error {
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
	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       strconv.FormatUint(link.AccountId, 10),
				Username: fmt.Sprintf("account:%d", link.AccountId),
			})
			scope.SetTag("accountId", strconv.FormatUint(link.AccountId, 10))
			scope.SetTag("linkId", strconv.FormatUint(link.LinkId, 10))
		})
	}

	log = c.log.WithFields(logrus.Fields{
		"accountId": link.AccountId,
		"linkId":    link.LinkId,
	})

	log.Trace("processing webhook")

	authenticatedRepo := repository.NewRepositoryFromSession(
		link.CreatedByUserId,
		link.AccountId,
		c.mustGetDatabase(ctx),
	)

	if hook.Error != nil {
		crumbs.Warn(c.getContext(ctx), "Webhook has an error", "plaid", hook.Error)
	}

	switch hook.WebhookType {
	case "TRANSACTIONS":
		switch hook.WebhookCode {
		case "INITIAL_UPDATE":
			_, err = c.job.TriggerPullInitialTransactions(link.AccountId, link.CreatedByUserId, link.LinkId)
		case "HISTORICAL_UPDATE":
			_, err = c.job.TriggerPullHistoricalTransactions(link.AccountId, link.LinkId)
		case "DEFAULT_UPDATE":
			_, err = c.job.TriggerPullLatestTransactions(link.AccountId, link.LinkId, hook.NewTransactions)
		case "TRANSACTIONS_REMOVED":
			_, err = c.job.TriggerRemoveTransactions(link.AccountId, link.LinkId, hook.RemovedTransactions)
		default:
			crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
		}
	case "ITEM":
		switch hook.WebhookCode {
		case "ERROR":
			code := hook.Error["error_code"]
			link.LinkStatus = models.LinkStatusError
			link.ErrorCode = myownsanity.StringP(code.(string))
			log.Warn("link is in an error state, updating")
			err = authenticatedRepo.UpdateLink(c.getContext(ctx), link)
		case "PENDING_EXPIRATION":
			link.LinkStatus = models.LinkStatusPendingExpiration
			link.ExpirationDate = hook.ConsentExpirationTime
			log.Warn("link is pending expiration")
			err = authenticatedRepo.UpdateLink(c.getContext(ctx), link)
		case "USER_PERMISSION_REVOKED":
			code := hook.Error["error_code"]
			link.LinkStatus = models.LinkStatusRevoked
			link.ErrorCode = myownsanity.StringP(code.(string))
			err = authenticatedRepo.UpdateLink(c.getContext(ctx), link)
		case "WEBHOOK_UPDATE_ACKNOWLEDGED":
			_, err = c.job.TriggerPullInitialTransactions(link.AccountId, link.CreatedByUserId, link.LinkId)
		default:
			crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
		}
	default:
		crumbs.Warn(c.getContext(ctx), "Plaid webhook will not be handled, it is not implemented.", "plaid", nil)
	}

	return err
}
