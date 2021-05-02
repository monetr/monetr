package models

type Price struct {
	tableName string `sql:"prices"`

	PriceId         uint64   `json:"priceId" sql:"price_id,pk,type:'bigserial'"`
	ProductId       uint64   `json:"productId" sql:"product_id,notnull,on_delete:cascade"`
	Product         *Product `json:"-" sql:"rel:has-one"`
	Interval        string   `json:"interval" sql:"interval,notnull"`
	IntervalCount   int16    `json:"intervalCount" sql:"interval_count,notnull"`
	TrialPeriodDays int32    `json:"trialPeriodDays" sql:"trial_period_days,notnull"`
	UnitAmount      int64    `json:"unitAmount" sql:"unit_amount,notnull"`
	StripePricingId string   `json:"-" sql:"stripe_pricing_id,notnull"`
}
