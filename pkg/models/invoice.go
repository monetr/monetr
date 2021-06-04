package models

type Invoice struct {
	tableName string `pg:"invoices"`

	InvoiceId      uint64 `json:"invoiceId"`
	AccountId      uint64 `json:"-"`
	SubscriptionId uint64 `json:"subscriptionId"`
}
