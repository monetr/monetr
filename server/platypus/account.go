package platypus

import (
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/plaid/plaid-go/v30/plaid"
)

type (
	BankAccount interface {
		// GetAccountId will return Plaid's unique identifier for the bank account.
		GetAccountId() string
		// GetBalances will return the bank account's balances.
		GetBalances() BankAccountBalances
		// GetMask typically returns the last 4 of the bank account number. This
		// could technically return anything that represents a small portion of the
		// bank account's identification. I don't currently know enough about this
		// to know what other values this might have.
		GetMask() string
		// GetName will return the name of the account specified by the user or the
		// financial institution itself.
		GetName() string
		// GetOfficialName will return the name of the account specified by the
		// financial institution itself.
		GetOfficialName() string
		// GetType will return the plaid type of the account. For our use case this
		// is typically "depository".
		GetType() string
		// GetSubType will return the sub-type of the account. This can be something
		// like "checking" or "savings".
		GetSubType() string
		// GetCurrencyCode will return the currency code derived from Plaid directly
		// or will fallback to monetr's global default currency code.
		GetCurrencyCode() string
	}

	BankAccountBalances interface {
		// GetAvailable returns the total amount available for the bank account in
		// cents.
		GetAvailable() int64
		// GetCurrent returns the current bank account balance in cents. This is
		// typically the total account value excluding pending transactions.
		GetCurrent() int64
		// GetLimit returns the limit of the account (this applies for credit
		// accounts) in cents.
		GetLimit() int64
		GetIsoCurrencyCode() string
		GetUnofficialCurrencyCode() string
	}
)

var (
	_ BankAccountBalances = PlaidBankAccountBalances{}
)

func NewPlaidBankAccountBalances(
	balances plaid.AccountBalance,
) (PlaidBankAccountBalances, error) {
	available, _ := currency.ParseFloatToAmount(
		balances.GetAvailable(),
		balances.GetIsoCurrencyCode(),
	)
	current, _ := currency.ParseFloatToAmount(
		balances.GetCurrent(),
		balances.GetIsoCurrencyCode(),
	)
	limit, _ := currency.ParseFloatToAmount(
		balances.GetLimit(),
		balances.GetIsoCurrencyCode(),
	)

	return PlaidBankAccountBalances{
		// We work with all amounts in cents. So we need to convert all balances to
		// cents in order to make them whole integers rather than floats.
		Available:              available,
		Current:                current,
		Limit:                  limit,
		IsoCurrencyCode:        balances.GetIsoCurrencyCode(),
		UnofficialCurrencyCode: balances.GetUnofficialCurrencyCode(),
	}, nil
}

type PlaidBankAccountBalances struct {
	Available              int64
	Current                int64
	Limit                  int64
	IsoCurrencyCode        string
	UnofficialCurrencyCode string
}

func (p PlaidBankAccountBalances) GetAvailable() int64 {
	return p.Available
}

func (p PlaidBankAccountBalances) GetCurrent() int64 {
	return p.Current
}

func (p PlaidBankAccountBalances) GetLimit() int64 {
	return p.Limit
}

func (p PlaidBankAccountBalances) GetIsoCurrencyCode() string {
	return p.IsoCurrencyCode
}

func (p PlaidBankAccountBalances) GetUnofficialCurrencyCode() string {
	return p.UnofficialCurrencyCode
}

var (
	_ BankAccount = PlaidBankAccount{}
)

func NewPlaidBankAccount(
	bankAccount plaid.AccountBase,
) (PlaidBankAccount, error) {
	balances, err := NewPlaidBankAccountBalances(bankAccount.GetBalances())
	if err != nil {
		return PlaidBankAccount{}, err
	}

	return PlaidBankAccount{
		AccountId:    bankAccount.GetAccountId(),
		Balances:     balances,
		Mask:         bankAccount.GetMask(),
		Name:         bankAccount.GetName(),
		OfficialName: bankAccount.GetOfficialName(),
		Type:         string(bankAccount.GetType()),
		SubType:      string(bankAccount.GetSubtype()),
		// Set the currency code of the bank account to be the first non-empty
		// currency code we get from Plaid, or fall back to USD as a default.
		Currency: myownsanity.CoalesceStrings(
			balances.GetIsoCurrencyCode(),
			consts.DefaultCurrencyCode,
		),
	}, nil
}

type PlaidBankAccount struct {
	AccountId    string
	Balances     PlaidBankAccountBalances
	Mask         string
	Name         string
	OfficialName string
	Type         string
	SubType      string
	Currency     string
}

func (p PlaidBankAccount) GetAccountId() string {
	return p.AccountId
}

func (p PlaidBankAccount) GetBalances() BankAccountBalances {
	return p.Balances
}

func (p PlaidBankAccount) GetMask() string {
	return p.Mask
}

func (p PlaidBankAccount) GetName() string {
	return p.Name
}

func (p PlaidBankAccount) GetOfficialName() string {
	return p.OfficialName
}

func (p PlaidBankAccount) GetType() string {
	return p.Type
}

func (p PlaidBankAccount) GetSubType() string {
	return p.SubType
}

func (p PlaidBankAccount) GetCurrencyCode() string {
	return p.Currency
}
