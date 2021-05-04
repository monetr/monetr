package controller

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"net/http"
	"strings"

	"github.com/kataras/iris/v12/context"
)

type PlaidWebhook struct {
	WebhookType         string                 `json:"webhook_type"`
	WebhookCode         string                 `json:"webhook_code"`
	ItemId              string                 `json:"item_id"`
	Error               map[string]interface{} `json:"error"`
	NewWebhookURL       string                 `json:"new_webhook_url"`
	NewTransactions     int64                  `json:"new_transactions"`
	RemovedTransactions []string               `json:"removed_transactions"`
}

func (c *Controller) handlePlaidWebhook(ctx *context.Context) {
	verification := ctx.GetHeader("Plaid-Verification")
	if strings.TrimSpace(verification) == "" {
		c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var kid string
	var claims jwt.StandardClaims
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

		verificationResponse, err := c.plaid.GetWebhookVerificationKey(kid)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve public verification key")
		}

		fmt.Sprint(verificationResponse)

		return nil, nil
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "unauthorized")
		return
	}

	fmt.Sprint(result)

	// TODO Properly verify webhooks from Plaid.

	var hook PlaidWebhook
	if err := ctx.ReadJSON(&hook); err != nil {
		c.badRequest(ctx, "malformed JSON")
		return
	}

	if err := c.processWebhook(hook); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to handle webhook")
		return
	}
}

func (c *Controller) processWebhook(hook PlaidWebhook) error {
	switch hook.WebhookType {
	case "TRANSACTIONS":
		switch hook.WebhookCode {
		case "INITIAL_UPDATE":

		}

	}

	return nil
}
