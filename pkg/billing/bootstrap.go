package billing

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/internal/stripe_helper"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Bootstrapper interface {
	Provision(ctx context.Context) error
	Close() error
}

var (
	_ Bootstrapper = &stripeBootstrapper{}
)

type stripeBootstrapper struct {
	log           *logrus.Entry
	repo          repository.BillingRepository
	configuration config.Stripe
	stripe        stripe_helper.Stripe
}

func (s *stripeBootstrapper) Provision(ctx context.Context) error {
	if !s.configuration.Enabled {
		return errors.Errorf("stripe is not enabled")
	}
	span := sentry.StartSpan(ctx, "Billing - Provision")
	defer span.Finish()

	if len(s.configuration.Prices) == 0 {
		s.log.Warn("no prices configured, nothing to provision")
		return nil
	}

	s.log.Debugf("retrieving data for %d price(s)", len(s.configuration.Prices))

	prices, err := s.stripe.GetPricesById(span.Context(), s.configuration.Prices)
	if err != nil {
		s.log.WithError(err).Error("could not retrieve prices from stripe to provision billing")
		return err
	}

	if len(prices) == 0 {
		err = errors.Errorf("no prices returned, expected: %d", len(s.configuration.Prices))
		s.log.WithError(err).Error("no prices to provision billing with")
		return err
	}

	if len(prices) != len(s.configuration.Prices) {
		s.log.Warnf("price mismatch, requested: %d retrieved: %d", len(s.configuration.Prices), len(prices))
	}

	productIds := make([]string, len(prices))
	for i, price := range prices {
		productIds[i] = price.Product.ID
	}

	localProducts, err := s.repo.GetProductsByStripeProductId(span.Context(), productIds)
	if err != nil {
		s.log.WithError(err).Error("failed to query local products for provisioning billing")
		return err
	}

	localProductsByStripeProductId := map[string]models.Product{}
	for _, product := range localProducts {
		localProductsByStripeProductId[product.StripeProductId] = product
	}


	panic("implement me")
}

func (s *stripeBootstrapper) Close() error {
	panic("implement me")
}
