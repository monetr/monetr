package models

import "time"

type BankAccountType string

const (
	DepositoryBankAccountType BankAccountType = "depository"
	CreditBankAccountType     BankAccountType = "credit"
	LoanBankAccountType       BankAccountType = "loan"
	InvestmentBankAccountType BankAccountType = "investment"
	OtherBankAccountType      BankAccountType = "other"
)

type BankAccountSubType string

const (
	CheckingBankAccountSubType       BankAccountSubType = "checking"
	SavingsBankAccountSubType        BankAccountSubType = "savings"
	HSABankAccountSubType            BankAccountSubType = "hsa"
	CDBankAccountSubType             BankAccountSubType = "cd"
	MoneyMarketBankAccountSubType    BankAccountSubType = "money market"
	PayPalBankAccountSubType         BankAccountSubType = "paypal"
	PrepaidBankAccountSubType        BankAccountSubType = "prepaid"
	CashManagementBankAccountSubType BankAccountSubType = "cash management"
	EBTBankAccountSubType            BankAccountSubType = "ebt"

	CreditCartBankAccountSubType BankAccountSubType = "credit card"

	AutoBankAccountSubType BankAccountSubType = "auto"
	// I'll add other bank account sub types later. Right now I'm really only working with depository anyway.
)

type BankAccount struct {
	tableName string `pg:"bank_accounts"`

	BankAccountId    uint64   `json:"bankAccountId" pg:"bank_account_id,notnull,pk,type:'bigserial'" example:"1234"`
	AccountId        uint64   `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE"`
	Account          *Account `json:"-" pg:"rel:has-one"`
	LinkId           uint64   `json:"linkId" pg:"link_id,notnull,on_delete:CASCADE" example:"2345"`
	Link             *Link    `json:"-,omitempty" pg:"rel:has-one"`
	PlaidAccountId   string   `json:"-" pg:"plaid_account_id"`
	AvailableBalance int64    `json:"availableBalance" pg:"available_balance,notnull,use_zero" example:"102356"`
	// Current Balance is a 64-bit representation of a bank account's total balance (excluding pending transactions) in
	// the form of an integer. To derive a decimal value divide this value by 100.
	CurrentBalance    int64              `json:"currentBalance" pg:"current_balance,notnull,use_zero" example:"102400"`
	Mask              string             `json:"mask" pg:"mask" example:"0000"`
	Name              string             `json:"name,omitempty" pg:"name,notnull" example:"Checking Account"`
	PlaidName         string             `json:"originalName" pg:"plaid_name" example:"Checking Account #1"`
	PlaidOfficialName string             `json:"officialName" pg:"plaid_official_name" example:"US Bank - Checking Account"`
	Type              BankAccountType    `json:"accountType" pg:"account_type" example:"depository"`
	SubType           BankAccountSubType `json:"accountSubType" pg:"account_sub_type" example:"checking"`
	LastUpdated       time.Time          `json:"lastUpdated" pg:"last_updated,notnull"`
}
