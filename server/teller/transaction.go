package teller

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	descriptionRegex = regexp.MustCompile(`\S+`)
)

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
	Date           string             `json:"date"`
	Description    string             `json:"description"`
	Details        TransactionDetails `json:"details"`
	Status         TransactionStatus  `json:"status"`
	Links          map[string]string  `json:"links"`
	Amount         string             `json:"amount"`
	RunningBalance *string            `json:"running_balance"`
	Type           string             `json:"type"`
}

func (t Transaction) GetDescription() string {
	pieces := descriptionRegex.FindAllString(t.Description, -1)
	return strings.Join(pieces, " ")
}

func (t Transaction) GetAmount() (int64, error) {
	amount, err := strconv.ParseFloat(t.Amount, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse amount")
	}

	// Convert to total cents and invert the value
	return int64(amount * -100), nil
}

func (t Transaction) GetRunningBalance() (*int64, error) {
	if t.RunningBalance == nil {
		return nil, nil
	}
	balance, err := strconv.ParseFloat(*t.RunningBalance, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse running balance")
	}

	// Convert to total cents and invert the value
	cents := int64(balance * 100)
	return &cents, nil
}

func (t Transaction) GetDate(timezone *time.Location) (time.Time, error) {
	date, err := time.ParseInLocation("2006-01-02", t.Date, timezone)
	return date, errors.Wrap(err, "failed to parse transaction date")
}

type TransactionDetails struct {
	ProcessingStatus TransactionProcessingStatus `json:"processing_status"`
	Category         TransactionCategory         `json:"category"`
	Counterparty     struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"counterparty"`
}
