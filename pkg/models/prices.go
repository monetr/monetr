package models

import "github.com/stripe/stripe-go/v72"

type Price struct {
	tableName string `pg:"prices"`

	PriceId         uint64                        `json:"priceId" pg:"price_id,pk,type:'bigserial'"`
	ProductId       uint64                        `json:"productId" pg:"product_id,notnull,on_delete:CASCADE"`
	Product         *Product                      `json:"-" pg:"rel:has-one"`
	Interval        stripe.PriceRecurringInterval `json:"interval" pg:"interval,notnull"`
	IntervalCount   int16                         `json:"intervalCount" pg:"interval_count,notnull"`
	UnitAmount      int64                         `json:"unitAmount" pg:"unit_amount,notnull"`
	StripePricingId string                        `json:"-" pg:"stripe_pricing_id,notnull"`
}
