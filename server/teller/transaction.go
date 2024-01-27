package teller

import "time"

type TransactionProcessingStatus string

const (
	TransactionProcessingStatusPending  TransactionProcessingStatus = "pending"
	TransactionProcessingStatusComplete TransactionProcessingStatus = "complete"
)

type TransactionCategory string

const (
	TransactionCategoryAccommodation  TransactionCategory = "accommodation"
	TransactionCategoryAdvertising    TransactionCategory = "advertising"
	TransactionCategoryBar            TransactionCategory = "bar"
	TransactionCategoryCharity        TransactionCategory = "charity"
	TransactionCategoryClothing       TransactionCategory = "clothing"
	TransactionCategoryDining         TransactionCategory = "dining"
	TransactionCategoryEducation      TransactionCategory = "education"
	TransactionCategoryElectronics    TransactionCategory = "electronics"
	TransactionCategoryEntertainment  TransactionCategory = "entertainment"
	TransactionCategoryFuel           TransactionCategory = "fuel"
	TransactionCategoryGeneral        TransactionCategory = "general"
	TransactionCategoryGroceries      TransactionCategory = "groceries"
	TransactionCategoryHealth         TransactionCategory = "health"
	TransactionCategoryHome           TransactionCategory = "home"
	TransactionCategoryIncome         TransactionCategory = "income"
	TransactionCategoryInsurance      TransactionCategory = "insurance"
	TransactionCategoryInvestment     TransactionCategory = "investment"
	TransactionCategoryLoan           TransactionCategory = "loan"
	TransactionCategoryOffice         TransactionCategory = "office"
	TransactionCategoryPhone          TransactionCategory = "phone"
	TransactionCategoryService        TransactionCategory = "service"
	TransactionCategoryShopping       TransactionCategory = "shopping"
	TransactionCategorySoftware       TransactionCategory = "software"
	TransactionCategorySport          TransactionCategory = "sport"
	TransactionCategoryTax            TransactionCategory = "tax"
	TransactionCategoryTransport      TransactionCategory = "transport"
	TransactionCategoryTransportation TransactionCategory = "transportation"
	TransactionCategoryUtilities      TransactionCategory = "utilities"
)

type TransactionStatus string

const (
	TransactionStatusPosted  TransactionStatus = "posted"
	TransactionStatusPending TransactionStatus = "pending"
)

type Transaction struct {
	Id             string             `json:"id"`
	AccountId      string             `json:"account_id"`
	Date           time.Time          `json:"date"`
	Description    string             `json:"description"`
	Details        TransactionDetails `json:"details"`
	Status         TransactionStatus  `json:"status"`
	Links          map[string]string  `json:"links"`
	Amount         string             `json:"amount"`
	RunningBalance *string            `json:"running_balance"`
	Type           string             `json:"type"`
}

type TransactionDetails struct {
	ProcessingStatus TransactionProcessingStatus `json:"processing_status"`
	Category         TransactionCategory         `json:"category"`
	Counterparty     struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"counterparty"`
}
