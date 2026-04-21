package schema

import (
	"regexp"

	"github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/models"
)

var (
	CreateBankAccount = zog.Struct(zog.Shape{
		"linkId":                 ID[models.Link]().Required(),
		"lunchFlowBankAccountId": zog.Ptr(ID[models.LunchFlowBankAccount]().Optional()),
		"currency":               Currency().Default(consts.DefaultCurrencyCode).Required(),
		"availableBalance":       zog.Int64().Default(0).Required(),
		"currentBalance":         zog.Int64().Default(0).Required(),
		"limitBalance":           zog.Int64().Default(0).Required(),
		"mask":                   zog.Ptr(zog.String().Len(4).Match(regexp.MustCompile(`\d{4}`)).Optional()),
		"name":                   Name().Required(),
		"originalName":           Name().Optional(),
		"accountType": BankAccountType().
			Default(models.DepositoryBankAccountType).
			Required(),
		"accountSubType": BankAccountSubType().
			Default(models.CheckingBankAccountSubType).
			Required(),
		"status": BankAccountStatus().
			Default(models.BankAccountStatusActive).
			Required(),
	})

	PatchBankAccount = zog.Struct(zog.Shape{
		"name": Name().Optional(),
	})

	PatchLunchFlowBankAccount = zog.Struct(zog.Shape{
		"currency":       Currency().Optional(),
		"name":           Name().Optional(),
		"accountType":    BankAccountType().Optional(),
		"accountSubType": BankAccountSubType().Optional(),
	})

	PatchManualBankAccount = zog.Struct(zog.Shape{
		"currency":         Currency().Optional(),
		"availableBalance": zog.Int64().Optional(),
		"currentBalance":   zog.Int64().Optional(),
		"limitBalance":     zog.Int64().Optional(),
		"mask":             zog.Ptr(zog.String().Len(4).Match(regexp.MustCompile(`\d{4}`)).Optional()),
		"name":             Name().Optional(),
		"accountType":      BankAccountType().Optional(),
		"accountSubType":   BankAccountSubType().Optional(),
		"status":           BankAccountStatus().Optional(),
	})
)

func BankAccountType() *zog.StringSchema[models.BankAccountType] {
	return zog.StringLike[models.BankAccountType]().
		OneOf([]models.BankAccountType{
			models.DepositoryBankAccountType,
			models.CreditBankAccountType,
			models.LoanBankAccountType,
			models.InvestmentBankAccountType,
			models.OtherBankAccountType,
		})
}

func BankAccountSubType() *zog.StringSchema[models.BankAccountSubType] {
	return zog.StringLike[models.BankAccountSubType]().
		OneOf([]models.BankAccountSubType{
			models.CheckingBankAccountSubType,
			models.SavingsBankAccountSubType,
			models.HSABankAccountSubType,
			models.CDBankAccountSubType,
			models.MoneyMarketBankAccountSubType,
			models.PayPalBankAccountSubType,
			models.PrepaidBankAccountSubType,
			models.CashManagementBankAccountSubType,
			models.EBTBankAccountSubType,
			models.CreditCardBankAccountSubType,
			models.AutoBankAccountSubType,
			models.OtherBankAccountSubType,
		})
}

func BankAccountStatus() *zog.StringSchema[models.BankAccountStatus] {
	return zog.StringLike[models.BankAccountStatus]().
		OneOf([]models.BankAccountStatus{
			models.BankAccountStatusActive,
			models.BankAccountStatusInactive,
		})
}

// WithDefaultCurrency is meant to be used with a merge, it returns a struct
// schema with a required currency field, but one that defaults to the specified
// currency if none is provided.
func WithDefaultCurrency(defaultCurrency string) *zog.StructSchema {
	return zog.Struct(zog.Shape{
		"currency": Currency().Default(defaultCurrency).Required(),
	})
}
