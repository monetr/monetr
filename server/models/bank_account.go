package models

import (
	"context"
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

	CreditCardBankAccountSubType BankAccountSubType = "credit card"

	AutoBankAccountSubType BankAccountSubType = "auto"
	// I'll add other bank account sub types later. Right now I'm really only
	// working with depository anyway.

	OtherBankAccountSubType BankAccountSubType = "other"
)

type BankAccountStatus string

const (
	BankAccountStatusUnknown  BankAccountStatus = "unknown"
	BankAccountStatusActive   BankAccountStatus = "active"
	BankAccountStatusInactive BankAccountStatus = "inactive"
)

var (
	_ bun.BeforeAppendModelHook = (*BankAccount)(nil)
	_ Identifiable              = BankAccount{}
)

type BankAccount struct {
	bun.BaseModel `bun:"table:bank_accounts,alias:bank_account"`

	BankAccountId          ID[BankAccount]           `json:"bankAccountId" bun:"bank_account_id,notnull,pk"`
	AccountId              ID[Account]               `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account                  `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	LinkId                 ID[Link]                  `json:"linkId" bun:"link_id,notnull,nullzero"`
	Link                   *Link                     `json:"-,omitempty" bun:"rel:belongs-to,join:link_id=link_id,join:account_id=account_id"`
	PlaidBankAccountId     *ID[PlaidBankAccount]     `json:"-" bun:"plaid_bank_account_id"`
	PlaidBankAccount       *PlaidBankAccount         `json:"plaidBankAccount,omitempty" bun:"rel:belongs-to,join:plaid_bank_account_id=plaid_bank_account_id,join:account_id=account_id"`
	LunchFlowBankAccountId *ID[LunchFlowBankAccount] `json:"lunchFlowBankAccountId" bun:"lunch_flow_bank_account_id"`
	LunchFlowBankAccount   *LunchFlowBankAccount     `json:"lunchFlowBankAccount,omitempty" bun:"rel:belongs-to,join:lunch_flow_bank_account_id=lunch_flow_bank_account_id,join:account_id=account_id"`
	Currency               string                    `json:"currency" bun:"currency,notnull,nullzero"`
	AvailableBalance       int64                     `json:"availableBalance" bun:"available_balance,notnull"`
	CurrentBalance         int64                     `json:"currentBalance" bun:"current_balance,notnull"`
	LimitBalance           int64                     `json:"limitBalance" bun:"limit_balance,notnull"`
	Mask                   *string                   `json:"mask" bun:"mask"`
	Name                   string                    `json:"name,omitempty" bun:"name,notnull,nullzero"`
	OriginalName           string                    `json:"originalName" bun:"original_name,notnull,nullzero"`
	AccountType            BankAccountType           `json:"accountType" bun:"account_type,nullzero"`
	AccountSubType         BankAccountSubType        `json:"accountSubType" bun:"account_sub_type,nullzero"`
	Status                 BankAccountStatus         `json:"status" bun:"status,notnull,nullzero"`
	LastUpdated            time.Time                 `json:"lastUpdated" bun:"last_updated,notnull,nullzero"`
	CreatedAt              time.Time                 `json:"createdAt" bun:"created_at,notnull,nullzero"`
	UpdatedAt              time.Time                 `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	DeletedAt              *time.Time                `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (BankAccount) IdentityPrefix() string {
	return "bac"
}

func (o *BankAccount) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.BankAccountId.IsZero() {
			o.BankAccountId = NewID[BankAccount]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}

		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = now
		}
	}

	return nil
}
