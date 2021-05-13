package models

import "github.com/monetrapp/rest-api/pkg/feature"

type Product struct {
	tableName string `pg:"products"`

	ProductId       uint64            `json:"productId" pg:"product_id,pk,type:'bigserial'"`
	Name            string            `json:"name" pg:"name,notnull"`
	Description     string            `json:"description" pg:"description"`
	StripeProductId string            `json:"-" pg:"stripe_product_id,notnull"`
	Features        []feature.Feature `json:"features" pg:"features,notnull,type:'text[]'"`
	FreeTrialDays   *uint32           `json:"freeTrialDays" pg:"free_trial_days"`

	Prices []Price `json:"prices" pg:"rel:has-many"`
}
