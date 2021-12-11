package models

import (
	"time"

	"github.com/uptrace/bun"
)

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
	bun.BaseModel `bun:"bank_accounts"`

	BankAccountId    uint64   `json:"bankAccountId" bun:"bank_account_id,notnull,pk,type:'bigserial'" example:"1234"`
	AccountId        uint64   `json:"-" bun:"account_id,notnull,pk,on_delete:CASCADE"`
	Account          *Account `json:"-" bun:"rel:has-one,join:account_id=account_id"`
	LinkId           uint64   `json:"linkId" bun:"link_id,notnull,on_delete:CASCADE" example:"2345"`
	Link             *Link    `json:"-,omitempty" bun:"rel:has-one,join:account_id=account_id,join:link_id=link_id"`
	PlaidAccountId   string   `json:"-" bun:"plaid_account_id"`
	AvailableBalance int64    `json:"availableBalance" bun:"available_balance,notnull,use_zero" example:"102356"`
	// Current Balance is a 64-bit representation of a bank account's total balance (excluding pending transactions) in
	// the form of an integer. To derive a decimal value divide this value by 100.
	CurrentBalance    int64              `json:"currentBalance" bun:"current_balance,notnull,use_zero" example:"102400"`
	Mask              string             `json:"mask" bun:"mask" example:"0000"`
	Name              string             `json:"name,omitempty" bun:"name,notnull" example:"Checking Account"`
	PlaidName         string             `json:"originalName" bun:"plaid_name" example:"Checking Account #1"`
	PlaidOfficialName string             `json:"officialName" bun:"plaid_official_name" example:"US Bank - Checking Account"`
	Type              BankAccountType    `json:"accountType" bun:"account_type" example:"depository"`
	SubType           BankAccountSubType `json:"accountSubType" bun:"account_sub_type" example:"checking"`
	LastUpdated       time.Time          `json:"lastUpdated" bun:"last_updated,notnull"`
}
