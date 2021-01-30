package models

type ProductPrice struct {
	tableName string `pg:"product_prices"`

	ProductPriceId uint64 `json:"productPriceId" pg:"product_price_id,notnull,pk,type:'bigserial'"`
	StripePriceId  string `json:"-" pg:"stripe_price_id,notnull"`
}
