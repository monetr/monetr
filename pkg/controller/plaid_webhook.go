package controller

import (
	"net/http"
	"net/url"
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

	// TODO Properly verify webhooks from Plaid.

	var hook PlaidWebhook
	if err := ctx.ReadJSON(&hook); err != nil {
		c.badRequest(ctx, "malformed JSON")
		return
	}

}

func (c *Controller) getWebhookUrl() string {
	if !c.configuration.EnableWebhooks {
		return ""
	}

	uri, err := url.Parse(c.configuration.APIDomainName)
	if err != nil {
		panic(err)
	}

	return uri.String()
}
