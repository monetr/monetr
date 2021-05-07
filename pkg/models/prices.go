package models

type Price struct {
	tableName string `pg:"prices"`

	PriceId         uint64   `json:"priceId" pg:"price_id,pk,type:'bigserial'"`
	PriceCode       string   `json:"priceCode" pg:"price_code,notnull,unique"`
	ProductId       uint64   `json:"productId" pg:"product_id,notnull,on_delete:cascade"`
	Product         *Product `json:"-" pg:"rel:has-one"`
	Interval        string   `json:"interval" pg:"interval,notnull"`
	IntervalCount   int16    `json:"intervalCount" pg:"interval_count,notnull"`
	TrialPeriodDays *int32   `json:"trialPeriodDays" pg:"trial_period_days"`
	UnitAmount      int64    `json:"unitAmount" pg:"unit_amount,notnull"`
	StripePricingId string   `json:"-" pg:"stripe_pricing_id,notnull"`
}
