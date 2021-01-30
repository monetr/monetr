package models

type ProductPrice struct {
	tableName string `sql:"product_prices"`

	ProductPriceId uint64 `json:"productPriceId" sql:"product_price_id,notnull,pk,type:'bigserial'"`
	StripePriceId  string `json:"-" sql:"stripe_price_id,notnull"`
}
