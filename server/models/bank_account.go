package models

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
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
	UnknownBankAccountStatus  BankAccountStatus = "unknown"
	ActiveBankAccountStatus   BankAccountStatus = "active"
	InactiveBankAccountStatus BankAccountStatus = "inactive"
)

// ParseBankAccountStatus takes a string or a BankAccountStatus value and
// validates it against the known values. If the status is valid then it is
// returned as a BankAccountStatus. If it is invalid it is returned as
// UnknownBankAccountStatus.
func ParseBankAccountStatus[T string | BankAccountStatus](input T) BankAccountStatus {
	value := BankAccountStatus(input)
	switch value {
	case ActiveBankAccountStatus, InactiveBankAccountStatus:
		return value
	default:
		return UnknownBankAccountStatus
	}
}

var (
	_ pg.BeforeInsertHook = (*BankAccount)(nil)
	_ Identifiable        = BankAccount{}
)

type BankAccount struct {
	tableName string `pg:"bank_accounts"`

	BankAccountId          ID[BankAccount]           `json:"bankAccountId" pg:"bank_account_id,notnull,pk"`
	AccountId              ID[Account]               `json:"-" pg:"account_id,notnull,pk"`
	Account                *Account                  `json:"-" pg:"rel:has-one"`
	LinkId                 ID[Link]                  `json:"linkId" pg:"link_id,notnull"`
	Link                   *Link                     `json:"-,omitempty" pg:"rel:has-one"`
	PlaidBankAccountId     *ID[PlaidBankAccount]     `json:"-" pg:"plaid_bank_account_id"`
	PlaidBankAccount       *PlaidBankAccount         `json:"plaidBankAccount,omitempty" pg:"rel:has-one"`
	LunchFlowBankAccountId *ID[LunchFlowBankAccount] `json:"lunchFlowBankAccountId" pg:"lunch_flow_bank_account_id"`
	LunchFlowBankAccount   *LunchFlowBankAccount     `json:"lunchFlowBankAccount,omitempty" pg:"rel:has-one"`
	Currency               string                    `json:"currency" pg:"currency,notnull"`
	AvailableBalance       int64                     `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance         int64                     `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	LimitBalance           int64                     `json:"limitBalance" pg:"limit_balance,notnull,use_zero"`
	Mask                   string                    `json:"mask" pg:"mask"`
	Name                   string                    `json:"name,omitempty" pg:"name,notnull"`
	OriginalName           string                    `json:"originalName" pg:"original_name,notnull"`
	AccountType            BankAccountType           `json:"accountType" pg:"account_type"`
	AccountSubType         BankAccountSubType        `json:"accountSubType" pg:"account_sub_type"`
	Status                 BankAccountStatus         `json:"status" pg:"status,notnull"`
	LastUpdated            time.Time                 `json:"lastUpdated" pg:"last_updated,notnull"`
	CreatedAt              time.Time                 `json:"createdAt" pg:"created_at,notnull"`
	UpdatedAt              time.Time                 `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt              *time.Time                `json:"deletedAt,omitempty" pg:"deleted_at"`
}

func (BankAccount) IdentityPrefix() string {
	return "bac"
}

func (o *BankAccount) BeforeInsert(ctx context.Context) (context.Context, error) {
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

	return ctx, nil
}

// CreateValidators returns an array of validators that can be passed to the
// BankAccount UnmarshalRequest function. Specifically for requests to create a
// bank account object.
func (BankAccount) CreateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		validators.Mask(),
		validators.Name(validators.Require),
		validation.Key(
			"originalName",
			validation.Length(1, 300).Error("Original name must be between 1 and 300 characters"),
		).Required(validators.Optional),
		validation.Key(
			"linkId",
			validation.Required.Error("Link ID must be provided"),
			ValidID[Link]().Error("Link ID must be valid"),
		),
		validation.Key(
			"lunchFlowBankAccountId",
			ValidID[LunchFlowBankAccount]().Error("Lunch Flow Bank Account ID must be valid if provided"),
		).Required(validators.Optional),
		validators.CurrencyCode(validators.Optional),
		validators.LimitBalance("limitBalance"),
		validators.Balance("currentBalance"),
		validators.Balance("availableBalance"),
		validation.Key(
			"status",
			validation.In(
				string(ActiveBankAccountStatus),
				string(InactiveBankAccountStatus),
				string(UnknownBankAccountStatus),
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

// UpdateValidator returns an array of validation rules that can be used to
// validate requests to PATCH endpoints.
func (o *BankAccount) UpdateValidator() []*validation.KeyRules {
	if o.PlaidBankAccountId != nil {
		return []*validation.KeyRules{
			validators.Name(false),
		}
	}

	return []*validation.KeyRules{
		validators.Mask(),
		validators.Name(validators.Optional),
		validators.CurrencyCode(validators.Optional),
		validators.LimitBalance("limitBalance"),
		validators.Balance("currentBalance"),
		validators.Balance("availableBalance"),
		validation.Key(
			"status",
			validation.In(
				string(ActiveBankAccountStatus),
				string(InactiveBankAccountStatus),
				string(UnknownBankAccountStatus),
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
	validators ...*validation.KeyRules,
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
