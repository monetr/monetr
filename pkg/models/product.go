package models

type Product struct {
	tableName string `pg:"products"`

	ProductId       uint64   `json:"productId" pg:"product_id,pk,type:'bigserial'"`
	ProductCode     string   `json:"productCode" pg:"product_code,unique,notnull"`
	Name            string   `json:"name" pg:"name,notnull"`
	Description     string   `json:"description" pg:"description"`
	StripeProductId string   `json:"-" pg:"stripe_product_id,notnull"`
	Features        Features `json:"features" pg:"features,notnull"`

	Prices []Price `json:"prices" pg:"rel:has-many"`
}
