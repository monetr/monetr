package models

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
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

func ParseBankAccountType[T string | BankAccountType](input T) BankAccountType {
	value := BankAccountType(input)
	switch value {
	case DepositoryBankAccountType,
		CreditBankAccountType,
		LoanBankAccountType,
		InvestmentBankAccountType:
		return value
	default:
		return OtherBankAccountType
	}
}

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

func ParseBankAccountSubType[T string | BankAccountSubType](input T) BankAccountSubType {
	value := BankAccountSubType(input)
	switch value {
	case CheckingBankAccountSubType,
		SavingsBankAccountSubType,
		HSABankAccountSubType,
		CDBankAccountSubType,
		MoneyMarketBankAccountSubType,
		PayPalBankAccountSubType,
		PrepaidBankAccountSubType,
		CashManagementBankAccountSubType,
		EBTBankAccountSubType,
		CreditCardBankAccountSubType,
		AutoBankAccountSubType:
		return value
	default:
		return OtherBankAccountSubType
	}
}

type BankAccountStatus string

const (
	BankAccountStatusUnknown  BankAccountStatus = "unknown"
	BankAccountStatusActive   BankAccountStatus = "active"
	BankAccountStatusInactive BankAccountStatus = "inactive"
)

// ParseBankAccountStatus takes a string or a BankAccountStatus value and
// validates it against the known values. If the status is valid then it is
// returned as a BankAccountStatus. If it is invalid it is returned as
// UnknownBankAccountStatus.
func ParseBankAccountStatus[T string | BankAccountStatus](input T) BankAccountStatus {
	value := BankAccountStatus(input)
	switch value {
	case BankAccountStatusActive, BankAccountStatusInactive:
		return value
	default:
		return BankAccountStatusUnknown
	}
}

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

// UpdateValidator returns an array of validation rules that can be used to
// validate requests to PATCH endpoints.
func (o *BankAccount) UpdateValidator() []*validation.KeyRules[string] {
	if o.PlaidBankAccountId != nil {
		return []*validation.KeyRules[string]{
			validators.Name(false),
		}
	}

	return []*validation.KeyRules[string]{
		validators.Mask(),
		validators.Name(validators.Optional),
		validators.CurrencyCode(validators.Optional),
		validators.LimitBalance("limitBalance"),
		validators.Balance("currentBalance"),
		validators.Balance("availableBalance"),
		validation.Key(
			"status",
			validation.In(
				string(BankAccountStatusActive),
				string(BankAccountStatusInactive),
				string(BankAccountStatusUnknown),
			).Error("Invalid bank account status"),
		).Required(validators.Optional),
		validation.Key(
			"accountType",
			validation.In(
				string(DepositoryBankAccountType),
				string(CreditBankAccountType),
				string(LoanBankAccountType),
				string(InvestmentBankAccountType),
				string(OtherBankAccountType),
			).Error("Invalid bank account type"),
		).Required(validators.Optional),
		validation.Key(
			"accountSubType",
			validation.In(
				string(CheckingBankAccountSubType),
				string(SavingsBankAccountSubType),
				string(HSABankAccountSubType),
				string(CDBankAccountSubType),
				string(MoneyMarketBankAccountSubType),
				string(PayPalBankAccountSubType),
				string(PrepaidBankAccountSubType),
				string(CashManagementBankAccountSubType),
				string(EBTBankAccountSubType),
				string(CreditCardBankAccountSubType),
				string(AutoBankAccountSubType),
				string(OtherBankAccountSubType),
			).Error("Invalid bank account sub type"),
		).Required(validators.Optional),
	}
}

// UnmarshalRequest consumes a request body and an array of validation rules in
// order to create an object that can be persisted to the database. For updates,
// this function should be called on the existing object that is already stored
// in the database. The provided validators should prevent key or sensitive
// fields from being overwritten by the client's request body. For creates, the
// initial object can be left blank; or default values can be specified ahead of
// calling this function in case some fields are omitted in the intial request.
func (o *BankAccount) UnmarshalRequest(
	ctx context.Context,
	reader io.Reader,
	validators ...*validation.KeyRules[string],
) error {
	rawData := map[string]any{}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&rawData); err != nil {
		return errors.WithStack(err)
	}

	if err := validation.ValidateWithContext(
		ctx,
		&rawData,
		validation.Map(
			validators...,
		),
	); err != nil {
		return err
	}

	if err := merge.Merge(
		o, rawData, merge.ErrorOnUnknownField,
	); err != nil {
		return errors.Wrap(err, "failed to merge patched data")
	}

	return nil
}
