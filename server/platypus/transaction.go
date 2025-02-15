package platypus

import (
	"math"
	"time"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v30/plaid"
)

type Transaction interface {
	GetAmount() int64
	GetBankAccountId() string
	GetCategory() []string
	GetCategoryDetail() *string
	GetDate() time.Time
	GetDateLocal(timezone *time.Location) time.Time
	GetISOCurrencyCode() string
	GetIsPending() bool
	GetMerchantName() string
	GetName() string
	GetOriginalDescription() string
	GetPendingTransactionId() *string
	GetTransactionId() string
	GetUnofficialCurrencyCode() string
}

var (
	_ Transaction = PlaidTransaction{}
)

type PlaidTransaction struct {
	Amount                 int64
	BankAccountId          string
	Category               []string
	CategoryDetail         string
	Date                   time.Time
	ISOCurrencyCode        string
	UnofficialCurrencyCode string
	IsPending              bool
	MerchantName           string
	Name                   string
	OriginalDescription    string
	PendingTransactionId   *string
	TransactionId          string
}

func NewTransactionFromPlaid(input plaid.Transaction) (Transaction, error) {
	date, err := time.Parse("2006-01-02", input.GetDate())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse transaction date")
	}
	pendingTransactionId, _ := input.GetPendingTransactionIdOk()

	// Get the number of fractional digits for the currency of this transaction.
	fractions, err := locale.GetCurrencyInternationalFractionalDigits(
		input.GetIsoCurrencyCode(),
	)
	if err != nil {
		fractions = 2
	}

	multiplier := math.Pow(10, float64(fractions))
	transaction := PlaidTransaction{
		Amount:                 int64(input.GetAmount() * multiplier),
		BankAccountId:          input.GetAccountId(),
		Category:               input.GetCategory(),
		CategoryDetail:         input.GetPersonalFinanceCategory().Detailed,
		Date:                   date,
		ISOCurrencyCode:        input.GetIsoCurrencyCode(),
		UnofficialCurrencyCode: input.GetUnofficialCurrencyCode(),
		IsPending:              input.GetPending(),
		MerchantName:           input.GetMerchantName(),
		Name:                   input.GetName(),
		OriginalDescription:    input.GetOriginalDescription(),
		PendingTransactionId:   pendingTransactionId,
		TransactionId:          input.GetTransactionId(),
	}

	return transaction, nil
}

func (p PlaidTransaction) GetAmount() int64 {
	return p.Amount
}

func (p PlaidTransaction) GetBankAccountId() string {
	return p.BankAccountId
}

func (p PlaidTransaction) GetCategory() []string {
	return p.Category
}

func (p PlaidTransaction) GetCategoryDetail() *string {
	if p.CategoryDetail == "" {
		return nil
	}
	return &p.CategoryDetail
}

func (p PlaidTransaction) GetDate() time.Time {
	return p.Date
}

func (p PlaidTransaction) GetDateLocal(timezone *time.Location) time.Time {
	return util.InLocal(p.Date, timezone)
}

func (p PlaidTransaction) GetISOCurrencyCode() string {
	return p.ISOCurrencyCode
}

func (p PlaidTransaction) GetIsPending() bool {
	return p.IsPending
}

func (p PlaidTransaction) GetMerchantName() string {
	return p.MerchantName
}

func (p PlaidTransaction) GetName() string {
	return p.Name
}

func (p PlaidTransaction) GetOriginalDescription() string {
	return p.OriginalDescription
}

func (p PlaidTransaction) GetPendingTransactionId() *string {
	return p.PendingTransactionId
}

func (p PlaidTransaction) GetTransactionId() string {
	return p.TransactionId
}

func (p PlaidTransaction) GetUnofficialCurrencyCode() string {
	return p.UnofficialCurrencyCode
}
