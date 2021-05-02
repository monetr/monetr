package models

type Product struct {
	tableName string `sql:"products"`

	ProductId       uint64 `json:"productId" sql:"product_id,pk,type:'bigserial'"`
	ProductCode     string `json:"productCode" sql:"product_code,unique"`
	Name            string `json:"name" sql:"name,notnull"`
	Description     string `json:"description" sql:"description"`
	StripeProductId string `json:"-" sql:"stripe_product_id,notnull"`

	Prices []Price `json:"prices" sql:"rel:has-many"`
}
