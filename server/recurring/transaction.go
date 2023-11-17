package recurring

import "time"

type Transaction struct {
	TransactionId        uint64    `json:"transactionId"`
	Amount               int64     `json:"amount"`
	OriginalCategories   []string  `json:"originalCategories"`
	Date                 time.Time `json:"date"`
	OriginalName         string    `json:"originalName"`
	OriginalMerchantName *string   `json:"originalMerchantName"`
}
