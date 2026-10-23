package platypus

import (
	"time"

	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/util"
	plaid "github.com/plaid/plaid-go/v49"
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

func NewTransactionFromPlaid(input *plaid.Transaction) (Transaction, error) {
	amount, _ := currency.ParseFloatToAmount(
		input.GetAmount(),
		deref(input.GetIsoCurrencyCode()),
	)
	transaction := PlaidTransaction{
		Amount:                 amount,
		BankAccountId:          input.GetAccountID(),
		Category:               input.GetCategory(),
		CategoryDetail:         input.GetPersonalFinanceCategory().GetDetailed(),
		Date:                   input.GetDate(),
		ISOCurrencyCode:        deref(input.GetIsoCurrencyCode()),
		UnofficialCurrencyCode: deref(input.GetUnofficialCurrencyCode()),
		IsPending:              input.GetPending(),
		MerchantName:           deref(input.GetMerchantName()),
		Name:                   input.GetName(),
		OriginalDescription:    deref(input.GetOriginalDescription()),
		PendingTransactionId:   input.GetPendingTransactionID(),
		TransactionId:          input.GetTransactionID(),
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
