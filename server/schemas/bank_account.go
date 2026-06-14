package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

var (
	CreateBankAccount = validation.Map(
		Name(Require),
		validation.Key(
			"mask",
			validation.OneOf(
				validation.Nil,
				Mask(),
			),
		).Required(Optional),
		validation.Key(
			"originalName",
			validation.IsString,
			is.PrintableUnicode,
			validation.Length(1, 300).Error("Original name must be between 1 and 300 characters"),
		).Required(validators.Optional),
		validation.Key(
			"linkId",
			validation.Required.Error("Link ID must be provided"),
			ValidID[models.Link](),
		).Required(Require),
		validation.Key(
			"lunchFlowBankAccountId",
			validation.OneOf(
				validation.Nil,
				ValidID[models.LunchFlowBankAccount](),
			),
		).Required(validators.Optional),
		validation.Key(
			"currency",
			// This one doesn't handle nil because IF the field is specified then it
			// needs to be valid.
			CurrencyCode(),
		).Required(Optional), // Optional because we default to USD.

		// Balances
		validation.Key(
			"limitBalance",
			// Balance is optional, so we should also handle if the field is provided
			// with a strict nil value.
			validation.OneOf(
				validation.Nil,
				validation.AllOf(
					validation.IsInteger,
					// Limit balance cannot be negative!
					validation.Min(float64(0)).Error("Limit balance cannot be negative"),
				),
			),
		).Required(Optional),
		// Current and available only enforce is integer because they allow both
		// negative and positive values.
		validation.Key(
			"currentBalance",
			validation.IsInteger,
		).Required(Optional),
		validation.Key(
			"availableBalance",
			validation.IsInteger,
		).Required(Optional),

		// Enums
		validation.Key(
			"status",
			validation.In(
				string(models.BankAccountStatusActive),
				string(models.BankAccountStatusInactive),
				string(models.BankAccountStatusUnknown),
			).Error("Invalid bank account status"),
		).Required(validators.Optional),
		validation.Key(
			"accountType",
			validation.In(
				string(models.DepositoryBankAccountType),
				string(models.CreditBankAccountType),
				string(models.LoanBankAccountType),
				string(models.InvestmentBankAccountType),
				string(models.OtherBankAccountType),
			).Error("Invalid bank account type"),
		).Required(validators.Optional),
		validation.Key(
			"accountSubType",
			validation.In(
				string(models.CheckingBankAccountSubType),
				string(models.SavingsBankAccountSubType),
				string(models.HSABankAccountSubType),
				string(models.CDBankAccountSubType),
				string(models.MoneyMarketBankAccountSubType),
				string(models.PayPalBankAccountSubType),
				string(models.PrepaidBankAccountSubType),
				string(models.CashManagementBankAccountSubType),
				string(models.EBTBankAccountSubType),
				string(models.CreditCardBankAccountSubType),
				string(models.AutoBankAccountSubType),
				string(models.OtherBankAccountSubType),
			).Error("Invalid bank account sub type"),
		).Required(validators.Optional),
	)
)
