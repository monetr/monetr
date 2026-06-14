package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

var (
	CreateBankAccount = validation.Map(
		validation.Key(
			"name",
			validation.Required.Error("Name is required"),
			Name(),
		).Required(Require),
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
			BankAccountStatus(),
		).Required(validators.Optional),
		validation.Key(
			"accountType",
			BankAccountAccountType(),
		).Required(validators.Optional),
		validation.Key(
			"accountSubType",
			BankAccountAccountSubType(),
		).Required(validators.Optional),
	)

	PatchManualBankAccount = validation.Map(
		// NOTE Every key here is optional so it can be left out of a patch
		// entirely. But if a client DOES send one it cannot be blanked out. The way
		// we enforce that depends on the field:
		//   - Non-nullable strings (name, currency, accountType, accountSubType)
		//     use Required, which rejects BOTH an explicit null and an empty
		//     string. Otherwise an empty string would skip the length/format rules
		//     (those rules skip empty values) and quietly wipe the field. This
		//     matches how CreateBankAccount already guards the name.
		//   - Non-nullable integers (currentBalance, availableBalance) use NotNil,
		//     which rejects an explicit null but still allows a legitimate zero.
		//   - Genuinely nullable fields (mask, limitBalance) use a one of null
		//     rule.
		validation.Key(
			"name",
			validation.Required.Error("Name is required"),
			Name(),
		).Required(Optional),
		validation.Key(
			"mask",
			validation.OneOf(
				validation.Nil,
				Mask(),
			),
		).Required(Optional),
		validation.Key(
			"currency",
			validation.Required.Error("Currency is required"),
			// This one doesn't handle nil because IF the field is specified then it
			// needs to be valid.
			CurrencyCode(),
		).Required(Optional), // Optional because we default to USD.

		// Manual bank accounts have user managed balances, nothing syncs them like
		// Plaid does, so the client is allowed to set them directly here.
		validation.Key(
			"limitBalance",
			validation.OneOf(
				validation.Nil,
				validation.AllOf(
					validation.IsInteger,
					validation.Min(float64(0)).Error("Limit balance cannot be negative"),
				),
			),
		).Required(Optional),
		validation.Key(
			"currentBalance",
			validation.NotNil,
			validation.IsInteger,
		).Required(Optional),
		validation.Key(
			"availableBalance",
			validation.NotNil,
			validation.IsInteger,
		).Required(Optional),

		// The account type and sub type can be reclassified on a manual account,
		// but the status is intentionally left out, monetr owns that.
		validation.Key(
			"accountType",
			validation.Required.Error("Account type is required"),
			BankAccountAccountType(),
		).Required(Optional),
		validation.Key(
			"accountSubType",
			validation.Required.Error("Account sub type is required"),
			BankAccountAccountSubType(),
		).Required(Optional),
	)

	PatchBankAccount = validation.Map(
		validation.Key(
			"name",
			validation.Required.Error("Name is required"),
			Name(),
		).Required(Optional),
	)
)

func BankAccountStatus() validation.Rule {
	return validation.In(
		string(models.BankAccountStatusActive),
		string(models.BankAccountStatusInactive),
		string(models.BankAccountStatusUnknown),
	).Error("Invalid bank account status")
}

func BankAccountAccountType() validation.Rule {
	return validation.In(
		string(models.DepositoryBankAccountType),
		string(models.CreditBankAccountType),
		string(models.LoanBankAccountType),
		string(models.InvestmentBankAccountType),
		string(models.OtherBankAccountType),
	).Error("Invalid bank account type")
}

func BankAccountAccountSubType() validation.Rule {
	return validation.In(
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
	).Error("Invalid bank account sub type")
}
