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

	CreditCardBankAccountSubType BankAccountSubType = "credit card"

	AutoBankAccountSubType BankAccountSubType = "auto"
	// I'll add other bank account sub types later. Right now I'm really only working with depository anyway.
)

type BankAccountStatus string

const (
	UnknownBankAccountStatus  BankAccountStatus = "unknown"
	ActiveBankAccountStatus   BankAccountStatus = "active"
	InactiveBankAccountStatus BankAccountStatus = "inactive"
)

type BankAccount struct {
	tableName string `pg:"bank_accounts"`

	BankAccountId      uint64             `json:"bankAccountId" pg:"bank_account_id,notnull,pk,type:'bigserial'"`
	AccountId          uint64             `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE"`
	Account            *Account           `json:"-" pg:"rel:has-one"`
	LinkId             uint64             `json:"linkId" pg:"link_id,notnull,on_delete:CASCADE"`
	Link               *Link              `json:"-,omitempty" pg:"rel:has-one"`
	PlaidBankAccountId *uint64            `json:"plaidBankAccountId" pg:"plaid_bank_account_id"`
	PlaidBankAccount   *PlaidBankAccount  `json:"-" pg:"rel:has-one,fk:plaid_bank_account_id"`
	AvailableBalance   int64              `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance     int64              `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	Mask               string             `json:"mask" pg:"mask"`
	Name               string             `json:"name,omitempty" pg:"name,notnull"`
	Type               BankAccountType    `json:"accountType" pg:"account_type"`
	SubType            BankAccountSubType `json:"accountSubType" pg:"account_sub_type"`
	Status             BankAccountStatus  `json:"status" pg:"status,notnull"`
	LastUpdated        time.Time          `json:"lastUpdated" pg:"last_updated,notnull"`
	CreatedAt          time.Time          `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId    uint64             `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser      *User              `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
