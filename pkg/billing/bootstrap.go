package billing

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/internal/stripe_helper"
)

type Bootstrapper interface {
	Provision(ctx context.Context) error
	Close() error
}

var (
	_ Bootstrapper = &stripeBootstrapper{}
)

type stripeBootstrapper struct {
	db            pg.DBI
	configuration config.Stripe
	stripe        stripe_helper.Stripe
}

func (s *stripeBootstrapper) Provision(ctx context.Context) error {
	panic("implement me")
}

func (s *stripeBootstrapper) Close() error {
	panic("implement me")
}
