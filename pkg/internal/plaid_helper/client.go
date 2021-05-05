package plaid_helper

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
)

type Client interface {
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
		log:    log,
		client: client,
	}
}

type plaidClient struct {
	log    *logrus.Entry
	client *plaid.Client
}

func (p *plaidClient) GetWebhookVerificationKey(ctx context.Context, keyId string) (plaid.GetWebhookVerificationKeyResponse, error) {
	span := sentry.StartSpan(ctx, "GetWebhookVerificationKey")
	defer span.Finish()

	result, err := p.client.GetWebhookVerificationKey(keyId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
	} else {
		span.Status = sentry.SpanStatusOK
	}

	return result, errors.Wrap(err, "failed to retrieve webhook verification key")
}

func (p *plaidClient) Close() error {
	p.client = nil
	return nil
}
