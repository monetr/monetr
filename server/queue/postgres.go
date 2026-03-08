package queue

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/storage"
)

var (
	_ Processor = &postgresProcessor{}
)

type postgresProcessor struct {
	log           *slog.Logger
	clock         clock.Clock
	configuration config.Configuration
	db            *pg.DB
	publisher     pubsub.Publisher
	plaidPlatypus platypus.Platypus
	kms           secrets.KeyManagement
	fileStorage   storage.Storage
	billing       billing.Billing
	email         communication.EmailCommunication
}

// enqueue implements [Processor].
func (p *postgresProcessor) enqueue(
	ctx context.Context,
	queue string,
	args any,
) error {
	panic("unimplemented")
}

// register implements [Processor].
func (p *postgresProcessor) register(
	ctx context.Context,
	queue string,
	job internalJobWrapper,
) error {
	panic("unimplemented")
}

// registerCron implements [Processor].
func (p *postgresProcessor) registerCron(
	ctx context.Context,
	queue string,
	schedule string,
	job internalJobWrapper,
) error {
	panic("unimplemented")
}
