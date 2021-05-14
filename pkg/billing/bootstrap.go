package billing

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/getsentry/sentry-go"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/feature"
	"github.com/monetrapp/rest-api/pkg/internal/stripe_helper"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
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

	log := s.log

	// If the configuration doesn't have any stripePrices then there is nothing to be done.
	if len(s.configuration.Prices) == 0 {
		s.log.Warn("no prices configured, nothing to provision")
		return nil
	}

	log.Debugf("retrieving data for %d price(s)", len(s.configuration.Prices))

	// Retrieve the prices in the configuration from stripe so we have the most up to date data.
	stripePrices, err := s.stripe.GetPricesById(span.Context(), s.configuration.Prices)
	if err != nil {
		log.WithError(err).Error("could not retrieve stripePrices from stripe to provision billing")
		return err
	}

	// If no stripePrices were retrieved from stripe then something is fucky.
	if len(stripePrices) == 0 {
		err = errors.Errorf("no stripePrices returned, expected: %d", len(s.configuration.Prices))
		log.WithError(err).Error("no stripePrices to provision billing with")
		return err
	}

	// If we did get something, but not everything (or maybe too much?) then just throw a warning into the logs.
	if len(stripePrices) != len(s.configuration.Prices) {
		log.Warnf("price mismatch, requested: %d retrieved: %d", len(s.configuration.Prices), len(stripePrices))
	}

	// From our stripePrices we want to get a list of productIds.
	stripeProductIds := make([]string, 0)
	linq.From(stripePrices).
		SelectT(func(price stripe.Price) string {
			return price.Product.ID
		}).
		Distinct().
		ToSlice(&stripeProductIds)

	log.Debugf("retrieving %d product(s) from stripe", len(stripeProductIds))
	stripeProducts, err := s.stripe.GetProductsById(span.Context(), stripeProductIds)
	if err != nil {
		log.WithError(err).Error("failed to retrieve product(s) from stripe")
		return err
	}

	stripeProductsById := map[string]stripe.Product{}
	for _, product := range stripeProducts {
		stripeProductsById[product.ID] = product
	}

	// See what products we already have locally.
	existingProducts, err := s.repo.GetProductsByStripeProductId(span.Context(), stripeProductIds)
	if err != nil {
		s.log.WithError(err).Error("failed to query local products for provisioning billing")
		return err
	}

	existingProductsByStripeId := map[string]models.Product{}
	for _, product := range existingProducts {
		existingProductsByStripeId[product.StripeProductId] = product
	}

	productsToCreate := make([]models.Product, 0)
	productsToUpdate := make([]models.Product, 0)
	pricesToUpdate := make([]models.Price, 0)
	pricesToCreate := make([]models.Price, 0)
	// I'm not going to deny, this is some of the most fucked up shit. I'm very sorry. I have no idea what I'm doing
	// I hate writing billing code.
	linq.From(stripePrices).
		// Group all the stripePrices by their product Id.
		GroupByT(
			func(price stripe.Price) string {
				return price.Product.ID
			},
			func(price stripe.Price) stripe.Price {
				return price
			},
		).
		ForEachT(func(group linq.Group) {
			stripeProductId := group.Key.(string)
			pricesForProduct := map[string]stripe.Price{}
			for _, item := range group.Group {
				price := item.(stripe.Price)
				pricesForProduct[price.ID] = price
			}

			stripeProduct, ok := stripeProductsById[stripeProductId]
			if !ok {
				log.Warnf("stripe product with Id (%s) is missing, cannot update prices for product", stripeProductId)
				return
			}

			existingProduct, ok := existingProductsByStripeId[stripeProductId]
			if !ok {
				// There is not an existing product for these stripePrices.
				newProduct := models.Product{
					Name:            stripeProduct.Name,
					Description:     stripeProduct.Description,
					StripeProductId: stripeProductId,
					Features: []feature.Feature{
						feature.FeatureManualBudgeting,
					},
					FreeTrialDays: nil,
					Prices:        make([]models.Price, 0, len(pricesForProduct)),
				}

				for _, stripePrice := range pricesForProduct {
					if stripePrice.Recurring == nil {
						log.Warnf("stripe price (%s) on product (%s) is not recurring", stripePrice.ID, stripeProductId)
						continue
					}

					newProduct.Prices = append(newProduct.Prices, models.Price{
						Interval:        stripePrice.Recurring.Interval,
						IntervalCount:   int16(stripePrice.Recurring.IntervalCount),
						UnitAmount:      stripePrice.UnitAmount,
						StripePricingId: stripePrice.ID,
					})
				}

				productsToCreate = append(productsToCreate, newProduct)
				return
			}

			var shouldUpdateProduct bool
			if existingProduct.Name != stripeProduct.Name {
				shouldUpdateProduct = true
			}

			if existingProduct.Description != stripeProduct.Description {
				shouldUpdateProduct = true
			}

			if shouldUpdateProduct {
				existingProduct.Name = stripeProduct.Name
				existingProduct.Description = stripeProduct.Description
				productsToUpdate = append(productsToUpdate, existingProduct)
			}

			for _, stripePrice := range pricesForProduct {
				if stripePrice.Recurring == nil {
					log.Warnf("stripe price (%s) on product (%s) is not recurring", stripePrice.ID, stripeProductId)
					continue
				}

				var existingPrice *models.Price
				for _, price := range existingProduct.Prices {
					if price.StripePricingId == stripePrice.ID {
						existingPrice = &price
						break
					}
				}

				if existingPrice == nil {
					pricesToCreate = append(pricesToCreate, models.Price{
						ProductId:       existingProduct.ProductId,
						Interval:        stripePrice.Recurring.Interval,
						IntervalCount:   int16(stripePrice.Recurring.IntervalCount),
						UnitAmount:      stripePrice.UnitAmount,
						StripePricingId: stripePrice.ID,
					})
					continue
				}

				if existingPrice.UnitAmount != stripePrice.UnitAmount {
					existingPrice.UnitAmount = stripePrice.UnitAmount
					pricesToUpdate = append(pricesToUpdate, *existingPrice)
				}
			}
		})

	if numberOfProductsToCreate := len(productsToCreate); numberOfProductsToCreate > 0 {
		log.Infof("creating %d product(s) from stripe", numberOfProductsToCreate)
	}

	panic("implement me")
}

func (s *stripeBootstrapper) Close() error {
	panic("implement me")
}
