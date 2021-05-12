package models

type SubscriptionItem struct {
	tableName string `pg:"subscription_items"`

	SubscriptionItemId       uint64        `json:"subscriptionItemId" pg:"subscription_item_id,pk,notnull,type:'bigserial'"`
	SubscriptionId           uint64        `json:"subscriptionId" pg:"subscription_id,notnull,on_delete:CASCADE"`
	Subscription             *Subscription `json:"-" pg:"rel:has-one"`
	AccountId                uint64        `json:"-" pg:"account_id,notnull,on_delete:CASCADE"`
	Account                  *Account      `json:"-" pg:"rel:has-one"`
	StripeSubscriptionItemId string        `json:"-" pg:"stripe_subscription_item_id,notnull,unique"`
	PriceId                  uint64        `json:"priceId" pg:"price_id,notnull,on_delete:RESTRICT"`
	Price                    *Price        `json:"price,omitempty" pg:"rel:has-one"`
	Quantity                 uint8         `json:"quantity" pg:"quantity,notnull"`
}
