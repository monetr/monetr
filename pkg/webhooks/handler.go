package webhooks

import "context"

type (
	PlaidWebhook struct {
		WebhookType         string                 `json:"webhook_type"`
		WebhookCode         string                 `json:"webhook_code"`
		ItemId              string                 `json:"item_id"`
		Error               map[string]interface{} `json:"error"`
		NewWebhookURL       string                 `json:"new_webhook_url"`
		NewTransactions     int64                  `json:"new_transactions"`
		RemovedTransactions []string               `json:"removed_transactions"`
	}

	Dispatcher interface {
		Dispatch(ctx context.Context, webhook PlaidWebhook) error
	}
)
