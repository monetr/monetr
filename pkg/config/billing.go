package config

import "github.com/monetrapp/rest-api/pkg/feature"

type Product struct {
	Name            string
	Description     string
	ProductCode     string
	StripeProductId string
	Features        []feature.Feature
	Prices          []Price
	Visible         bool
}

type Price struct {
	PriceCode       string
	StripePriceId   string
	Visible         bool
}
